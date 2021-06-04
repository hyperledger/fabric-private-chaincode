/*
   Copyright 2021 Intel Corporation

   SPDX-License-Identifier: Apache-2.0
*/

package eas

import (
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/golang/protobuf/proto"
	fpc "github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/gateway"
	testutils "github.com/hyperledger/fabric-private-chaincode/integration/client_sdk/go/utils"
	pb "github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/protos"
)

func unmarshalStatus(statusBytes []byte) (*pb.Status, error) {
	status := &pb.Status{}
	err := proto.Unmarshal(statusBytes, status)
	if err != nil {
		return nil, err
	}

	return status, nil
}

func NewExperiment(studyId string, experimentId string, workerCredentials *pb.WorkerCredentials) error {
	ccID := "experiment-approval-service"
	// get network
	network, err := testutils.SetupNetwork("mychannel")
	if err != nil {
		return err
	}

	// Get FPC Contract
	contract := fpc.GetContract(network, ccID)

	//build experiment proposal
	experimentProposal := pb.ExperimentProposal{
		StudyId:           studyId,
		ExperimentId:      experimentId,
		WorkerCredentials: workerCredentials,
	}

	experimentProposalBytes, err := proto.Marshal(&experimentProposal)
	if err != nil {
		return err
	}

	//encode to base64 (because the submit transaction API accepts strings)
	experimentProposalBytesB64 := base64.StdEncoding.EncodeToString(experimentProposalBytes)

	response, err := contract.SubmitTransaction("newExperiment", experimentProposalBytesB64)
	if err != nil {
		return err
	}
	fmt.Printf("New experiment with Id=%s proposed at FPC Experiment Approval Service!\n", experimentId)

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

func RequestEvaluationPack(experimentId string) (evaluationPack *pb.EncryptedEvaluationPack, e error) {
	ccID := "experiment-approval-service"
	// get network
	network, err := testutils.SetupNetwork("mychannel")
	if err != nil {
		return nil, err
	}

	// Get FPC Contract
	contract := fpc.GetContract(network, ccID)

	//build experiment proposal
	evaluationPackRequest := pb.EvaluationPackRequest{
		ExperimentId: experimentId,
	}

	evaluationPackRequesBytes, err := proto.Marshal(&evaluationPackRequest)
	if err != nil {
		return nil, err
	}

	//encode to base64 (because the submit transaction API accepts strings)
	evaluationPackRequesByteB64 := base64.StdEncoding.EncodeToString(evaluationPackRequesBytes)

	response, err := contract.SubmitTransaction("requestEvaluationPack", evaluationPackRequesByteB64)
	if err != nil {
		return nil, err
	}

	encryptedEvaluationPackBytes, err := base64.StdEncoding.DecodeString(string(response))
	if err != nil {
		return nil, err
	}

	encryptedEvaluationPack := &pb.EncryptedEvaluationPack{}
	err = proto.Unmarshal(encryptedEvaluationPackBytes, encryptedEvaluationPack)
	if err != nil || encryptedEvaluationPack.GetEncryptedEvaluationpack() == nil {
		//error decoding means something wrong with making the pack
		status, e := unmarshalStatus(encryptedEvaluationPackBytes)
		if e != nil {
			//cannot even unmarshal status, so just return the error
			return nil, err
		}

		//return error from status
		return nil, errors.New("Error getExperimentProposal: " + string(status.GetReturnCode()) + string(", ") + status.GetMsg())
	}

	fmt.Println("Received evaluation pack from FPC Experiment Approval Service!")
	return encryptedEvaluationPack, nil
}
