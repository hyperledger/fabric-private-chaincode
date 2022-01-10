/*
Copyright Riccardo Zappoli (riccardo.zappoli@unifr.ch)
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package enclave_go

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

var logger = flogging.MustGetLogger("enclave_go")

type EnclaveStub struct {
	csp          crypto.CSP
	privateKey   []byte
	publicKey    []byte
	enclaveId    string
	ccPrivateKey []byte
	stateKey     []byte
	ccRef        shim.Chaincode
}

func NewEnclaveStub(cc shim.Chaincode) *EnclaveStub {
	return &EnclaveStub{
		csp:   crypto.GetDefaultCSP(),
		ccRef: cc,
	}
}

func (e *EnclaveStub) Init(serializedChaincodeParams, serializedHostParamsBytes, serializedAttestationParams []byte) ([]byte, error) {

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
	publicKey, privateKey, err := e.csp.NewECDSAKeys()
	if err != nil {
		return nil, err
	}
	e.privateKey = privateKey
	e.publicKey = publicKey

	// create chaincode encryption keys keys
	ccPublicKey, ccPrivateKey, err := e.csp.NewRSAKeys()
	if err != nil {
		return nil, err
	}
	e.ccPrivateKey = ccPrivateKey

	// create state key
	stateKey, err := e.csp.NewSymmetricKey()
	if err != nil {
		return nil, err
	}
	e.stateKey = stateKey

	// calculate enclave id
	e.enclaveId, _ = e.GetEnclaveId()

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

func (e EnclaveStub) GenerateCCKeys() ([]byte, error) {
	panic("implement me")
	// -> *protos.SignedCCKeyRegistrationMessage
}

func (e EnclaveStub) ExportCCKeys(credentials []byte) ([]byte, error) {
	panic("implement me")
	// credentials *protos.Credentials -> *protos.SignedExportMessage,
}

func (e EnclaveStub) ImportCCKeys() ([]byte, error) {
	panic("implement me")
	// -> *protos.SignedCCKeyRegistrationMessage
}

func (e *EnclaveStub) GetEnclaveId() (string, error) {
	hash := sha256.Sum256(e.publicKey)
	return strings.ToUpper(hex.EncodeToString(hash[:])), nil
}

func (e *EnclaveStub) ChaincodeInvoke(stub shim.ChaincodeStubInterface, chaincodeRequestMessageBytes []byte) ([]byte, error) {
	logger.Debug("ChaincodeInvoke")

	signedProposal, err := stub.GetSignedProposal()
	if err != nil {
		shim.Error(err.Error())
	}
	chaincodeRequestMessageHash := sha256.Sum256(chaincodeRequestMessageBytes)

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
	keyTransportMessageBytes, err := e.csp.PkDecryptMessage(e.ccPrivateKey, chaincodeRequestMessage.GetEncryptedKeyTransportMessage())
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
	clearChaincodeRequestBytes, err := e.csp.DecryptMessage(keyTransportMessage.GetRequestEncryptionKey(), chaincodeRequestMessage.GetEncryptedRequest())
	if err != nil {
		return nil, errors.Wrap(err, "decryption of request failed")
	}

	cleartextChaincodeRequest := &protos.CleartextChaincodeRequest{}
	err = proto.Unmarshal(clearChaincodeRequestBytes, cleartextChaincodeRequest)
	if err != nil {
		return nil, err
	}

	// construct rwset
	rwset := &kvrwset.KVRWSet{
		Reads:  []*kvrwset.KVRead{},
		Writes: []*kvrwset.KVWrite{},
	}
	fpcKvSet := &protos.FPCKVSet{
		RwSet:           rwset,
		ReadValueHashes: [][]byte{},
	}

	// Invoke Simple Chaincode
	fpcStub := NewFpcStubInterface(stub, cleartextChaincodeRequest.GetInput(), fpcKvSet, e.stateKey)
	peerResponse := e.ccRef.Invoke(fpcStub)

	fmt.Println(fpcKvSet)

	// If payload is empty (probably due to a shim.Error), the response will contain the message
	var b64ResponseData string
	if peerResponse.GetPayload() != nil {
		b64ResponseData = base64.StdEncoding.EncodeToString(peerResponse.GetPayload())
	} else {
		b64ResponseData = base64.StdEncoding.EncodeToString([]byte(peerResponse.GetMessage()))
	}

	//encrypt response
	encryptedResponse, err := e.csp.EncryptMessage(keyTransportMessage.GetResponseEncryptionKey(), []byte(b64ResponseData))
	if err != nil {
		return nil, err
	}

	response := &protos.ChaincodeResponseMessage{
		EncryptedResponse:           encryptedResponse,
		FpcRwSet:                    fpcKvSet,
		EnclaveId:                   e.enclaveId,
		Proposal:                    signedProposal,
		ChaincodeRequestMessageHash: chaincodeRequestMessageHash[:],
	}

	responseBytes, err := proto.Marshal(response)
	if err != nil {
		return nil, err
	}

	// create signature
	sig, err := e.csp.SignMessage(e.privateKey, responseBytes)
	if err != nil {
		return nil, err
	}

	signedResponse := &protos.SignedChaincodeResponseMessage{
		ChaincodeResponseMessage: responseBytes,
		Signature:                sig,
	}

	return proto.Marshal(signedResponse)
}
