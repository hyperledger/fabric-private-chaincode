/*
   Copyright 2021 Intel Corporation

   SPDX-License-Identifier: Apache-2.0
*/

package data_provider

import (
	"encoding/base64"
	"errors"

	"github.com/golang/protobuf/proto"
	fpc "github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/gateway"
	testutils "github.com/hyperledger/fabric-private-chaincode/integration/client_sdk/go/utils"
	pb "github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/protos"
)

func RegisterData(studyId string, uuid []byte, publicKey []byte, decryptionKey []byte, dataHandler string) error {
	ccID := "experiment-approval-service"
	// get network
	network, err := testutils.SetupNetwork("mychannel")
	if err != nil {
		return err
	}

	// Get FPC Contract
	contract := fpc.GetContract(network, ccID)

	userIdentity := pb.Identity{
		Uuid:      string(uuid),
		PublicKey: publicKey,
	}

	//build request
	registerDataRequest := pb.RegisterDataRequest{
		Participant:   &userIdentity,
		DecryptionKey: decryptionKey,
		DataHandler:   dataHandler,
		StudyId:       studyId,
	}

	registerDataRequestBytes, err := proto.Marshal(&registerDataRequest)
	if err != nil {
		return err
	}

	//encode to base64 (because the submit transaction API accepts strings)
	registerDataRequestBytesB64 := base64.StdEncoding.EncodeToString(registerDataRequestBytes)

	response, err := contract.SubmitTransaction("registerData", registerDataRequestBytesB64)
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
