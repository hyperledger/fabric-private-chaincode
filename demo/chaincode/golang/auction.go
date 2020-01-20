/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package golang

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	cid "github.com/hyperledger/fabric/core/chaincode/shim/ext/cid"
	pb "github.com/hyperledger/fabric/protos/peer"
)

const auctionKeyPrefix = "auction_"
const auctionStatusKeyPostfix = "_status"
const bidKeyPrefix = "bid_"

func storeAuctionStatus(stub shim.ChaincodeStubInterface, auctionId int, status *AuctionStatus) error {
	statusJson, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("json error: %s", err)
	}

	key := fmt.Sprintf("%s%d%s", auctionKeyPrefix, auctionId, auctionStatusKeyPostfix)
	err = stub.PutState(key, statusJson)
	if err != nil {
		return fmt.Errorf("putState error: %s", err)
	}

	return nil
}

func fetchAuctionStatus(stub shim.ChaincodeStubInterface, auctionId int) (*AuctionStatus, error) {
	key := fmt.Sprintf("%s%d%s", auctionKeyPrefix, auctionId, auctionStatusKeyPostfix)
	auctionStatusJson, err := stub.GetState(key)
	if err != nil {
		return nil, fmt.Errorf("getState error: %s", err)
	}

	if auctionStatusJson == nil {
		return nil, fmt.Errorf("auction status with id %d does not exist", auctionId)
	}

	var auctionStatus AuctionStatus
	err = json.Unmarshal(auctionStatusJson, &auctionStatus)
	if err != nil {
		return nil, fmt.Errorf("json error: %s", err)
	}

	return &auctionStatus, nil
}

func fetchAuctionDetails(stub shim.ChaincodeStubInterface, auctionId int) (*Auction, error) {
	key := fmt.Sprintf("%s%d", auctionKeyPrefix, auctionId)
	auctionJson, err := stub.GetState(key)
	if err != nil {
		return nil, fmt.Errorf("getState error: %s", err)
	}

	if auctionJson == nil {
		return nil, fmt.Errorf("auction details with id %d does not exist", auctionId)
	}

	var auction Auction
	err = json.Unmarshal(auctionJson, &auction)
	return &auction, nil
}

func createAuction(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// hardcoded auction id
	request := &AuctionRequest{1}

	// static state from input
	var auction Auction
	err := json.Unmarshal([]byte(args[0]), &auction)
	if err != nil {
		return shim.Error("json error: " + err.Error())
	}

	// assign owner
	clientIdentity, err := cid.New(stub)
	if err != nil {
		return shim.Error("clientidentity constructor error: " + err.Error())
	}

	ownerMSPId, err := clientIdentity.GetMSPID()
	if err != nil {
		return shim.Error("GetMSPID error: " + err.Error())
	}
	// Note: as we require x509, we use this instead of GetID()
	// which would be more generic but creates a more complicated
	// non-standard encoding (base64-encoding of concatation
	// of 'x509::' and DN ..
	ownerCert, err := clientIdentity.GetX509Certificate()
	if err != nil {
		return shim.Error("GetX509Certificate error: " + err.Error())
	}
	ownerDN := ownerCert.Subject.String()

	auction.Owner = &Principal{ownerMSPId, ownerDN}
	logger.Info(fmt.Sprintf("CreateAuction: new owner mspid='%v', dn='%v')\n", ownerMSPId, ownerDN))

	auctionJson, err := json.Marshal(auction)
	if err != nil {
		return shim.Error("json error: " + err.Error())
	}

	key := fmt.Sprintf("%s%d", auctionKeyPrefix, request.AuctionId)
	err = stub.PutState(key, auctionJson)
	if err != nil {
		return shim.Error("PutState error: " + err.Error())
	}

	// dynamic state
	auctionStatus := &AuctionStatus{"clock", 1, true}
	err = storeAuctionStatus(stub, request.AuctionId, auctionStatus)
	if err != nil {
		return shim.Error(err.Error())
	}

	status := &Status{0, "OK"}

	return shim.Success(buildReturnJson(status, request))
}

func getAuctionDetails(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var request AuctionRequest
	err := json.Unmarshal([]byte(args[0]), &request)
	if err != nil {
		return shim.Error("json error: " + err.Error())
	}

	auction, err := fetchAuctionDetails(stub, request.AuctionId)
	if err != nil {
		status := &Status{-1, err.Error()}
		return shim.Success(buildReturnJson(status, nil))
	}

	status := &Status{0, "OK"}
	return shim.Success(buildReturnJson(status, auction))
}

func getAuctionStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var request AuctionRequest
	err := json.Unmarshal([]byte(args[0]), &request)
	if err != nil {
		return shim.Error("json error: " + err.Error())
	}

	auctionStatus, err := fetchAuctionStatus(stub, request.AuctionId)
	if err != nil {
		status := &Status{-1, err.Error()}
		return shim.Success(buildReturnJson(status, nil))
	}

	status := &Status{0, "OK"}
	return shim.Success(buildReturnJson(status, auctionStatus))
}

func submitClockBid(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	bidId := "1"
	err := stub.PutState(bidKeyPrefix+bidId, []byte(args[0]))
	if err != nil {
		return shim.Error("PutState error: " + err.Error())
	}

	status := &Status{0, "OK"}
	return shim.Success(buildReturnJson(status, nil))
}

func endRound(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var request AuctionRequest
	err := json.Unmarshal([]byte(args[0]), &request)
	if err != nil {
		return shim.Error("json error: " + err.Error())
	}

	auctionStatus, err := fetchAuctionStatus(stub, request.AuctionId)
	if err != nil {
		status := &Status{-1, err.Error()}
		return shim.Success(buildReturnJson(status, nil))
	}

	auctionStatus.RoundActive = false

	err = storeAuctionStatus(stub, request.AuctionId, auctionStatus)
	if err != nil {
		return shim.Error(err.Error())
	}

	status := &Status{0, "OK"}
	return shim.Success(buildReturnJson(status, nil))
}

func startNextRound(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var request AuctionRequest
	err := json.Unmarshal([]byte(args[0]), &request)
	if err != nil {
		return shim.Error("json error: " + err.Error())
	}

	auctionStatus, err := fetchAuctionStatus(stub, request.AuctionId)
	if err != nil {
		status := &Status{-1, err.Error()}
		// TODO discuss if this is a shim error os success?
		return shim.Success(buildReturnJson(status, nil))
	}

	auctionStatus.ClockRound = auctionStatus.ClockRound + 1
	auctionStatus.RoundActive = true

	err = storeAuctionStatus(stub, request.AuctionId, auctionStatus)
	if err != nil {
		return shim.Error(err.Error())
	}

	status := &Status{0, "OK"}
	return shim.Success(buildReturnJson(status, nil))
}

func getRoundInfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var request RoundRequest
	err := json.Unmarshal([]byte(args[0]), &request)
	if err != nil {
		return shim.Error("json error: " + err.Error())
	}

	auction, err := fetchAuctionDetails(stub, request.AuctionId)
	if err != nil {
		status := &Status{-1, err.Error()}
		return shim.Success(buildReturnJson(status, nil))
	}

	// gather information for round
	auctionStatus, err := fetchAuctionStatus(stub, request.AuctionId)
	if err != nil {
		status := &Status{-1, err.Error()}
		// TODO discuss if this is a shim error os success?
		return shim.Success(buildReturnJson(status, nil))
	}

	var prices []Price
	for _, territory := range auction.Territories {
		prices = append(prices, Price{
			TerritoryId: territory.Id,
			MinPrice:    territory.MinPrice,
			ClockPrice:  int(float64(territory.MinPrice) * (1 + float64(auction.ClockPriceIncrementPercentage)/100.0)),
		})
	}

	response := &RoundInfo{
		Prices: prices,
		Active: auctionStatus.RoundActive,
	}

	status := &Status{0, "OK"}
	return shim.Success(buildReturnJson(status, response))
}

func getBidderRoundResults(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var request RoundRequest
	err := json.Unmarshal([]byte(args[0]), &request)
	if err != nil {
		return shim.Error("json error: " + err.Error())
	}

	results := []Result{{
		TerritoryId:       1,
		PostedPrice:       50,
		ExcessDemand:      1000,
		ProcessedLicenses: 40,
	}}

	response := &BidderRoundResult{
		Result:                 results,
		FutureEligibility:      31,
		RequiredFutureActivity: 5187,
	}

	status := &Status{0, "OK"}
	return shim.Success(buildReturnJson(status, response))
}
