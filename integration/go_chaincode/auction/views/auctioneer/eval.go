/*
   Copyright IBM Corp. All Rights Reserved.

   SPDX-License-Identifier: Apache-2.0
*/

package auctioneer

import (
	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric/services/fpc"
	"github.com/hyperledger-labs/fabric-smart-client/platform/view/view"
	"github.com/pkg/errors"
)

type EvalView struct {
	*Auction
}

func (c *EvalView) Call(context view.Context) (interface{}, error) {
	// chaincode details
	cid := ChaincodeName
	f := "eval"
	arg := c.Name

	response, err := fpc.GetDefaultChannel(context).Chaincode(cid).Invoke(f, arg).Call()
	if err != nil {
		return nil, errors.Wrapf(err, "error invoking %s", f)
	}

	return response, nil
}

func NewEvalView(auction *Auction) view.View {
	return &EvalView{
		Auction: auction,
	}
}
