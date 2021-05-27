/*
   Copyright 2021 Intel Corporation

   SPDX-License-Identifier: Apache-2.0
*/

package pi

import (
	"encoding/base64"
	"errors"

	"github.com/golang/protobuf/proto"
	fpc "github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/gateway"
	testutils "github.com/hyperledger/fabric-private-chaincode/integration/client_sdk/go/utils"
	pb "github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/protos"
)

func getContractHandler() (fpc.Contract, error) {
	ccID := "experiment-approval-service"
	// get network
	network, err := testutils.SetupNetwork("mychannel")
	if err != nil {
		return nil, err
	}

	// Get FPC Contract
	contract := fpc.GetContract(network, ccID)

	return contract, nil
}

func CreateIdentity(uuid []byte, publicKey []byte, publicEncryptionKey []byte) *pb.Identity {
	identity := &pb.Identity{}
	if uuid != nil {
		identity.Uuid = string(uuid)
	}
	if publicKey != nil {
		identity.PublicKey = publicKey
	}
	if publicEncryptionKey != nil {
		identity.PublicEncryptionKey = publicEncryptionKey
	}
	return identity
}

func unmarshalStatus(statusBytes []byte) (*pb.Status, error) {
	status := &pb.Status{}
	err := proto.Unmarshal(statusBytes, status)
	if err != nil {
		return nil, err
	}

	return status, nil
}

func RegisterStudy(studyId string, metadata string, userIdentities []*pb.Identity) error {
	contract, err := getContractHandler()
	if err != nil {
		return err
	}

	//build request
	studyDetailsMessage := pb.StudyDetailsMessage{
		StudyId:        studyId,
		Metadata:       metadata,
		UserIdentities: userIdentities,
	}

	studyDetailsMessageBytes, err := proto.Marshal(&studyDetailsMessage)
	if err != nil {
		return err
	}

	//encode to base64 (because the submit transaction API accepts strings)
	requestBytesB64 := base64.StdEncoding.EncodeToString(studyDetailsMessageBytes)

	response, err := contract.SubmitTransaction("registerStudy", requestBytesB64)
	if err != nil {
		return err
	}

	//response should be a base64 Status
	statusBytes, err := base64.StdEncoding.DecodeString(string(response))
	if err != nil {
		return err
	}

	status := pb.Status{}
	err = proto.Unmarshal(statusBytes, &status)
	if err != nil {
		return err
	}

	if status.GetMsg() != "" || status.GetReturnCode() != pb.Status_OK {
		return errors.New("Error RegisterData: " + string(status.GetReturnCode()) + string(", ") + status.GetMsg())
	}

	return nil
}

func GetExperimentProposal(experimentId string) (*pb.ExperimentProposal, error) {
	contract, err := getContractHandler()
	if err != nil {
		return nil, err
	}

	getExperimentRequest := pb.GetExperimentRequest{
		ExperimentId: experimentId,
	}

	getExperimentRequestByte, err := proto.Marshal(&getExperimentRequest)
	if err != nil {
		return nil, err
	}

	//encode to base64 (because the submit transaction API accepts strings)
	requestBytesB64 := base64.StdEncoding.EncodeToString(getExperimentRequestByte)

	response, err := contract.SubmitTransaction("getExperimentProposal", requestBytesB64)
	if err != nil {
		return nil, err
	}

	experimentProposalBytes, err := base64.StdEncoding.DecodeString(string(response))
	if err != nil {
		return nil, err
	}

	experimentProposal := &pb.ExperimentProposal{}
	err = proto.Unmarshal(experimentProposalBytes, experimentProposal)
	if err != nil {
		//error decoding means that maybe the experiment was not found
		//so, check if status was returned instead
		status, e := unmarshalStatus(experimentProposalBytes)
		if e != nil {
			//cannot even unmarshal status, so just return the error
			return nil, err
		}

		//return error from status
		return nil, errors.New("Error getExperimentProposal: " + string(status.GetReturnCode()) + string(", ") + status.GetMsg())
	}

	return experimentProposal, nil
}

func DecideOnExperiment(experimentId string, experimentBytes []byte, decision string) error {
	if decision != "approved" && decision != "rejected" {
		return errors.New("wrong decision")
	}

	contract, err := getContractHandler()
	if err != nil {
		return err
	}

	approvalDecision := pb.Approval_UNDEFINED
	if decision == "approved" {
		approvalDecision = pb.Approval_APPROVED
	}
	if decision == "reject" {
		approvalDecision = pb.Approval_REJECTED
	}

	approval := pb.Approval{
		ExperimentId:       experimentId,
		ExperimentProposal: experimentBytes,
		Decision:           approvalDecision,
	}

	approvalBytes, err := proto.Marshal(&approval)
	if err != nil {
		return err
	}

	signedApproval := pb.SignedApprovalMessage{
		Approval: approvalBytes,
		//Signature: signature,
	}

	signedApprovalBytes, err := proto.Marshal(&signedApproval)
	if err != nil {
		return err
	}

	//encode to base64 (because the submit transaction API accepts strings)
	requestBytesB64 := base64.StdEncoding.EncodeToString(signedApprovalBytes)

	response, err := contract.SubmitTransaction("approveExperiment", requestBytesB64)
	if err != nil {
		return err
	}

	//response should be a base64 Status
	statusBytes, err := base64.StdEncoding.DecodeString(string(response))
	if err != nil {
		return err
	}

	status := pb.Status{}
	err = proto.Unmarshal(statusBytes, &status)
	if err != nil {
		return err
	}

	if status.GetMsg() != "" || status.GetReturnCode() != pb.Status_OK {
		return errors.New("Error approveExperiment: " + string(status.GetReturnCode()) + string(", ") + status.GetMsg())
	}

	return nil

}
