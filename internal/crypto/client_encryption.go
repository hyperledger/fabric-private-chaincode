/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package crypto

import (
	"encoding/base64"
	"fmt"

	"github.com/hyperledger/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

var logger = flogging.MustGetLogger("fpc-client-crypto")

type EncryptionProvider interface {
	NewEncryptionContext() (EncryptionContext, error)
}

type EncryptionProviderImpl struct {
	GetCcEncryptionKey func() ([]byte, error)
}

func (p EncryptionProviderImpl) NewEncryptionContext() (EncryptionContext, error) {
	// pick request encryption key
	requestEncryptionKey, err := NewSymmetricKey()
	if err != nil {
		return nil, err
	}

	// pick response encryption key
	resultEncryptionKey, err := NewSymmetricKey()
	if err != nil {
		return nil, err
	}

	ccEncryptionKey, err := p.GetCcEncryptionKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get chaincode encryption key from ercc: %s", err.Error())
	}
	//decode key
	ccEncryptionKey, err = base64.StdEncoding.DecodeString(string(ccEncryptionKey))
	if err != nil {
		return nil, err
	}

	return &EncryptionContextImpl{
		requestEncryptionKey:   requestEncryptionKey,
		responseEncryptionKey:  resultEncryptionKey,
		chaincodeEncryptionKey: ccEncryptionKey,
	}, nil
}

// EncryptionContext defines the interface of an object responsible to encrypt the contents of a transaction invocation
// and DecryptMessage the corresponding response.
// Conceal and Reveal must be called only once during the lifetime of an object that implements this interface. That is,
// an EncryptionContext is only valid for a single transaction invocation.
type EncryptionContext interface {
	Conceal(function string, args []string) (string, error)
	Reveal(r []byte) ([]byte, error)
}

type EncryptionContextImpl struct {
	requestEncryptionKey   []byte
	responseEncryptionKey  []byte
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

	clearResponseB64, err := DecryptMessage(e.responseEncryptionKey, response.EncryptedResponse)
	if err != nil {
		return nil, errors.Wrap(err, "decryption of response failed")
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
		ReturnEncryptionKey: e.responseEncryptionKey,
	}
	logger.Debugf("prepping chaincode params: %s", ccRequest)

	serializedCcRequest, err := proto.Marshal(ccRequest)
	if err != nil {
		return "", err
	}

	encryptedRequest, err := EncryptMessage(e.requestEncryptionKey, serializedCcRequest)
	if err != nil {
		return "", errors.Wrap(err, "encryption of request failed")
	}

	encryptedRequestEncryptionKey, err := PkEncryptMessage(e.chaincodeEncryptionKey, e.requestEncryptionKey)
	if err != nil {
		return "", errors.Wrap(err, "encryption of request encryption key failed")
	}

	encryptedCcRequest := &protos.ChaincodeRequestMessage{
		EncryptedRequest:              encryptedRequest,
		EncryptedRequestEncryptionKey: encryptedRequestEncryptionKey,
	}

	serializedEncryptedCcRequest, err := proto.Marshal(encryptedCcRequest)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(serializedEncryptedCcRequest), nil
}
