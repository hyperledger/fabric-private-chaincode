/*
* Copyright IBM Corp. 2018 All Rights Reserved.
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/golang/protobuf/proto"
	commonerrors "github.com/hyperledger/fabric/common/errors"
	"github.com/hyperledger/fabric/common/flogging"
	. "github.com/hyperledger/fabric/core/handlers/validation/api/state"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/rwsetutil"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/protos/utils"
	"github.com/hyperledger-labs/fabric-secure-chaincode/ecc/crypto"
	sgx_utils "github.com/hyperledger-labs/fabric-secure-chaincode/utils"
)

var logger = flogging.MustGetLogger("vscc")

// New creates a new instance of the ercc VSCC
// Typically this will only be invoked once per peer
func New(stateFetcher StateFetcher) *VSCCECC {
	return &VSCCECC{
		verifier: &crypto.ECDSAVerifier{},
		sf:       stateFetcher,
	}
}

type VSCCECC struct {
	verifier crypto.Verifier
	sf       StateFetcher
}

// Validate validates the given envelope corresponding to a transaction with an endorsement
// policy as given in its serialized form
func (vscc *VSCCECC) Validate(envelopeBytes []byte, policyBytes []byte) commonerrors.TxValidationError {
	// get the envelope...
	env, err := utils.GetEnvelopeFromBlock(envelopeBytes)
	if err != nil {
		logger.Errorf("ECC-VSCC error: GetEnvelope failed, err %s", err)
		return policyErr(err)
	}

	// ...and the payload...
	payl, err := utils.GetPayload(env)
	if err != nil {
		logger.Errorf("ECC-VSCC error: GetPayload failed, err %s", err)
		return policyErr(err)
	}

	// ...and the transaction...
	tx, err := utils.GetTransaction(payl.Data)
	if err != nil {
		logger.Errorf("ECC-VSCC error: GetTransaction failed, err %s", err)
		return policyErr(err)
	}

	// loop through each of the actions within
	for _, act := range tx.Actions {
		// first get proposal response
		cap, err := utils.GetChaincodeActionPayload(act.Payload)
		if err != nil {
			logger.Errorf("ECC-VSCC error: GetChaincodeActionPayload failed, err %s", err)
			return policyErr(err)
		}

		pRespPayload, err := utils.GetProposalResponsePayload(cap.Action.ProposalResponsePayload)
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
		cpp, err := utils.GetChaincodeProposalPayload(cap.ChaincodeProposalPayload)
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

	for _, ns := range txRWSet.NsRwSets {
		logger.Debugf("Namespace %s", ns.NameSpace)

		// TODO make this more flexible
		if ns.NameSpace != "ecc" {
			continue
		}

		// get the args of the ecc invocation
		// carefull we need only args[0] (function) as it includes all arguments
		args := cis.ChaincodeSpec.Input.Args[0]
		logger.Debugf("args: %s\n", string(args))

		// get the enclave response
		response := &sgx_utils.Response{}
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

		// Next, reproduce sorted read/writeset
		var readset, writeset [][]byte

		// normal reads
		var readKeys []string
		for _, r := range ns.KvRwSet.Reads {
			k := sgx_utils.TransformToSGX(r.Key, sgx_utils.SEP)
			readKeys = append(readKeys, k)
		}

		// range query reads
		for _, rqi := range ns.KvRwSet.RangeQueriesInfo {
			for _, qr := range rqi.GetRawReads().KvReads {
				k := sgx_utils.TransformToSGX(qr.Key, sgx_utils.SEP)
				readKeys = append(readKeys, k)
			}
		}

		// writes
		var writeKeys []string
		writesetMap := make(map[string][]byte)
		for _, w := range ns.KvRwSet.Writes {
			k := sgx_utils.TransformToSGX(w.Key, sgx_utils.SEP)
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

		isValid, err := vscc.verifier.Verify(args, response.ResponseData, readset, writeset, response.Signature, response.PublicKey)
		if err != nil {
			return fmt.Errorf("Response invalid! Signature verification failed! Error: %s", err)
		}
		if !isValid {
			return fmt.Errorf("Response invalid! Signature verification failed!")
		}

	}

	return nil
}

type state struct {
	State
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
