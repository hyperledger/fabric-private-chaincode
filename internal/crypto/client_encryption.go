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
	"github.com/hyperledger/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/pkg/errors"
)

var logger = flogging.MustGetLogger("fpc-client-crypto")

type EncryptionProvider interface {
	NewEncryptionContext() (EncryptionContext, error)
}

type EncryptionProviderImpl struct {
	CSP                CSP
	GetCcEncryptionKey func() ([]byte, error)
}

func (p EncryptionProviderImpl) NewEncryptionContext() (EncryptionContext, error) {
	// pick request encryption key
	requestEncryptionKey, err := p.CSP.NewSymmetricKey()
	if err != nil {
		return nil, err
	}

	// pick response encryption key
	resultEncryptionKey, err := p.CSP.NewSymmetricKey()
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
		csp:                    p.CSP,
		requestEncryptionKey:   requestEncryptionKey,
		responseEncryptionKey:  resultEncryptionKey,
		chaincodeEncryptionKey: ccEncryptionKey,
	}, nil
}

// EncryptionContext defines the interface of an object responsible to encrypt the contents of a transaction invocation
// and to decrypt the corresponding response.
// Conceal and Reveal must be called only once during the lifetime of an object that implements this interface. That is,
// an EncryptionContext is only valid for a single transaction invocation.
type EncryptionContext interface {
	Conceal(function string, args []string) (string, error)
	Reveal(r []byte) ([]byte, error)
}

type EncryptionContextImpl struct {
	csp                    CSP
	requestEncryptionKey   []byte
	responseEncryptionKey  []byte
	chaincodeEncryptionKey []byte
}

func (e *EncryptionContextImpl) Reveal(signedResponseBytesB64 []byte) ([]byte, error) {
	signedResponseBytes, err := base64.StdEncoding.DecodeString(string(signedResponseBytesB64))
	if err != nil {
		return nil, err
	}

	signedResponse, err := utils.UnmarshalSignedChaincodeResponseMessage(signedResponseBytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to extract signed response message")
	}

	responseBytes := signedResponse.GetChaincodeResponseMessage()
	if responseBytes == nil {
		return nil, fmt.Errorf("no chaincode response message")
	}

	response, err := utils.UnmarshalChaincodeResponseMessage(responseBytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to extract response message")
	}

	clearResponseBytes, err := e.csp.DecryptMessage(e.responseEncryptionKey, response.EncryptedResponse)
	if err != nil {
		return nil, errors.Wrap(err, "decryption of response failed")
	}

	return clearResponseBytes, nil
}

func (e *EncryptionContextImpl) Conceal(function string, args []string) (string, error) {
	args = append([]string{function}, args...)
	bytes := make([][]byte, len(args))
	for i, v := range args {
		bytes[i] = []byte(v)
	}

	// prepare KeyTransportMessage
	keyTransport := &protos.KeyTransportMessage{
		RequestEncryptionKey:  e.requestEncryptionKey,
		ResponseEncryptionKey: e.responseEncryptionKey,
	}

	serializedKeyTransport, err := utils.MarshallProto(keyTransport)
	if err != nil {
		return "", err
	}

	encryptedKeyTransport, err := e.csp.PkEncryptMessage(e.chaincodeEncryptionKey, serializedKeyTransport)
	if err != nil {
		return "", errors.Wrap(err, "encryption of request encryption key failed")
	}

	// prepare CleartextChaincodeRequest
	ccRequest := &protos.CleartextChaincodeRequest{
		Input: &peer.ChaincodeInput{Args: bytes},
	}
	logger.Debugf("prepping chaincode params: %s", ccRequest)

	serializedCcRequest, err := utils.MarshallProto(ccRequest)
	if err != nil {
		return "", err
	}

	encryptedRequest, err := e.csp.EncryptMessage(e.requestEncryptionKey, serializedCcRequest)
	if err != nil {
		return "", errors.Wrap(err, "encryption of request failed")
	}

	// prepare ChaincodeRequestMessage
	encryptedCcRequest := &protos.ChaincodeRequestMessage{
		EncryptedRequest:             encryptedRequest,
		EncryptedKeyTransportMessage: encryptedKeyTransport,
	}

	serializedEncryptedCcRequest, err := utils.MarshallProto(encryptedCcRequest)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(serializedEncryptedCcRequest), nil
}
