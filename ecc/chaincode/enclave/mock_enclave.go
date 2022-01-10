//go:build mock_ecc
// +build mock_ecc

/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package enclave

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-private-chaincode/internal/crypto"
	"github.com/hyperledger/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger/fabric-protos-go/ledger/rwset/kvrwset"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

var logger = flogging.MustGetLogger("mock_enclave")

type MockEnclaveStub struct {
	csp          crypto.CSP
	privateKey   []byte
	publicKey    []byte
	enclaveId    string
	ccPrivateKey []byte
}

func NewEnclaveStub() *MockEnclaveStub {
	return &MockEnclaveStub{
		csp: crypto.GetDefaultCSP(),
	}
}

func (m *MockEnclaveStub) Init(serializedChaincodeParams, serializedHostParamsBytes, serializedAttestationParams []byte) ([]byte, error) {

	hostParams := &protos.HostParameters{}
	err := proto.Unmarshal(serializedHostParamsBytes, hostParams)
	if err != nil {
		return nil, err
	}

	chaincodeParams := &protos.CCParameters{}
	err = proto.Unmarshal(serializedChaincodeParams, chaincodeParams)
	if err != nil {
		return nil, err
	}

	// create enclave keys
	publicKey, privateKey, err := m.csp.NewECDSAKeys()
	if err != nil {
		return nil, err
	}
	m.privateKey = privateKey
	m.publicKey = publicKey

	// create chaincode encryption keys keys
	ccPublicKey, ccPrivateKey, err := m.csp.NewRSAKeys()
	if err != nil {
		return nil, err
	}
	m.ccPrivateKey = ccPrivateKey

	// calculate enclave id
	m.enclaveId, _ = m.GetEnclaveId()

	logger.Debug("Init")

	serializedAttestedData, _ := anypb.New(&protos.AttestedData{
		EnclaveVk:   publicKey,
		CcParams:    chaincodeParams,
		HostParams:  hostParams,
		ChaincodeEk: ccPublicKey,
	})
	credentials := &protos.Credentials{
		Attestation:            []byte("{\"attestation_type\":\"simulated\",\"attestation\":\"MA==\"}"),
		SerializedAttestedData: serializedAttestedData,
	}
	logger.Infof("Create credentials: %s", credentials)

	return proto.Marshal(credentials)
}

func (m MockEnclaveStub) GenerateCCKeys() ([]byte, error) {
	panic("implement me")
	// -> *protos.SignedCCKeyRegistrationMessage
}

func (m MockEnclaveStub) ExportCCKeys(credentials []byte) ([]byte, error) {
	panic("implement me")
	// credentials *protos.Credentials -> *protos.SignedExportMessage,
}

func (m MockEnclaveStub) ImportCCKeys() ([]byte, error) {
	panic("implement me")
	// -> *protos.SignedCCKeyRegistrationMessage
}

func (m *MockEnclaveStub) GetEnclaveId() (string, error) {
	hash := sha256.Sum256(m.publicKey)
	return strings.ToUpper(hex.EncodeToString(hash[:])), nil
}

func (m *MockEnclaveStub) ChaincodeInvoke(stub shim.ChaincodeStubInterface, chaincodeRequestMessageBytes []byte) ([]byte, error) {
	logger.Debug("ChaincodeInvoke")

	signedProposal, err := stub.GetSignedProposal()
	if err != nil {
		shim.Error(err.Error())
	}

	// unmarshal chaincodeRequest
	chaincodeRequestMessage := &protos.ChaincodeRequestMessage{}
	err = proto.Unmarshal(chaincodeRequestMessageBytes, chaincodeRequestMessage)
	if err != nil {
		return nil, err
	}

	if chaincodeRequestMessage.GetEncryptedRequest() == nil {
		return nil, fmt.Errorf("no encrypted request")
	}

	if chaincodeRequestMessage.GetEncryptedKeyTransportMessage() == nil {
		return nil, fmt.Errorf("no encrypted key transport message")
	}

	// decrypt key transport message with chaincode decryption key
	keyTransportMessageBytes, err := m.csp.PkDecryptMessage(m.ccPrivateKey, chaincodeRequestMessage.GetEncryptedKeyTransportMessage())
	if err != nil {
		return nil, errors.Wrap(err, "decryption of key transport message failed")
	}

	keyTransportMessage := &protos.KeyTransportMessage{}
	err = proto.Unmarshal(keyTransportMessageBytes, keyTransportMessage)
	if err != nil {
		return nil, err
	}

	// check that we have booth, request and response encryption key
	if keyTransportMessage.GetRequestEncryptionKey() == nil {
		return nil, fmt.Errorf("no request encryption key")
	}

	if keyTransportMessage.GetRequestEncryptionKey() == nil {
		return nil, fmt.Errorf("no response encryption key")
	}

	// decrypt request
	clearChaincodeRequestBytes, err := m.csp.DecryptMessage(keyTransportMessage.GetRequestEncryptionKey(), chaincodeRequestMessage.GetEncryptedRequest())
	if err != nil {
		return nil, errors.Wrap(err, "decryption of request failed")
	}

	cleartextChaincodeRequest := &protos.CleartextChaincodeRequest{}
	err = proto.Unmarshal(clearChaincodeRequestBytes, cleartextChaincodeRequest)
	if err != nil {
		return nil, err
	}

	// do some read/write ops
	_ = stub.PutState("SomeOtherKey", []byte("some value"))
	v, _ := stub.GetState("helloKey")
	v_hash := sha256.Sum256(v)
	logger.Debugf("get state: %s with hash %s", v, v_hash)

	// construct rwset for validation
	rwset := &kvrwset.KVRWSet{
		Reads: []*kvrwset.KVRead{{
			Key:     "helloKey",
			Version: nil,
		}},
		Writes: []*kvrwset.KVWrite{{
			Key:      "SomeOtherKey",
			IsDelete: false,
			Value:    []byte("some value"),
		}},
	}

	fpcKvSet := &protos.FPCKVSet{
		RwSet:           rwset,
		ReadValueHashes: [][]byte{v_hash[:]},
	}

	requestMessageHash := sha256.Sum256(chaincodeRequestMessageBytes)

	//create dummy response
	responseData := []byte("some response")

	//response must be encoded
	b64ResponseData := base64.StdEncoding.EncodeToString(responseData)

	//encrypt response
	encryptedResponse, err := m.csp.EncryptMessage(keyTransportMessage.GetResponseEncryptionKey(), []byte(b64ResponseData))
	if err != nil {
		return nil, err
	}

	response := &protos.ChaincodeResponseMessage{
		EncryptedResponse:           encryptedResponse,
		FpcRwSet:                    fpcKvSet,
		EnclaveId:                   m.enclaveId,
		Proposal:                    signedProposal,
		ChaincodeRequestMessageHash: requestMessageHash[:],
	}

	responseBytes, err := proto.Marshal(response)
	if err != nil {
		return nil, err
	}

	// create signature
	sig, err := m.csp.SignMessage(m.privateKey, responseBytes)
	if err != nil {
		return nil, err
	}

	signedResponse := &protos.SignedChaincodeResponseMessage{
		ChaincodeResponseMessage: responseBytes,
		Signature:                sig,
	}

	return proto.Marshal(signedResponse)
}
