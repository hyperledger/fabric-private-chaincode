/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package crypto

import (
	"encoding/base64"
	"log"

	"github.com/hyperledger-labs/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger/fabric-protos-go/peer"
	"google.golang.org/protobuf/proto"
)

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

func (e *EncryptionContextImpl) Reveal(responseBytes []byte) ([]byte, error) {
	response := &protos.ChaincodeResponseMessage{}
	err := proto.Unmarshal(responseBytes, response)
	if err != nil {
		return nil, err
	}

	return decrypt(response.EncryptedResponse, e.resultEncryptionKey)
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
	log.Printf("prepping chaincode params: %s\n", ccRequest)

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
