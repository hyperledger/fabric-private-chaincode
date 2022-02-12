/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

const OK = "OK"
const AUCTION_DRAW = "DRAW"
const AUCTION_NO_BIDS = "NO_BIDS"
const AUCTION_ALREADY_EXISTING = "AUCTION_ALREADY_EXISTING"
const AUCTION_NOT_EXISTING = "AUCTION_NOT_EXISTING"
const AUCTION_ALREADY_CLOSED = "AUCTION_ALREADY_CLOSED"
const AUCTION_STILL_OPEN = "AUCTION_STILL_OPEN"

const INITIALIZED_KEY = "initialized"
const AUCTION_HOUSE_NAME_KEY = "auction_house_name"

const PREFIX = "somePrefix"

type Auction struct {
}

type auctionType struct {
	Name   string
	IsOpen bool
}

type bidType struct {
	BidderName string
	Value      int
}

func (t *Auction) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (t *Auction) Invoke(stub shim.ChaincodeStubInterface) pb.Response {

	initialized := true
	var auctionHouseName string

	init, err := stub.GetState(INITIALIZED_KEY)
	if err != nil || bytes.Equal(init, []byte{0x0}) {
		initialized = false
		auctionHouseName = "(uninitialized)"
	} else {
		ahn, err := stub.GetState(AUCTION_HOUSE_NAME_KEY)
		if err != nil {
			auctionHouseName = "(uninitialized)"
		} else {
			auctionHouseName = string(ahn)
		}
	}

	fmt.Println("AuctionCC: +++ Executing", auctionHouseName, "auction chaincode invocation +++")
	functionName, params := stub.GetFunctionAndParameters()
	fmt.Println("AuctionCC: Function:", functionName, "Params:", params)

	auctionName := params[0]
	var result string

	if !initialized && functionName != "init" {
		return shim.Error("AuctionCC: Auction not yet initialized / No re-initialized allowed")
	}

	switch functionName {
	case "init":
		result = t.initAuctionHouse(stub, params[0])
	case "create":
		result = t.auctionCreate(stub, auctionName)
	case "submit":
		value, _ := strconv.Atoi(params[2])
		result = t.auctionSubmit(stub, auctionName, params[1], value)
	case "close":
		result = t.auctionClose(stub, auctionName)
	case "eval":
		result = t.auctionEval(stub, auctionName)
	default:
		return shim.Error("AuctionCC: RECEIVED UNKNOWN transaction")
	}

	fmt.Println("AuctionCC: Response:", result)
	fmt.Println("AuctionCC: +++ Executing done +++")
	return shim.Success([]byte(result))

}

func (t *Auction) initAuctionHouse(stub shim.ChaincodeStubInterface, auctionHouseName string) string {
	stub.PutState(AUCTION_HOUSE_NAME_KEY, []byte(auctionHouseName))
	stub.PutState(INITIALIZED_KEY, []byte{0x1})
	return "OK"
}

func (t *Auction) auctionCreate(stub shim.ChaincodeStubInterface, auctionName string) string {
	value, err := stub.GetState(auctionName)
	if value != nil && err == nil {
		fmt.Println("AuctionCC: Auction already exists")
		return AUCTION_ALREADY_EXISTING
	}
	auction := &auctionType{Name: auctionName, IsOpen: true}
	auctionBytes, _ := json.Marshal(auction)
	stub.PutState(auctionName, auctionBytes)
	return "OK"
}

func (t *Auction) auctionSubmit(stub shim.ChaincodeStubInterface, auctionName string, bidderName string, value int) string {
	auctionBytes, err := stub.GetState(auctionName)
	if err != nil {
		fmt.Println("AuctionCC: Auction does not exist")
		return AUCTION_NOT_EXISTING
	}

	var auction auctionType
	json.Unmarshal(auctionBytes, &auction)

	if !auction.IsOpen {
		fmt.Println("AuctionCC: Auction is already closed")
		return AUCTION_ALREADY_CLOSED
	}

	key, _ := stub.CreateCompositeKey(PREFIX, []string{auctionName, bidderName})
	bid := &bidType{BidderName: bidderName, Value: value}

	bidBytes, _ := json.Marshal(bid)
	stub.PutState(key, bidBytes)
	return "OK"
}

func (t *Auction) auctionClose(stub shim.ChaincodeStubInterface, auctionName string) string {
	auctionBytes, err := stub.GetState(auctionName)
	if err != nil {
		fmt.Println("AuctionCC: Auction does not exist")
		return AUCTION_NOT_EXISTING
	}

	var auction auctionType
	json.Unmarshal(auctionBytes, &auction)

	if !auction.IsOpen {
		fmt.Println("AuctionCC: Auction is already closed")
		return AUCTION_ALREADY_CLOSED
	}

	auction.IsOpen = false

	auctionBytes, _ = json.Marshal(auction)
	stub.PutState(auctionName, auctionBytes)
	return "OK"
}

func (t *Auction) auctionEval(stub shim.ChaincodeStubInterface, auctionName string) string {
	auctionBytes, err := stub.GetState(auctionName)
	if err != nil {
		fmt.Println("AuctionCC: Auction does not exist")
		return AUCTION_NOT_EXISTING
	}

	var auction auctionType
	json.Unmarshal(auctionBytes, &auction)

	if auction.IsOpen {
		fmt.Println("AuctionCC: Auction is still open")
		return AUCTION_STILL_OPEN
	}

	var result string
	iter, _ := stub.GetStateByPartialCompositeKey(PREFIX, []string{auctionName})

	if !iter.HasNext() {
		fmt.Println("AuctionCC: No bids")
		result = AUCTION_NO_BIDS
	} else {
		var winner bidType
		high := -1
		draw := false

		fmt.Println("AuctionCC: All considered bids:")
		for iter.HasNext() {
			var bid bidType
			bidBytes, _ := iter.Next()
			json.Unmarshal(bidBytes.Value, &bid)
			fmt.Println("AuctionCC:", bid.BidderName, "value", bid.Value)

			if bid.Value > high {
				draw = false
				high = bid.Value
				winner = bid
			} else if bid.Value == high {
				draw = true
			}
		}

		if !draw {
			fmt.Println("AuctionCC: Winner is:", winner.BidderName, "with", winner.Value)
			winnerBytes, _ := json.Marshal(winner)
			result = string(winnerBytes)
		} else {
			fmt.Println("AuctionCC: DRAW")
			result = AUCTION_DRAW
		}
	}

	resultKey, _ := stub.CreateCompositeKey("Outcome", []string{auctionName})
	stub.PutState(resultKey, []byte(result))

	return result
}
