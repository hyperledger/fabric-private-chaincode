/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"encoding/base64"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric/protoutil"
)

type Extractors interface {
	GetInitEnclaveMessage(stub shim.ChaincodeStubInterface) (*protos.InitEnclaveMessage, error)
	GetSerializedChaincodeRequest(stub shim.ChaincodeStubInterface) ([]byte, error)
	GetChaincodeResponseMessages(stub shim.ChaincodeStubInterface) (*protos.SignedChaincodeResponseMessage, *protos.ChaincodeResponseMessage, error)
	GetChaincodeParams(stub shim.ChaincodeStubInterface) (*protos.CCParameters, error)
	GetHostParams(stub shim.ChaincodeStubInterface) (*protos.HostParameters, error)
}

type ExtractorImpl struct {
}

func (s *ExtractorImpl) GetInitEnclaveMessage(stub shim.ChaincodeStubInterface) (*protos.InitEnclaveMessage, error) {
	if len(stub.GetStringArgs()) < 2 {
		return nil, fmt.Errorf("initEnclaveMessage missing")
	}

	serializedInitEnclaveMessage, err := base64.StdEncoding.DecodeString(stub.GetStringArgs()[1])
	if err != nil {
		return nil, err
	}

	initMsg, err := utils.UnmarshalInitEnclaveMessage(serializedInitEnclaveMessage)
	if err != nil {
		return nil, err
	}

	return initMsg, err
}

func (s *ExtractorImpl) GetSerializedChaincodeRequest(stub shim.ChaincodeStubInterface) ([]byte, error) {
	if len(stub.GetStringArgs()) < 2 {
		return nil, fmt.Errorf("chaincodeRequestMessage missing")
	}

	chaincodeRequestMessage, err := base64.StdEncoding.DecodeString(stub.GetStringArgs()[1])
	if err != nil {
		return nil, err
	}

	return chaincodeRequestMessage, nil
}

func (s *ExtractorImpl) GetChaincodeResponseMessages(stub shim.ChaincodeStubInterface) (*protos.SignedChaincodeResponseMessage, *protos.ChaincodeResponseMessage, error) {
	if len(stub.GetStringArgs()) < 2 {
		return nil, nil, fmt.Errorf("initEnclaveMessage missing")
	}

	serializedSignedChaincodeResponseMessage, err := base64.StdEncoding.DecodeString(stub.GetStringArgs()[1])
	if err != nil {
		return nil, nil, err
	}

	signedResponseMsg, err := utils.UnmarshalSignedChaincodeResponseMessage(serializedSignedChaincodeResponseMessage)
	if err != nil {
		return nil, nil, err
	}

	responseMsg, err := utils.UnmarshalChaincodeResponseMessage(signedResponseMsg.GetChaincodeResponseMessage())
	if err != nil {
		return nil, nil, err
	}

	return signedResponseMsg, responseMsg, err
}

func (s *ExtractorImpl) GetChaincodeParams(stub shim.ChaincodeStubInterface) (*protos.CCParameters, error) {
	signedProposal, err := stub.GetSignedProposal()
	if err != nil {
		return nil, err
	}

	proposal, err := protoutil.UnmarshalProposal(signedProposal.ProposalBytes)
	if err != nil {
		return nil, err
	}

	cpp, err := protoutil.UnmarshalChaincodeProposalPayload(proposal.Payload)
	if err != nil {
		return nil, err
	}

	cis, err := protoutil.UnmarshalChaincodeInvocationSpec(cpp.Input)
	if err != nil {
		return nil, err
	}

	chaincodeId := cis.ChaincodeSpec.ChaincodeId.Name
	ccDef, err := utils.GetChaincodeDefinition(chaincodeId, stub)
	if err != nil {
		return nil, err
	}

	return &protos.CCParameters{
		ChaincodeId: chaincodeId,
		Version:     ccDef.Version,
		Sequence:    ccDef.Sequence,
		ChannelId:   stub.GetChannelID(),
	}, nil
}

func (s *ExtractorImpl) GetHostParams(stub shim.ChaincodeStubInterface) (*protos.HostParameters, error) {
	initMsg, err := s.GetInitEnclaveMessage(stub)
	if err != nil {
		return nil, err
	}

	mspid, err := cid.GetMSPID(stub)
	if err != nil {
		return nil, err
	}

	if initMsg == nil {
		return nil, fmt.Errorf("initEnclaveMessage is nil")
	}

	return &protos.HostParameters{
		PeerMspId:    mspid,
		PeerEndpoint: initMsg.PeerEndpoint,
		Certificate:  nil, // todo
	}, nil
}
