/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package utils

import (
	"encoding/hex"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer/lifecycle"
)

func GetMrEnclave(chaincodeId string, stub shim.ChaincodeStubInterface) (string, error) {
	ccDef, err := GetChaincodeDefinition(chaincodeId, stub)
	if err != nil {
		return "", err
	}

	mrenclave, err := ExtractMrEnclaveFromChaincodeDefinition(ccDef)
	if err != nil {
		return "", err
	}

	return mrenclave, nil
}

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

	if resp.Payload == nil {
		// no chaincode definition found
		return nil, fmt.Errorf("no chaincode definition found for chaincode='%s'", chaincodeId)
	}

	df := &lifecycle.QueryApprovedChaincodeDefinitionResult{}
	if err := proto.Unmarshal(resp.Payload, df); err != nil {
		return nil, err
	}
	return df, nil
}

func ExtractMrEnclaveFromChaincodeDefinition(ccDef *lifecycle.QueryApprovedChaincodeDefinitionResult) (string, error) {
	if ccDef == nil {
		return "", fmt.Errorf("chaincode definition input is nil")
	}

	if err := isValidMrEnclave(ccDef.Version); err != nil {
		return "", err
	}

	return ccDef.Version, nil
}

// checks that mrenclave is encoded as hex string and has correct length
func isValidMrEnclave(input string) error {
	expectedLength := 32

	mrenclave, err := hex.DecodeString(input)
	if err != nil {
		return err
	}

	if len(mrenclave) != expectedLength {
		return fmt.Errorf("mrenclave has wrong length! expteced %d but got %d", expectedLength, len(mrenclave))
	}

	return nil
}
