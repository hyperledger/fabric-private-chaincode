/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package utils

import (
	"encoding/hex"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer/lifecycle"
	"github.com/hyperledger/fabric/protoutil"
)

const MrEnclaveLength = 32

func GetChaincodeDefinition(chaincodeId string, stub shim.ChaincodeStubInterface) (*lifecycle.QueryChaincodeDefinitionResult, error) {
	channelId := stub.GetChannelID()

	function := "QueryChaincodeDefinition"
	args := &lifecycle.QueryChaincodeDefinitionArgs{
		Name: chaincodeId,
	}
	// note that we use Fabric's Marshall as it still uses protobuf V1
	argsBytes, err := protoutil.Marshal(args)
	if err != nil {
		return nil, err
	}

	resp := stub.InvokeChaincode("_lifecycle", [][]byte{[]byte(function), argsBytes}, channelId)
	if resp.Status != shim.OK {
		return nil, fmt.Errorf("error while retrieving chaincode definition: [%d] %s", resp.Status, resp.Message)
	}

	if resp.Payload == nil {
		// no chaincode definition found
		return nil, fmt.Errorf("no chaincode definition found for chaincode='%s'", chaincodeId)
	}

	return UnmarshalQueryChaincodeDefinitionResult(resp.Payload)
}

func GetMrEnclave(chaincodeId string, stub shim.ChaincodeStubInterface) (string, error) {
	ccDef, err := GetChaincodeDefinition(chaincodeId, stub)
	if err != nil {
		return "", err
	}

	return ExtractMrEnclave(ccDef)
}

func ExtractMrEnclave(ccDef *lifecycle.QueryChaincodeDefinitionResult) (string, error) {
	mrenclave := ccDef.Version
	if err := isValidMrEnclaveString(mrenclave); err != nil {
		return "", err
	}

	return mrenclave, nil
}

// checks that mrenclave is encoded as hex string and has correct length
func isValidMrEnclaveString(input string) error {
	mrenclave, err := hex.DecodeString(input)
	if err != nil {
		return err
	}

	if len(mrenclave) != MrEnclaveLength {
		return fmt.Errorf("mrenclave has wrong length! expteced %d but got %d", MrEnclaveLength, len(mrenclave))
	}

	return nil
}
