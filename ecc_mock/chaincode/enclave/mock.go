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

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/hyperledger-labs/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger-labs/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/ledger/rwset/kvrwset"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/protoutil"
)

var logger = flogging.MustGetLogger("enclave")

type MockEnclave struct {
	privateKey *ecdsa.PrivateKey
	enclaveId  string
}

func (m *MockEnclave) Init(chaincodeParams *protos.CCParameters, hostParams *protos.HostParameters, attestationParams []byte) ([]byte, error) {
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
	m.enclaveId = hex.EncodeToString(hash[:])

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

func (m MockEnclave) GenerateCCKeys() (*protos.SignedCCKeyRegistrationMessage, error) {
	panic("implement me")
}

func (m MockEnclave) ExportCCKeys(credentials *protos.Credentials) (*protos.SignedExportMessage, error) {
	panic("implement me")
}

func (m MockEnclave) ImportCCKeys() (*protos.SignedCCKeyRegistrationMessage, error) {
	panic("implement me")
}

func (m *MockEnclave) GetEnclaveId() (string, error) {
	pubBytes, err := x509.MarshalPKIXPublicKey(m.privateKey.Public())
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(pubBytes)
	return hex.EncodeToString(hash[:]), nil
}

func (m *MockEnclave) ChaincodeInvoke(stub shim.ChaincodeStubInterface) ([]byte, error) {
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
	readset, writeset, err := utils.PerformReadWrites(stub, response.RwSet)
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
