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
	"google.golang.org/protobuf/proto"
)

func extractChaincodeParams(stub shim.ChaincodeStubInterface) (*protos.CCParameters, error) {
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

	return &protos.CCParameters{
		ChaincodeId: chaincodeId,
		Version:     ccDef.Version,
		Sequence:    ccDef.Sequence,
		ChannelId:   stub.GetChannelID(),
	}, nil
}

func extractHostParams(stub shim.ChaincodeStubInterface, initMsg *protos.InitEnclaveMessage) (*protos.HostParameters, error) {
	mspid, err := cid.GetMSPID(stub)
	if err != nil {
		return nil, err
	}

	return &protos.HostParameters{
		PeerMspId:    mspid,
		PeerEndpoint: initMsg.PeerEndpoint,
		Certificate:  nil, // todo
	}, nil
}

func extractInitEnclaveMessage(stub shim.ChaincodeStubInterface) (*protos.InitEnclaveMessage, error) {
	if len(stub.GetStringArgs()) < 2 {
		return nil, fmt.Errorf("initEnclaveMessage missing")
	}

	serializedInitEnclaveMessage, err := base64.StdEncoding.DecodeString(stub.GetStringArgs()[1])
	if err != nil {
		return nil, err
	}

	initMsg := &protos.InitEnclaveMessage{}
	err = proto.Unmarshal(serializedInitEnclaveMessage, initMsg)
	if err != nil {
		return nil, err
	}

	return initMsg, err
}

func extractChaincodeResponseMessages(stub shim.ChaincodeStubInterface) (*protos.SignedChaincodeResponseMessage, *protos.ChaincodeResponseMessage, error) {
	if len(stub.GetStringArgs()) < 2 {
		return nil, nil, fmt.Errorf("initEnclaveMessage missing")
	}

	serializedSignedChaincodeResponseMessage, err := base64.StdEncoding.DecodeString(stub.GetStringArgs()[1])
	if err != nil {
		return nil, nil, err
	}

	signedResponseMsg := &protos.SignedChaincodeResponseMessage{}
	err = proto.Unmarshal(serializedSignedChaincodeResponseMessage, signedResponseMsg)
	if err != nil {
		return nil, nil, err
	}

	serializedChaincodeResponseMessage := signedResponseMsg.ChaincodeResponseMessage

	responseMsg := &protos.ChaincodeResponseMessage{}
	err = proto.Unmarshal(serializedChaincodeResponseMessage, responseMsg)
	if err != nil {
		return nil, nil, err
	}

	return signedResponseMsg, responseMsg, err
}
