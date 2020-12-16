// +build mock_ecc

/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package enclave

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
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
	privateKey *ecdsa.PrivateKey
	enclaveId  string
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

	// create some dummy keys for our mock enclave
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	m.privateKey = privateKey

	pubBytes, err := x509.MarshalPKIXPublicKey(privateKey.Public())
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256(pubBytes)
	m.enclaveId = strings.ToUpper(hex.EncodeToString(hash[:]))

	logger.Debug("Init")
	credentials := &protos.Credentials{
		Attestation: []byte("{\"attestation_type\":\"simulated\",\"attestation\":\"MA==\"}"),
		SerializedAttestedData: &any.Any{
			TypeUrl: proto.MessageName(&protos.AttestedData{}),
			Value: protoutil.MarshalOrPanic(&protos.AttestedData{
				EnclaveVk:  pubBytes,
				CcParams:   chaincodeParams,
				HostParams: hostParams,
			}),
		},
	}
	logger.Infof("Create credentials: %s", credentials)

	return proto.Marshal(credentials)
}

func (m MockEnclaveStub) GenerateCCKeys() (*protos.SignedCCKeyRegistrationMessage, error) {
	panic("implement me")
}

func (m MockEnclaveStub) ExportCCKeys(credentials *protos.Credentials) (*protos.SignedExportMessage, error) {
	panic("implement me")
}

func (m MockEnclaveStub) ImportCCKeys() (*protos.SignedCCKeyRegistrationMessage, error) {
	panic("implement me")
}

func (m *MockEnclaveStub) GetEnclaveId() (string, error) {
	pubBytes, err := x509.MarshalPKIXPublicKey(m.privateKey.Public())
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(pubBytes)
	return strings.ToUpper(hex.EncodeToString(hash[:])), nil
}

func (m *MockEnclaveStub) ChaincodeInvoke(stub shim.ChaincodeStubInterface) ([]byte, error) {
	logger.Debug("ChaincodeInvoke")

	signedProposal, err := stub.GetSignedProposal()
	if err != nil {
		shim.Error(err.Error())
	}

	v, _ := stub.GetState("SomeOtherKey")
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

	response := &protos.ChaincodeResponseMessage{
		EncryptedResponse: []byte("some response"),
		RwSet:             rwset,
		Signature:         nil,
		EnclaveId:         m.enclaveId,
		Proposal:          signedProposal,
	}

	// get the read/write set in the same format as processed by the chaincode enclaves
	readset, writeset, err := utils.ReplayReadWrites(stub, response.RwSet)
	if err != nil {
		shim.Error(err.Error())
	}

	// create signature
	hash := utils.ComputedHash(response, readset, writeset)
	sig, err := ecdsa.SignASN1(rand.Reader, m.privateKey, hash[:])
	if err != nil {
		return nil, err
	}

	response.Signature = sig

	return proto.Marshal(response)
}
