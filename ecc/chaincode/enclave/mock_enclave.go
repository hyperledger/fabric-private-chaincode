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

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/hyperledger-labs/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger-labs/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/ledger/rwset/kvrwset"
	"github.com/hyperledger/fabric/protoutil"
)

type MockEnclaveStub struct {
	privateKey   []byte
	publicKey    []byte
	enclaveId    string
	ccPrivateKey []byte
}

// NewEnclave starts a new enclave
func NewEnclaveStub() StubInterface {
	return &MockEnclaveStub{}
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
	publicKey, privateKey, err := utils.NewECDSAKeys()
	if err != nil {
		return nil, err
	}
	m.privateKey = privateKey
	m.publicKey = publicKey

	// create chaincode encryption keys keys
	ccPublicKey, ccPrivateKey, err := utils.NewRSAKeys()
	if err != nil {
		return nil, err
	}
	m.ccPrivateKey = ccPrivateKey

	// calculate enclave id
	m.enclaveId, _ = m.GetEnclaveId()

	logger.Debug("Init")
	credentials := &protos.Credentials{
		Attestation: []byte("{\"attestation_type\":\"simulated\",\"attestation\":\"MA==\"}"),
		SerializedAttestedData: &any.Any{
			TypeUrl: proto.MessageName(&protos.AttestedData{}),
			Value: protoutil.MarshalOrPanic(&protos.AttestedData{
				EnclaveVk:   publicKey,
				CcParams:    chaincodeParams,
				HostParams:  hostParams,
				ChaincodeEk: ccPublicKey,
			}),
		},
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

func (m *MockEnclaveStub) ChaincodeInvoke(stub shim.ChaincodeStubInterface, chaincodeRequestMessage []byte) ([]byte, error) {
	logger.Debug("ChaincodeInvoke")

	signedProposal, err := stub.GetSignedProposal()
	if err != nil {
		shim.Error(err.Error())
	}

	// unmarshal request
	chaincodeRequestMessageProto := &protos.ChaincodeRequestMessage{}
	err = proto.Unmarshal(chaincodeRequestMessage, chaincodeRequestMessageProto)
	if err != nil {
		return nil, err
	}

	encryptedRequestBytes := chaincodeRequestMessageProto.GetEncryptedRequest()
	if encryptedRequestBytes == nil {
		return nil, fmt.Errorf("no encrypted request")
	}

	// decrypt request
	requestBytes, err := utils.PkDecryptMessage(m.ccPrivateKey, encryptedRequestBytes)
	if err != nil {
		return nil, err
	}

	requestProto := &protos.CleartextChaincodeRequest{}
	err = proto.Unmarshal(requestBytes, requestProto)
	if err != nil {
		return nil, err
	}

	// get return encryption key
	returnEncryptionKey := requestProto.GetReturnEncryptionKey()
	if returnEncryptionKey == nil {
		return nil, fmt.Errorf("no return encryption key")
	}

	v, _ := stub.GetState("SomeOtherKey")
	v_hash := sha256.Sum256(v)
	logger.Debug("get state: %s", v)

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

	readValueHashes := [][]byte{v_hash[:]}

	fpcKvSet := &protos.FPCKVSet{
		RwSet:           rwset,
		ReadValueHashes: readValueHashes,
	}

	requestMessageHash := sha256.Sum256(chaincodeRequestMessage)

	//create dummy response
	responseData := []byte("some response")

	//response must be encoded
	b64ResponseData := base64.StdEncoding.EncodeToString(responseData)

	//encrypt response
	encryptedResponse, err := utils.EncryptMessage(returnEncryptionKey, []byte(b64ResponseData))
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
	sig, err := utils.SignMessage(m.privateKey, responseBytes)
	if err != nil {
		return nil, err
	}

	signedResponse := &protos.SignedChaincodeResponseMessage{
		ChaincodeResponseMessage: responseBytes,
		Signature:                sig,
	}

	return proto.Marshal(signedResponse)
}
