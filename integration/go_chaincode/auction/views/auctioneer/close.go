/*
   Copyright IBM Corp. All Rights Reserved.

   SPDX-License-Identifier: Apache-2.0
*/

package auctioneer

import (
	"encoding/json"

	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric/services/fpc"
	"github.com/hyperledger-labs/fabric-smart-client/platform/view/view"
	"github.com/pkg/errors"
)

type CloseView struct {
	*Auction
}

func (c *CloseView) Call(context view.Context) (interface{}, error) {
	// chaincode details
	cid := ChaincodeName
	f := "close"
	arg := c.Name

	_, err := fpc.GetDefaultChannel(context).Chaincode(cid).Invoke(f, arg).Call()
	if err != nil {
		return nil, errors.Wrapf(err, "error invoking %s", f)
	}

	return context.RunView(NewEvalView(c.Auction))
}

type CloseViewFactory struct{}

func (c *CloseViewFactory) NewView(in []byte) (view.View, error) {
	f := &CloseView{Auction: &Auction{}}
	if err := json.Unmarshal(in, f.Auction); err != nil {
		return nil, err
	}
	return f, nil
}
