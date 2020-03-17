/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package golang

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger-labs/fabric-private-chaincode/utils"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/flogging"
)

var logger = flogging.MustGetLogger("MockAuction")

type MockAuction struct {
	dispatcher map[string]func(shim.ChaincodeStubInterface, []string) pb.Response
}

// Init initializes the chaincode
func (t *MockAuction) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (t *MockAuction) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	logger.Info(fmt.Sprintf("invoke: `%s` with args: %s\n", function, args))

	action, exists := t.dispatcher[function]
	if !exists {
		return shim.Error("Invalid invoke function name")
	}

	response := action(stub, args)
	logger.Info(fmt.Sprintf("Response: %s\n", response.Payload))

	if response.Status != shim.OK {
		// if we have an error return
		return response
	}

	// otherwise, wrap response in a ecc compatible response
	responseMsg := &utils.Response{
		ResponseData: response.Payload,
		Signature:    []byte("fake-signature"),
		PublicKey:    []byte("fake-pk"),
	}
	responseBytes, _ := json.Marshal(responseMsg)
	return shim.Success(responseBytes)
}

func NewMockAuction() shim.Chaincode {
	actions := map[string]func(shim.ChaincodeStubInterface, []string) pb.Response{
		"createAuction":         createAuction,
		"getAuctionDetails":     getAuctionDetails,
		"getAuctionStatus":      getAuctionStatus,
		"submitClockBid":        submitClockBid,
		"endRound":              endRound,
		"startNextRound":        startNextRound,
		"getRoundInfo":          getRoundInfo,
		"getBidderRoundResults": getBidderRoundResults,
	}
	return &MockAuction{dispatcher: actions}
}
