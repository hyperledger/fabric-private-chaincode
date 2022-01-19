/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package endorsement

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-private-chaincode/internal/crypto"
	"github.com/hyperledger/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/pkg/errors"
)

var logger = flogging.MustGetLogger("validate")

type Validation interface {
	ReplayReadWrites(stub shim.ChaincodeStubInterface, fpcrwset *protos.FPCKVSet) error
	Validate(signedResponseMessage *protos.SignedChaincodeResponseMessage, attestedData *protos.AttestedData) error
}

func NewValidator() *ValidatorImpl {
	return &ValidatorImpl{csp: crypto.GetDefaultCSP()}
}

type ValidatorImpl struct {
	csp crypto.CSP
}

func (v *ValidatorImpl) ReplayReadWrites(stub shim.ChaincodeStubInterface, fpcrwset *protos.FPCKVSet) (err error) {
	// nil rwset => nothing to do
	if fpcrwset == nil {
		return nil
	}

	rwset := fpcrwset.GetRwSet()
	if rwset == nil {
		return fmt.Errorf("no rwset found")
	}

	// normal reads
	if rwset.GetReads() != nil {
		logger.Debugf("Replaying reads")
		if fpcrwset.GetReadValueHashes() == nil {
			return fmt.Errorf("no read value hash associated to reads")
		}
		if len(fpcrwset.ReadValueHashes) != len(rwset.Reads) {
			return fmt.Errorf("%d read value hashes but %d reads", len(fpcrwset.ReadValueHashes), len(rwset.Reads))
		}

		for i := 0; i < len(rwset.Reads); i++ {
			k := utils.TransformToFPCKey(rwset.Reads[i].Key)

			// check if composite key, if so, derive Fabric key
			if utils.IsFPCCompositeKey(k) {
				comp := utils.SplitFPCCompositeKey(k)
				k, _ = stub.CreateCompositeKey(comp[0], comp[1:])
			}

			v, err := stub.GetState(k)
			if err != nil {
				return fmt.Errorf("error (%s) reading key %s", err, k)
			}

			logger.Debugf("read key='%s' value(hex)='%s'", k, hex.EncodeToString(v))

			// compute value hash
			// TODO: use CSP hash for consistency
			h := sha256.New()
			h.Write(v)
			valueHash := h.Sum(nil)

			// check hashes
			if !bytes.Equal(valueHash, fpcrwset.ReadValueHashes[i]) {
				logger.Debugf("value(hex): %s", hex.EncodeToString(v))
				logger.Debugf("computed hash(hex): %s", hex.EncodeToString(valueHash))
				logger.Debugf("received hash(hex): %s", hex.EncodeToString(fpcrwset.ReadValueHashes[i]))
				return fmt.Errorf("value hash mismatch for key %s", k)
			}
		}
	}

	// range query reads
	if rwset.GetRangeQueriesInfo() != nil {
		return fmt.Errorf("RangeQuery support not implemented, missing hash check")
		// TODO implement me when enabling support for range queries
		//logger.Debugf("Replaying range queries")
		//for _, rqi := range rwset.RangeQueriesInfo {
		//	if rqi.GetRawReads() == nil {
		//		// no raw reads available in this range query
		//		continue
		//	}
		//	for _, qr := range rqi.GetRawReads().KvReads {
		//		k := utils.TransformToFPCKey(qr.Key)
		//		v, err := stub.GetState(k)
		//		if err != nil {
		//			return fmt.Errorf("error (%s) reading key %s", err, k)
		//		}
		//
		//		_ = v
		//		return fmt.Errorf("TODO: not implemented, missing hash check")
		//	}
		//}
	}

	// writes
	if rwset.GetWrites() != nil {
		logger.Debugf("Replaying writes")
		for _, w := range rwset.Writes {
			k := utils.TransformToFPCKey(w.Key)

			// check if composite key, if so, derive Fabric key
			if utils.IsFPCCompositeKey(k) {
				comp := utils.SplitFPCCompositeKey(k)
				k, _ = stub.CreateCompositeKey(comp[0], comp[1:])
			}

			if w.IsDelete {
				if err := stub.DelState(k); err != nil {
					return fmt.Errorf("error (%s) deleting key %s", err, k)
				}
				logger.Debugf("key %s deleted", k)
			} else {
				if err := stub.PutState(k, w.Value); err != nil {
					return fmt.Errorf("error (%s) writing key %s value(hex) %s", err, k, hex.EncodeToString(w.Value))
				}
				logger.Debugf("written key %s value(hex) %s", k, hex.EncodeToString(w.Value))
			}
		}
	}

	return nil
}

func (v *ValidatorImpl) Validate(signedResponseMessage *protos.SignedChaincodeResponseMessage, attestedData *protos.AttestedData) error {
	if signedResponseMessage.GetSignature() == nil {
		return fmt.Errorf("no enclave signature")
	}

	if signedResponseMessage.GetChaincodeResponseMessage() == nil {
		return fmt.Errorf("no chaincode response")
	}

	if attestedData.GetEnclaveVk() == nil {
		return fmt.Errorf("no enclave verification key")
	}

	// verify enclave signature
	err := v.csp.VerifyMessage(attestedData.EnclaveVk, signedResponseMessage.ChaincodeResponseMessage, signedResponseMessage.Signature)
	if err != nil {
		return fmt.Errorf("enclave signature verification failed")
	}

	// verify signed proposal input hash matches input hash
	chaincodeResponseMessage, err := utils.UnmarshalChaincodeResponseMessage(signedResponseMessage.GetChaincodeResponseMessage())
	if err != nil {
		return errors.Wrap(err, "failed to extract response message")
	}

	originalSignedProposal := chaincodeResponseMessage.GetProposal()
	if originalSignedProposal == nil {
		return fmt.Errorf("cannot get the signed proposal that the enclave received")
	}
	chaincodeRequestMessageBytes, err := utils.GetChaincodeRequestMessageFromSignedProposal(originalSignedProposal)
	if err != nil {
		return errors.Wrap(err, "failed to extract chaincode request message")
	}
	expectedChaincodeRequestMessageHash := sha256.Sum256(chaincodeRequestMessageBytes)
	chaincodeRequestMessageHash := chaincodeResponseMessage.GetChaincodeRequestMessageHash()
	if chaincodeRequestMessageHash == nil {
		return fmt.Errorf("cannot get the chaincode request message hash")
	}
	if !bytes.Equal(expectedChaincodeRequestMessageHash[:], chaincodeRequestMessageHash) {
		logger.Debugf("expected chaincode request message hash: %s", strings.ToUpper(hex.EncodeToString(expectedChaincodeRequestMessageHash[:])))
		logger.Debugf("received chaincode request message hash: %s", strings.ToUpper(hex.EncodeToString(chaincodeRequestMessageHash[:])))
		return fmt.Errorf("chaincode request message hash mismatch")
	}

	return nil
}
