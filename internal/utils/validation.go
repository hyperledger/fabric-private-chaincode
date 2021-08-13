// +build WITH_PDO_CRYPTO

/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/pkg/errors"
)

// Note that we use the WITH_PDO_CRYPTO build tag to enable this code.
// This is necessary as this package is imported by the go client sdk and may lead to issues when used by applications
// that do not build FPC.

// #cgo CFLAGS: -I${SRCDIR}/../../common/crypto
// #cgo LDFLAGS: -L${SRCDIR}/../../common/crypto/_build -L${SRCDIR}/../../common/logging/_build -Wl,--start-group -lupdo-crypto-adapt -lupdo-crypto -Wl,--end-group -lcrypto -lulogging -lstdc++ -lgcov
// #include <stdio.h>
// #include <stdlib.h>
// #include <stdbool.h>
// #include <stdint.h>
// #include "pdo-crypto-c-wrapper.h"
import "C"

var logger = flogging.MustGetLogger("validate")

func ReplayReadWrites(stub shim.ChaincodeStubInterface, fpcrwset *protos.FPCKVSet) (err error) {
	//TODO error checking

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
			k := TransformToFPCKey(rwset.Reads[i].Key)
			v, err := stub.GetState(k)
			if err != nil {
				return fmt.Errorf("error (%s) reading key %s", err, k)
			}

			logger.Debugf("read key %s value(hex) %s", k, hex.EncodeToString(v))

			// compute value hash
			// TODO: use pdo hash for consistency
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
		logger.Debugf("Replaying range queries")
		for _, rqi := range rwset.RangeQueriesInfo {
			if rqi.GetRawReads() == nil {
				// no raw reads available in this range query
				continue
			}
			for _, qr := range rqi.GetRawReads().KvReads {
				k := TransformToFPCKey(qr.Key)
				v, err := stub.GetState(k)
				if err != nil {
					return fmt.Errorf("error (%s) reading key %s", err, k)
				}

				_ = v
				return fmt.Errorf("TODO: not implemented, missing hash check")
			}
		}
	}

	// writes
	if rwset.GetWrites() != nil {
		logger.Debugf("Replaying writes")
		for _, w := range rwset.Writes {
			k := TransformToFPCKey(w.Key)

			// check if composite key, if so, derive Fabric key
			if IsFPCCompositeKey(k) {
				comp := SplitFPCCompositeKey(k)
				k, _ = stub.CreateCompositeKey(comp[0], comp[1:])
			}

			err := stub.PutState(k, w.Value)
			if err != nil {
				return fmt.Errorf("error (%s) writing key %s value(hex) %s", err, k, hex.EncodeToString(w.Value))
			}

			logger.Debugf("written key %s value(hex) %s", k, hex.EncodeToString(w.Value))
		}
	}

	return nil
}

func Validate(signedResponseMessage *protos.SignedChaincodeResponseMessage, attestedData *protos.AttestedData) error {
	if signedResponseMessage.Signature == nil {
		return fmt.Errorf("absent enclave signature")
	}

	// prepare and do signature verification
	enclaveVkPtr := C.CBytes(attestedData.EnclaveVk)
	defer C.free(enclaveVkPtr)

	responseMessagePtr := C.CBytes(signedResponseMessage.ChaincodeResponseMessage)
	defer C.free(responseMessagePtr)

	signaturePtr := C.CBytes(signedResponseMessage.Signature)
	defer C.free(signaturePtr)

	ret := C.verify_signature((*C.uint8_t)(enclaveVkPtr), C.uint32_t(len(attestedData.EnclaveVk)), (*C.uint8_t)(responseMessagePtr), C.uint32_t(len(signedResponseMessage.ChaincodeResponseMessage)), (*C.uint8_t)(signaturePtr), C.uint32_t(len(signedResponseMessage.Signature)))
	if !ret {
		return fmt.Errorf("enclave signature verification failed")
	}

	// verify signed proposal input hash matches input hash
	chaincodeResponseMessage, err := UnmarshalChaincodeResponseMessage(signedResponseMessage.GetChaincodeResponseMessage())
	if err != nil {
		return errors.Wrap(err, "failed to extract response message")
	}

	originalSignedProposal := chaincodeResponseMessage.GetProposal()
	if originalSignedProposal == nil {
		return fmt.Errorf("cannot get the signed proposal that the enclave received")
	}
	chaincodeRequestMessageBytes, err := GetChaincodeRequestMessageFromSignedProposal(originalSignedProposal)
	if err != nil {
		return fmt.Errorf("failed to extract chaincode request message: %s", err)
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
