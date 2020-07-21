/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package utils

import (
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer/lifecycle"
)

func GetChaincodeDefinition(chaincodeId string, stub shim.ChaincodeStubInterface) (*lifecycle.QueryApprovedChaincodeDefinitionResult, error) {
	function := "QueryApprovedChaincodeDefinition"
	args := &lifecycle.QueryApprovedChaincodeDefinitionArgs{
		Name:     chaincodeId,
		Sequence: 0, // TODO get current sequence number
	}
	argsBytes, err := proto.Marshal(args)
	if err != nil {
		return nil, err
	}

	resp := stub.InvokeChaincode("_lifecycle", [][]byte{[]byte(function), argsBytes}, stub.GetChannelID())

	df := &lifecycle.QueryApprovedChaincodeDefinitionResult{}
	if err := proto.Unmarshal(resp.Payload, df); err != nil {
		return nil, err
	}
	return df, nil
}

func ExtractMrEnclaveFromChaincodeDefinition(chaincodeId string, stub shim.ChaincodeStubInterface) (string, error) {
	ccDef, err := GetChaincodeDefinition(chaincodeId, stub)
	if err != nil {
		return "", err
	}

	// note that mrenclave is hex-encoded
	return ccDef.Version, nil
}
