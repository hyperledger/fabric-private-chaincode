/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package crypto

import (
	"encoding/base64"

	"github.com/hyperledger-labs/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/flogging"
	"google.golang.org/protobuf/proto"

	"fmt"
)

// #cgo CFLAGS: -I${SRCDIR}/../../../../common/crypto
// #cgo LDFLAGS: -L${SRCDIR}/../../../../common/crypto/_build -L${SRCDIR}/../../../../common/logging/_build -Wl,--start-group -lupdo-crypto-adapt -lupdo-crypto -Wl,--end-group -lcrypto -lulogging -lstdc++
// #include <stdio.h>
// #include <stdlib.h>
// #include <stdbool.h>
// #include <stdint.h>
// #include "pdo-crypto-c-wrapper.h"
import "C"

var logger = flogging.MustGetLogger("fpc-client-crypto")

func encrypt(input []byte, encryptionKey []byte) ([]byte, error) {
	return input, nil
}

func keyGen() ([]byte, error) {
	return []byte("fake key"), nil
}

func decrypt(encryptedResponse []byte, resultEncryptionKey []byte) ([]byte, error) {
	return encryptedResponse, nil
}

type EncryptionProvider interface {
	NewEncryptionContext() (EncryptionContext, error)
}

type EncryptionProviderImpl struct {
	GetCcEncryptionKey func() ([]byte, error)
}

func (e EncryptionProviderImpl) NewEncryptionContext() (EncryptionContext, error) {
	// pick response encryption key
	resultEncryptionKey, err := keyGen()
	if err != nil {
		return nil, err
	}

	ccEncryptionKey, err := e.GetCcEncryptionKey()
	if err != nil {
		return nil, err
	}

	return &EncryptionContextImpl{
		resultEncryptionKey:    resultEncryptionKey,
		chaincodeEncryptionKey: ccEncryptionKey,
	}, nil
}

// EncryptionContext defines the interface of an object responsible to encrypt the contents of a transaction invocation
// and decrypt the corresponding response.
// Conceal and Reveal must be called only once during the lifetime of an object that implements this interface. That is,
// an EncryptionContext is only valid for a single transaction invocation.
type EncryptionContext interface {
	Conceal(function string, args []string) (string, error)
	Reveal(r []byte) ([]byte, error)
}

type EncryptionContextImpl struct {
	resultEncryptionKey    []byte
	chaincodeEncryptionKey []byte
}

func (e *EncryptionContextImpl) Reveal(signedResponseBytesB64 []byte) ([]byte, error) {
	signedResponseBytes, err := base64.StdEncoding.DecodeString(string(signedResponseBytesB64))
	if err != nil {
		return nil, err
	}

	signedResponse := &protos.SignedChaincodeResponseMessage{}
	err = proto.Unmarshal(signedResponseBytes, signedResponse)
	if err != nil {
		return nil, err
	}

	responseBytes := signedResponse.GetChaincodeResponseMessage()
	if responseBytes == nil {
		return nil, fmt.Errorf("no chaincode response message")
	}

	response := &protos.ChaincodeResponseMessage{}
	err = proto.Unmarshal(responseBytes, response)
	if err != nil {
		return nil, err
	}

	clearResponseB64, err := decrypt(response.EncryptedResponse, e.resultEncryptionKey)
	if err != nil {
		return nil, err
	}
	// TODO: above should eventually be a (protobuf but not base64 serialized) fabric response object,
	//   rather than just the (base64-serialized) response string.
	//   so we also get fpc chaincode return-code/error-message as in for normal fabric
	clearResponse, err := base64.StdEncoding.DecodeString(string(clearResponseB64))
	if err != nil {
		return nil, err
	}

	return clearResponse, nil
}

func (e *EncryptionContextImpl) Conceal(function string, args []string) (string, error) {
	args = append([]string{function}, args...)
	bytes := make([][]byte, len(args))
	for i, v := range args {
		bytes[i] = []byte(v)
	}

	ccRequest := &protos.CleartextChaincodeRequest{
		Input:               &peer.ChaincodeInput{Args: bytes},
		ReturnEncryptionKey: e.resultEncryptionKey,
	}
	logger.Debugf("prepping chaincode params: %s", ccRequest)

	serializedCcRequest, err := proto.Marshal(ccRequest)
	if err != nil {
		return "", err
	}

	encryptedParams, err := encrypt(serializedCcRequest, e.chaincodeEncryptionKey)
	if err != nil {
		return "", err
	}

	encryptedCcRequest := &protos.ChaincodeRequestMessage{
		EncryptedRequest: encryptedParams,
	}

	serializedEncryptedCcRequest, err := proto.Marshal(encryptedCcRequest)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(serializedEncryptedCcRequest), nil
}
