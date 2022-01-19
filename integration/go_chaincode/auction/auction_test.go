/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package auction

import (
	"fmt"
	"testing"

	"github.com/hyperledger-labs/fabric-smart-client/integration"
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/common"
	"github.com/hyperledger/fabric-private-chaincode/integration/go_chaincode/auction/views/auctioneer"
	"github.com/hyperledger/fabric-private-chaincode/integration/go_chaincode/auction/views/bidder"
	"github.com/stretchr/testify/assert"
)

const (
	auctionID = "pineapple"
)

func TestFlow(t *testing.T) {

	// setup fabric network
	ii, err := integration.Generate(23000, false, Topology()...)
	assert.NoError(t, err)
	ii.Start()

	// give me some time
	//fmt.Println("time to sleep!!")
	//time.Sleep(45 * time.Second)

	defer ii.Stop()

	// init auction house
	_, err = ii.Client("alice").CallView("init", nil)
	assert.NoError(t, err)

	// create auction
	_, err = ii.Client("alice").CallView("create", common.JSONMarshall(&auctioneer.Auction{
		Name: auctionID,
	}))
	assert.NoError(t, err)

	// bob submits
	_, err = ii.Client("bob").CallView("submit", common.JSONMarshall(&bidder.Bid{
		AuctionName: auctionID,
		BidderName:  "bob",
		Value:       100,
	}))
	assert.NoError(t, err)

	// charly submits
	_, err = ii.Client("charly").CallView("submit", common.JSONMarshall(&bidder.Bid{
		AuctionName: auctionID,
		BidderName:  "charly",
		Value:       140,
	}))
	assert.NoError(t, err)

	// close and eval auction
	response, err := ii.Client("alice").CallView("close", common.JSONMarshall(&auctioneer.Auction{
		Name: auctionID,
	}))
	assert.NoError(t, err)

	fmt.Printf("Winner: %s\n", response)
}
