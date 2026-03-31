/*
   Copyright IBM Corp. All Rights Reserved.

   SPDX-License-Identifier: Apache-2.0
*/

package kv

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric/services/fpc"
	"github.com/hyperledger-labs/fabric-smart-client/platform/view/view"
	"github.com/pkg/errors"
)

type Client struct {
	CID      string
	Function string
	Args     []string
}

type ClientView struct {
	*Client
}

func (c *ClientView) Call(context view.Context) (interface{}, error) {
	fmt.Printf("Call FPC (CID='%s') with f='%s' and Args='%v'\n", c.CID, c.Function, c.Args)
	_, err := fpc.GetDefaultChannel(context).Chaincode(c.CID).Invoke(c.Function, fpc.StringsToArgs(c.Args)...).Call()
	if err != nil {
		return nil, errors.Wrapf(err, "error invoking %s", c.Function)
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
