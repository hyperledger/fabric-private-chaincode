/*
   Copyright IBM Corp. All Rights Reserved.

   SPDX-License-Identifier: Apache-2.0
*/

package bidder

import (
	"encoding/json"

	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric/services/fpc"
	"github.com/hyperledger-labs/fabric-smart-client/platform/view/view"
	"github.com/pkg/errors"
)

const ChaincodeName = "auction"

type Bid struct {
	AuctionName string
	BidderName  string
	Value       int
}

type SubmitView struct {
	*Bid
}

func (c *SubmitView) Call(context view.Context) (interface{}, error) {
	// chaincode details
	cid := ChaincodeName
	f := "submit"
	//arg := [...]string{c.AuctionName, c.BidderName, strconv.Itoa(c.Value)}

	_, err := fpc.GetDefaultChannel(context).Chaincode(cid).Invoke(f, c.AuctionName, c.BidderName, c.Value).Call()
	if err != nil {
		return nil, errors.Wrapf(err, "error invoking %s", f)
	}

	return nil, nil
}

type SubmitViewFactory struct{}

func (c *SubmitViewFactory) NewView(in []byte) (view.View, error) {
	f := &SubmitView{Bid: &Bid{}}
	if err := json.Unmarshal(in, f.Bid); err != nil {
		return nil, err
	}
	return f, nil
}
