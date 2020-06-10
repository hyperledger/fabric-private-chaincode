/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger-labs/fabric-private-chaincode/ecc/crypto"
	"github.com/hyperledger-labs/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric-protos-go/peer"
	commonerrors "github.com/hyperledger/fabric/common/errors"
	"github.com/hyperledger/fabric/common/flogging"
	validation "github.com/hyperledger/fabric/core/handlers/validation/api/state"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/rwsetutil"
	"github.com/hyperledger/fabric/protoutil"
)

var logger = flogging.MustGetLogger("ecc-vscc")

// New creates a new instance of the ercc VSCC
// Typically this will only be invoked once per peer
func New(stateFetcher validation.StateFetcher) *VSCCECC {
	return &VSCCECC{
		verifier: &crypto.ECDSAVerifier{},
		sf:       stateFetcher,
	}
}

type VSCCECC struct {
	verifier crypto.Verifier
	sf       validation.StateFetcher
}

// Validate validates the given envelope corresponding to a transaction with an endorsement
// policy as given in its serialized form
func (vscc *VSCCECC) Validate(envelopeBytes []byte, policyBytes []byte) commonerrors.TxValidationError {
	// get the envelope...
	env, err := protoutil.GetEnvelopeFromBlock(envelopeBytes)
	if err != nil {
		logger.Errorf("ECC-VSCC error: GetEnvelope failed, err %s", err)
		return policyErr(err)
	}

	// ...and the payload...
	payl, err := protoutil.UnmarshalPayload(env.Payload)
	if err != nil {
		logger.Errorf("ECC-VSCC error: GetPayload failed, err %s", err)
		return policyErr(err)
	}

	// ...and the transaction...
	tx, err := protoutil.UnmarshalTransaction(payl.Data)
	if err != nil {
		logger.Errorf("ECC-VSCC error: GetTransaction failed, err %s", err)
		return policyErr(err)
	}

	// loop through each of the actions within
	for _, act := range tx.Actions {
		// first get proposal response
		cap, err := protoutil.UnmarshalChaincodeActionPayload(act.Payload)
		if err != nil {
			logger.Errorf("ECC-VSCC error: GetChaincodeActionPayload failed, err %s", err)
			return policyErr(err)
		}

		pRespPayload, err := protoutil.UnmarshalProposalResponsePayload(cap.Action.ProposalResponsePayload)
		if err != nil {
			logger.Errorf("ECC-VSCC error: GetProposalResponsePayload failed, err %s", err)
			return policyErr(err)
		}

		ccAction := &peer.ChaincodeAction{}
		if err = proto.Unmarshal(pRespPayload.Extension, ccAction); err != nil {
			logger.Errorf("ECC-VSCC error: GetProposalResponsePayload failed, err %s", err)
			return policyErr(err)
		}

		// next get invocation specs
		cpp, err := protoutil.UnmarshalChaincodeProposalPayload(cap.ChaincodeProposalPayload)
		if err != nil {
			return policyErr(err)
		}

		cis := &peer.ChaincodeInvocationSpec{}
		if err = proto.Unmarshal(cpp.Input, cis); err != nil {
			return policyErr(err)
		}

		// finally validate proposal and response
		if err = vscc.checkEnclaveEndorsement(cis, ccAction); err != nil {
			logger.Errorf("ECC-VSCC error: checkEnclaveEndorsement failed, err %s", err)
			return policyErr(err)
		}
	}
	return nil
}

func (vscc *VSCCECC) checkEnclaveEndorsement(cis *peer.ChaincodeInvocationSpec, respPayload *peer.ChaincodeAction) error {
	logger.Debug("checkEnclaveEndorsement starts")

	channelState, err := vscc.sf.FetchState()
	if err != nil {
		return fmt.Errorf("Fetch channel state failed, err %s", err)
	}
	defer channelState.Done()
	state := &state{channelState}

	txRWSet := &rwsetutil.TxRwSet{}
	if err := txRWSet.FromProtoBytes(respPayload.Results); err != nil {
		return err
	}

	function := cis.ChaincodeSpec.Input.Args[0]
	logger.Debugf("function: %s\n", string(function))

	if string(function) == "__setup" {
		// We skip validation of `__setup` transaction as there is no enclave endorsement for this tx type.
		// In particular, `__setup` invokes `registerEnclave` at ercc via a cc2cc call. The corresponding ercc-vscc
		// is responsible to validate this transaction triggered through `__setup`
		return nil
	}

	// Note: we encode all args before sending to enclave (and this is also what is
	// ultimately signed), see enclave_chaincode.go::{init,invoke} for encoding.
	var txType []byte
	var argss []string
	if string(function) == "__init" {
		txType = []byte("init")
		argss = make([]string, len(cis.ChaincodeSpec.Input.Args[1:]))
		for i, v := range cis.ChaincodeSpec.Input.Args[1:] { // drop "__init"
			argss[i] = string(v)
		}
	} else {
		txType = []byte("invoke")
		argss = make([]string, len(cis.ChaincodeSpec.Input.Args))
		for i, v := range cis.ChaincodeSpec.Input.Args {
			argss[i] = string(v)
		}
	}
	var encoded_args []byte
	encoded_args, err = json.Marshal(argss)
	if err != nil {
		return fmt.Errorf("Couldn't json encode arguments, err %s", err)
	}
	logger.Debugf("txType: %s / encoded_args: %s\n", string(txType), string(encoded_args))

	// get the enclave response
	response := &utils.Response{}
	if err := json.Unmarshal(respPayload.Response.Payload, response); err != nil {
		return fmt.Errorf("Unmarshalling of SGX response failed, err: %s", err)
	}
	logger.Debugf("response: %s\n", string(response.ResponseData))

	// check that response.PublicKey is registred
	enclavePkHash := sha256.Sum256(response.PublicKey)
	base64PublicKey := base64.StdEncoding.EncodeToString(enclavePkHash[:])
	logger.Debugf("pk: %s", base64PublicKey)

	// FIXME: remove hardcoding of those strings
	attestation, err := state.GetState("ercc", base64PublicKey)
	if err != nil || attestation == nil {
		return fmt.Errorf("Enclave PK not found in registry")
	}
	logger.Debugf("found attestation at ercc for pk: %s", base64PublicKey)

	var readset, writeset [][]byte
	// go through the read/writeset as produced within the enclave
	for _, ns := range txRWSet.NsRwSets {
		if ns.NameSpace != cis.ChaincodeSpec.ChaincodeId.Name {
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

	isValid, err := vscc.verifier.Verify(txType, encoded_args, response.ResponseData, readset, writeset, response.Signature, response.PublicKey)
	if err != nil {
		return fmt.Errorf("Response invalid! Signature verification failed! Error: %s", err)
	}
	if !isValid {
		return fmt.Errorf("Response invalid! Signature verification failed!")
	}

	logger.Debug("Enclave signature validation successfully passed")
	return nil
}

type state struct {
	validation.State
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
