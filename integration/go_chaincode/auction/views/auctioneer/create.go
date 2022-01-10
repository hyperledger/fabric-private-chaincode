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

const ChaincodeName = "auction"

type Auction struct {
	Name string
}

type CreateView struct {
	*Auction
}

func (c *CreateView) Call(context view.Context) (interface{}, error) {
	// chaincode details
	cid := ChaincodeName
	f := "create"
	arg := c.Name

	_, err := fpc.GetDefaultChannel(context).Chaincode(cid).Invoke(f, arg).Call()
	if err != nil {
		return nil, errors.Wrapf(err, "error invoking %s", f)
	}

	return nil, nil
}

type CreateViewFactory struct{}

func (c *CreateViewFactory) NewView(in []byte) (view.View, error) {
	f := &CreateView{Auction: &Auction{}}
	if err := json.Unmarshal(in, f.Auction); err != nil {
		return nil, err
	}
	return f, nil
}
