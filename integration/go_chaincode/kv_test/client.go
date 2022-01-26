/*
   Copyright IBM Corp. All Rights Reserved.

   SPDX-License-Identifier: Apache-2.0
*/

package kv

import (
	"encoding/json"

	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric/services/fpc"
	"github.com/hyperledger-labs/fabric-smart-client/platform/view/view"
	"github.com/pkg/errors"
)

type Client struct {
	cid      string
	function string
	args     []string
}

type ClientView struct {
	*Client
}

func (c *ClientView) Call(context view.Context) (interface{}, error) {
	_, err := fpc.GetDefaultChannel(context).Chaincode(c.cid).Invoke(c.function, fpc.StringsToArgs(c.args)...).Call()
	if err != nil {
		return nil, errors.Wrapf(err, "error invoking %s", c.function)
	}

	return nil, nil
}

type ClientViewFactory struct{}

func (c *ClientViewFactory) NewView(in []byte) (view.View, error) {
	f := &ClientView{Client: &Client{}}
	if err := json.Unmarshal(in, f.Client); err != nil {
		return nil, err
	}
	return f, nil
}
