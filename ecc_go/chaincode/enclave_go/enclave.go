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
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-private-chaincode/internal/crypto"
	"github.com/hyperledger/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

var logger = flogging.MustGetLogger("enclave_go")

type EnclaveStub struct {
	csp             crypto.CSP
	ccRef           shim.Chaincode
	identity        *EnclaveIdentity
	ccIdentity      *ChaincodeIdentity
	hostParams      *protos.HostParameters
	chaincodeParams *protos.CCParameters
}

func NewEnclaveStub(cc shim.Chaincode) *EnclaveStub {
	return &EnclaveStub{
		csp:   crypto.GetDefaultCSP(),
		ccRef: cc,
	}
}

func (e *EnclaveStub) Init(serializedChaincodeParams, serializedHostParamsBytes, serializedAttestationParams []byte) ([]byte, error) {
	logger.Debug("Init enclave")

	var err error

	// generate new enclave identity
	e.identity, err = NewEnclaveIdentity(e.csp)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create new enclave identity")
	}

	// as we currently support a single enclave instance per chaincode, we also generate a new chaincode identity here
	// this needs to be refactored once multi enclave support will be integrated
	e.ccIdentity, err = NewChaincodeIdentity(e.csp)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create new enclave identity")
	}

	e.hostParams = &protos.HostParameters{}
	if err := proto.Unmarshal(serializedHostParamsBytes, e.hostParams); err != nil {
		return nil, err
	}

	e.chaincodeParams = &protos.CCParameters{}
	if err := proto.Unmarshal(serializedChaincodeParams, e.chaincodeParams); err != nil {
		return nil, err
	}

	serializedAttestedData, _ := anypb.New(&protos.AttestedData{
		EnclaveVk:   e.identity.GetPublicKey(),
		CcParams:    e.chaincodeParams,
		HostParams:  e.hostParams,
		ChaincodeEk: e.ccIdentity.GetPublicKey(),
	})

	attestation, err := createAttestation(serializedAttestedData)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create attestation")
	}

	credentials := &protos.Credentials{
		Attestation:            attestation,
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
	if e.identity == nil {
		return "", fmt.Errorf("enclave not yet initliazed")
	}

	return e.identity.GetEnclaveId(), nil
}

func (e *EnclaveStub) ChaincodeInvoke(stub shim.ChaincodeStubInterface, chaincodeRequestMessageBytes []byte) ([]byte, error) {
	logger.Debug("ChaincodeInvoke")

	signedProposal, err := stub.GetSignedProposal()
	if err != nil {
		shim.Error(err.Error())
	}

	if err := e.verifySignedProposal(stub, chaincodeRequestMessageBytes); err != nil {
		return nil, errors.Wrap(err, "signed proposal verification failed")
	}

	// unmarshal chaincodeRequest
	chaincodeRequestMessage := &protos.ChaincodeRequestMessage{}
	err = proto.Unmarshal(chaincodeRequestMessageBytes, chaincodeRequestMessage)
	if err != nil {
		return nil, err
	}

	// get key transport message including the encryption keys for request and response
	keyTransportMessage, err := e.extractKeyTransportMessage(chaincodeRequestMessage)
	if err != nil {
		return nil, errors.Wrap(err, "cannot extract keyTransportMessage")
	}

	// decrypt request
	cleartextChaincodeRequest, err := e.extractCleartextChaincodeRequest(chaincodeRequestMessage, keyTransportMessage)
	if err != nil {
		return nil, errors.Wrap(err, "cannot decrypt chaincode request")
	}

	// create a new instance of a FPC RWSet that we pass to the stub and later return with the response
	rwset := newReadWriteSet()

	// Invoke chaincode
	// we wrap the stub with our FpcStubInterface
	fpcStub := NewFpcStubInterface(stub, cleartextChaincodeRequest.GetInput(), rwset, e.ccIdentity)
	ccResponse := e.ccRef.Invoke(fpcStub)

	// If payload is empty (probably due to a shim.Error), the response will contain the message
	var b64ResponseData string
	if ccResponse.GetPayload() != nil {
		b64ResponseData = base64.StdEncoding.EncodeToString(ccResponse.GetPayload())
	} else {
		b64ResponseData = base64.StdEncoding.EncodeToString([]byte(ccResponse.GetMessage()))
	}

	//encrypt response
	encryptedResponse, err := e.csp.EncryptMessage(keyTransportMessage.GetResponseEncryptionKey(), []byte(b64ResponseData))
	if err != nil {
		return nil, err
	}

	chaincodeRequestMessageHash := sha256.Sum256(chaincodeRequestMessageBytes)

	response := &protos.ChaincodeResponseMessage{
		EncryptedResponse:           encryptedResponse,
		FpcRwSet:                    rwset.toFPCKVSet(),
		EnclaveId:                   e.identity.GetEnclaveId(),
		Proposal:                    signedProposal,
		ChaincodeRequestMessageHash: chaincodeRequestMessageHash[:],
	}

	responseBytes, err := proto.Marshal(response)
	if err != nil {
		return nil, err
	}

	// create signature
	sig, err := e.identity.Sign(responseBytes)
	if err != nil {
		return nil, err
	}

	signedResponse := &protos.SignedChaincodeResponseMessage{
		ChaincodeResponseMessage: responseBytes,
		Signature:                sig,
	}

	return proto.Marshal(signedResponse)
}

func (e *EnclaveStub) verifySignedProposal(stub shim.ChaincodeStubInterface, chaincodeRequestMessageBytes []byte) error {
	signedProposal, err := stub.GetSignedProposal()
	if err != nil {
		return err
	}

	proposal, err := protoutil.UnmarshalProposal(signedProposal.GetProposalBytes())
	if err != nil {
		return errors.Wrap(err, "cannot unmarshal proposal")
	}

	header, err := protoutil.UnmarshalHeader(proposal.GetHeader())
	if err != nil {
		return errors.Wrap(err, "cannot unmarshal proposal header")
	}

	channelHeader, err := protoutil.UnmarshalChannelHeader(header.GetChannelHeader())
	if err != nil {
		return errors.Wrap(err, "cannot unmarshal channel header")
	}

	if channelHeader.GetChannelId() != e.chaincodeParams.ChaincodeId {
		return fmt.Errorf("channel id of the tx proposal does not match as initialized with cc_parameters")
	}

	// TODO perform signature check
	//signature := signedProposal.GetSignature()
	// our chance to also support Idemix signatures here

	return nil
}

func (e *EnclaveStub) extractKeyTransportMessage(chaincodeRequestMessage *protos.ChaincodeRequestMessage) (*protos.KeyTransportMessage, error) {
	if chaincodeRequestMessage == nil {
		return nil, fmt.Errorf("chaincodeRequestMessage is nil")
	}

	if chaincodeRequestMessage.GetEncryptedKeyTransportMessage() == nil {
		return nil, fmt.Errorf("chaincodeRequestMessages does not contain a encrypted keyTransportMessage")
	}

	// decrypt key transport message with chaincode decryption key
	keyTransportMessageBytes, err := e.ccIdentity.PkDecryptMessage(chaincodeRequestMessage.GetEncryptedKeyTransportMessage())
	if err != nil {
		return nil, errors.Wrap(err, "decryption of key transport message failed")
	}

	keyTransportMessage := &protos.KeyTransportMessage{}
	if err := proto.Unmarshal(keyTransportMessageBytes, keyTransportMessage); err != nil {
		return nil, err
	}

	// check that we have booth, request and response encryption key
	if keyTransportMessage.GetRequestEncryptionKey() == nil {
		return nil, fmt.Errorf("no request encryption key")
	}

	if keyTransportMessage.GetRequestEncryptionKey() == nil {
		return nil, fmt.Errorf("no response encryption key")
	}
	return keyTransportMessage, err
}

func (e *EnclaveStub) extractCleartextChaincodeRequest(chaincodeRequestMessage *protos.ChaincodeRequestMessage, keyTransportMessage *protos.KeyTransportMessage) (*protos.CleartextChaincodeRequest, error) {
	if chaincodeRequestMessage.GetEncryptedRequest() == nil {
		return nil, fmt.Errorf("no encrypted request")
	}

	if keyTransportMessage.GetRequestEncryptionKey() == nil {
		return nil, fmt.Errorf("no encryption key")
	}

	// decrypt request
	clearChaincodeRequestBytes, err := e.csp.DecryptMessage(keyTransportMessage.GetRequestEncryptionKey(), chaincodeRequestMessage.GetEncryptedRequest())
	if err != nil {
		return nil, errors.Wrap(err, "decryption of request failed")
	}

	// unmarshal cleartextChaincodeRequest
	cleartextChaincodeRequest := &protos.CleartextChaincodeRequest{}
	err = proto.Unmarshal(clearChaincodeRequestBytes, cleartextChaincodeRequest)
	if err != nil {
		return nil, err
	}

	return cleartextChaincodeRequest, nil
}

func createAttestation(attestedData *anypb.Any) ([]byte, error) {
	// TODO implement proper attestation, supporting dcap and simulation
	return []byte("{\"attestation_type\":\"simulated\",\"attestation\":\"MA==\"}"), nil
}
