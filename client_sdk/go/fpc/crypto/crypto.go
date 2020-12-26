/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package crypto

import (
	"crypto/rand"
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

func keyGen() ([]byte, error) {
	// generate random symmetric key of 16 bytes
	// the length is required by the pdo crypto library
	key := make([]byte, 16)
	n, err := rand.Read(key)
	if n != len(key) || err != nil {
		return nil, err
	}

	return key, nil
}

func encrypt(input []byte, encryptionKey []byte) ([]byte, error) {
	//This is an RSA encryption performed with the pdo crypto library
	//Importantly, the library uses 2048bit RSA keys, so the input size has to be ~200bytes
	//TODO-1: bump up the key length to 3072
	//TODO-2: extend procedure for large input sizes

	inputMessagePtr := C.CBytes(input)
	defer C.free(inputMessagePtr)

	encryptionKeyPtr := C.CBytes(encryptionKey)
	defer C.free(encryptionKeyPtr)

	//pdo crypto uses 2048bit (256bytes) RSA keys, the encrypted message size buffer is set accordingly
	const encryptedMessageSize = 256
	encryptedMessagePtr := C.malloc(encryptedMessageSize)
	defer C.free(encryptedMessagePtr)

	encryptedMessageActualSize := C.uint32_t(0)

	ret := C.pk_encrypt_message((*C.uint8_t)(encryptionKeyPtr), C.uint32_t(len(encryptionKey)), (*C.uint8_t)(inputMessagePtr), C.uint32_t(len(input)), (*C.uint8_t)(encryptedMessagePtr), C.uint32_t(encryptedMessageSize), &encryptedMessageActualSize)
	if ret == false {
		return nil, fmt.Errorf("encryption failed")
	}

	return C.GoBytes(encryptedMessagePtr, C.int(encryptedMessageActualSize)), nil
}

func decrypt(encryptedResponse []byte, resultEncryptionKey []byte) ([]byte, error) {
	encryptedResponsePtr := C.CBytes(encryptedResponse)
	defer C.free(encryptedResponsePtr)

	resultEncryptionKeyPtr := C.CBytes(resultEncryptionKey)
	defer C.free(resultEncryptionKeyPtr)

	// the (decrypted) response size is estimated to be <= the encrypted response size
	responseSize := len(encryptedResponse)
	responsePtr := C.malloc(C.ulong(responseSize))
	defer C.free(responsePtr)

	responseActualSize := C.uint32_t(0)

	ret := C.decrypt_message((*C.uint8_t)(resultEncryptionKeyPtr), C.uint32_t(len(resultEncryptionKey)), (*C.uint8_t)(encryptedResponsePtr), C.uint32_t(len(encryptedResponse)), (*C.uint8_t)(responsePtr), C.uint32_t(responseSize), &responseActualSize)
	if ret == false {
		return nil, fmt.Errorf("decryption failed")
	}

	return C.GoBytes(responsePtr, C.int(responseActualSize)), nil
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
	//decode key
	ccEncryptionKey, err = base64.StdEncoding.DecodeString(string(ccEncryptionKey))
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
