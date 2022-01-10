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

type InitView struct {
}

func (c *InitView) Call(context view.Context) (interface{}, error) {
	// chaincode details
	cid := ChaincodeName
	f := "init"
	arg := "LittleAuctionHouse"

	_, err := fpc.GetDefaultChannel(context).Chaincode(cid).Invoke(f, arg).Call()
	if err != nil {
		return nil, errors.Wrapf(err, "error invoking %s", f)
	}

	return nil, nil
}

type InitViewFactory struct{}

func (c *InitViewFactory) NewView(in []byte) (view.View, error) {
	f := &InitView{}
	return f, nil
}
