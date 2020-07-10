/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fpc

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/hyperledger-labs/fabric-private-chaincode/ecc/crypto"
	"github.com/hyperledger-labs/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric-protos-go/common"
	commonerrors "github.com/hyperledger/fabric/common/errors"
	"github.com/hyperledger/fabric/common/flogging"
	vs "github.com/hyperledger/fabric/core/handlers/validation/api/state"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/rwsetutil"
	"github.com/hyperledger/fabric/protoutil"
)

var logger = flogging.MustGetLogger("fpc-vscc")

// New creates a new instance of the ercc VSCC
// Typically this will only be invoked once per peer
func New(sf vs.StateFetcher) *FPCValidator {
	return &FPCValidator{
		verifier:     &crypto.ECDSAVerifier{},
		stateFetcher: sf,
	}
}

type FPCValidator struct {
	verifier     crypto.Verifier
	stateFetcher vs.StateFetcher
}

type validationArtifacts struct {
	chaincodeId string
	args        [][]byte
	rwset       []byte
	response    []byte
}

// Validate validates the given envelope corresponding to a transaction with an endorsement
// policy as given in its serialized form
func (vscc *FPCValidator) Validate(block *common.Block, namespace string, txPosition int, actionPosition int, policy []byte) commonerrors.TxValidationError {

	// get first all information we need to validate enclave endorsenebt
	va, err := vscc.extractValidationArtifacts(block, txPosition, actionPosition)
	if err != nil {
		return policyErr(err)
	}

	// finally validate proposal and response
	txverr := vscc.validateFPCEndorsement(va)
	if txverr != nil {
		logger.Errorf("VSCC error: validateFPCEndorsement failed, err %s", txverr)
		return txverr
	}
	return nil
}

func (vscc *FPCValidator) extractValidationArtifacts(
	block *common.Block,
	txPosition int,
	actionPosition int,
) (*validationArtifacts, error) {
	// get the envelope...
	env, err := protoutil.GetEnvelopeFromBlock(block.Data.Data[txPosition])
	if err != nil {
		logger.Errorf("VSCC error: GetEnvelope failed, err %s", err)
		return nil, err
	}

	// ...and the payload...
	payl, err := protoutil.UnmarshalPayload(env.Payload)
	if err != nil {
		logger.Errorf("VSCC error: GetPayload failed, err %s", err)
		return nil, err
	}

	chdr, err := protoutil.UnmarshalChannelHeader(payl.Header.ChannelHeader)
	if err != nil {
		return nil, err
	}

	// validate the payload type
	if common.HeaderType(chdr.Type) != common.HeaderType_ENDORSER_TRANSACTION {
		logger.Errorf("Only Endorser Transactions are supported, provided type %d", chdr.Type)
		err = fmt.Errorf("Only Endorser Transactions are supported, provided type %d", chdr.Type)
		return nil, err
	}

	// ...and the transaction...
	tx, err := protoutil.UnmarshalTransaction(payl.Data)
	if err != nil {
		logger.Errorf("VSCC error: GetTransaction failed, err %s", err)
		return nil, err
	}

	cap, err := protoutil.UnmarshalChaincodeActionPayload(tx.Actions[actionPosition].Payload)
	if err != nil {
		logger.Errorf("VSCC error: GetChaincodeActionPayload failed, err %s", err)
		return nil, err
	}

	pRespPayload, err := protoutil.UnmarshalProposalResponsePayload(cap.Action.ProposalResponsePayload)
	if err != nil {
		err = fmt.Errorf("GetProposalResponsePayload error %s", err)
		return nil, err
	}
	if pRespPayload.Extension == nil {
		err = fmt.Errorf("nil pRespPayload.Extension")
		return nil, err
	}
	respPayload, err := protoutil.UnmarshalChaincodeAction(pRespPayload.Extension)
	if err != nil {
		err = fmt.Errorf("GetChaincodeAction error %s", err)
		return nil, err
	}

	// next get invocation specs
	cpp, err := protoutil.UnmarshalChaincodeProposalPayload(cap.ChaincodeProposalPayload)
	if err != nil {
		err = fmt.Errorf("GetChaincodeProposalPayload error %s", err)
		return nil, err
	}

	cis, err := protoutil.UnmarshalChaincodeInvocationSpec(cpp.Input)
	if err != nil {
		err = fmt.Errorf("GetChaincodeInvocationSpec error %s", err)
		return nil, err
	}

	return &validationArtifacts{
		chaincodeId: cis.ChaincodeSpec.ChaincodeId.Name,
		args:        cis.ChaincodeSpec.Input.Args,
		rwset:       respPayload.Results,
		response:    respPayload.Response.Payload,
	}, nil
}

func (vscc *FPCValidator) validateFPCEndorsement(va *validationArtifacts) commonerrors.TxValidationError {
	logger.Debug("validateFPCEndorsement starts")

	functionName := string(va.args[0])
	logger.Debugf("function: %s\n", functionName)
	if functionName == "__setup" {
		// We skip vs of `__setup` transaction as there is no enclave endorsement for this tx type.
		// In particular, `__setup` invokes `registerEnclave` at ercc via a cc2cc call. The corresponding ercc-vscc
		// is responsible to validate this transaction triggered through `__setup`
		return nil
	}

	// get transaction type and arguments as
	txType, encodedArgs, err := encodeArgs(va.args)
	if err != nil {
		return policyErr(err)
	}
	logger.Debugf("txType: %s / encodedArgs: %s\n", string(txType), string(encodedArgs))

	// get the read/write set in the same format as processed by the chaincode enclaves
	readset, writeset, err := extractFPCRWSet(va.chaincodeId, va.rwset)
	if err != nil {
		return policyErr(err)
	}

	// get the enclave response
	response, err := utils.UnmarshalResponse(va.response)
	if err != nil {
		return policyErr(err)
	}
	logger.Debugf("response: %s\n", string(response.ResponseData))

	err = vscc.checkIsRegistered(response.PublicKey)
	if err != nil {
		return policyErr(err)
	}

	isValid, err := vscc.verifier.Verify(txType, encodedArgs, response.ResponseData, readset, writeset, response.Signature, response.PublicKey)
	if err != nil {
		return policyErr(fmt.Errorf("signature verification failed, err: %s", err))
	}
	if !isValid {
		return policyErr(fmt.Errorf("FPC response invalid! Signature verification failed"))
	}

	logger.Debug("FPC transaction vs successfully passed")
	return nil
}

func extractFPCRWSet(chaincodeId string, rwset []byte) (readset [][]byte, writeset [][]byte, err error) {
	txRWSet := &rwsetutil.TxRwSet{}
	if err := txRWSet.FromProtoBytes(rwset); err != nil {
		return nil, nil, err
	}

	// go through the read/writeset as produced within the enclave
	for _, ns := range txRWSet.NsRwSets {
		if ns.NameSpace != chaincodeId {
			// note that the read/write set may contain other namespaces than our fpc chaincode such as _lifecycle and ercc
			// here we only care about FPC chaincode namespace and therefore filter in order to verify enclave signature
			continue
		}

		// normal reads
		var readKeys []string
		for _, r := range ns.KvRwSet.Reads {
			k := utils.TransformToFPCKey(r.Key)
			readKeys = append(readKeys, k)
		}

		// range query reads
		for _, rqi := range ns.KvRwSet.RangeQueriesInfo {
			if rqi.GetRawReads() == nil {
				// no raw reads available in this range query
				continue
			}
			for _, qr := range rqi.GetRawReads().KvReads {
				k := utils.TransformToFPCKey(qr.Key)
				readKeys = append(readKeys, k)
			}
		}

		// writes
		var writeKeys []string
		writesetMap := make(map[string][]byte)
		for _, w := range ns.KvRwSet.Writes {
			k := utils.TransformToFPCKey(w.Key)
			writeKeys = append(writeKeys, k)
			writesetMap[k] = w.Value
		}

		// sort readset and writeset as enclave uses a sorted map
		sort.Strings(readKeys)
		sort.Strings(writeKeys)

		logger.Debug("reads:")
		for _, k := range readKeys {
			logger.Debugf("\t%s\n", k)
			readset = append(readset, []byte(k))
		}

		logger.Debug("writes:")
		for _, k := range writeKeys {
			logger.Debugf("\t%s - %s\n", k, string(writesetMap[k]))
			writeset = append(writeset, []byte(k))
			writeset = append(writeset, writesetMap[k])
		}
	}

	return readset, writeset, nil
}

func (vscc *FPCValidator) checkIsRegistered(publicKey []byte) error {
	if publicKey == nil {
		return fmt.Errorf("wrong input; publicKey is nil")
	}

	channelState, err := vscc.stateFetcher.FetchState()
	if err != nil {
		return fmt.Errorf("fetch channel state failed, err %s", err)
	}
	defer channelState.Done()
	state := &state{channelState}

	// first hash key
	enclavePkHash := sha256.Sum256(publicKey)
	// then encode using base64;
	base64PublicKey := base64.StdEncoding.EncodeToString(enclavePkHash[:])
	logger.Debugf("lookup attestation for pk: %s", base64PublicKey)

	// check that public key is registered
	// FIXME: remove hardcoded ercc string - is there a better way how to get the ercc namespace
	attestation, err := state.GetState("ercc", base64PublicKey)
	if err != nil {
		return fmt.Errorf("getState failed err: %s", err)
	}

	if attestation == nil {
		return fmt.Errorf("no attestation found for pk: %s", base64PublicKey)
	}

	logger.Debugf("attestation found for pk: %s", base64PublicKey)
	return nil
}

func encodeArgs(args [][]byte) (txType []byte, encodedArgs []byte, err error) {
	// Note: we encode all args before sending to enclave (and this is also what is
	// ultimately signed), see enclave_chaincode.go::{init,invoke} for encoding.
	functionName := string(args[0])
	var argss []string
	if functionName == "__init" {
		txType = []byte("init")
		argss = make([]string, len(args[1:]))
		for i, v := range args[1:] { // drop "__init"
			argss[i] = string(v)
		}
	} else {
		txType = []byte("invoke")
		argss = make([]string, len(args))
		for i, v := range args {
			argss[i] = string(v)
		}
	}

	encodedArgs, err = json.Marshal(argss)
	if err != nil {
		return nil, nil, fmt.Errorf("json encoding failed, err %s", err)
	}

	return txType, encodedArgs, nil
}

type state struct {
	vs.State
}

// GetState retrieves the value for the given key in the given namespace
func (s *state) GetState(namespace string, key string) ([]byte, error) {
	values, err := s.GetStateMultipleKeys(namespace, []string{key})
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, nil
	}
	return values[0], nil
}

func policyErr(err error) *commonerrors.VSCCEndorsementPolicyError {
	return &commonerrors.VSCCEndorsementPolicyError{
		Err: err,
	}
}
