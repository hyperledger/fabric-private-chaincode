/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package kv

import (
	"fmt"
	"testing"
	"time"

	"github.com/hyperledger-labs/fabric-smart-client/integration"
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/common"
	"github.com/stretchr/testify/assert"
)

func TestFlow(t *testing.T) {

	// setup fabric network
	ii, err := integration.Generate(23000, false, Topology()...)
	assert.NoError(t, err)
	ii.Start()

	// give me some time
	fmt.Println("time to sleep!!")
	time.Sleep(15 * time.Second)

	defer ii.Stop()

	_, err = ii.Client("client").CallView("invoke", common.JSONMarshall(&Client{
		cid:      ChaincodeName,
		function: "put_state",
		args:     []string{"echo-0", "echo-0"},
	}))
	assert.NoError(t, err)

	_, err = ii.Client("client").CallView("invoke", common.JSONMarshall(&Client{
		cid:      ChaincodeName,
		function: "get_state",
		args:     []string{"echo-0"},
	}))
	assert.NoError(t, err)

	_, err = ii.Client("client").CallView("invoke", common.JSONMarshall(&Client{
		cid:      ChaincodeName,
		function: "del_state",
		args:     []string{"echo-0"},
	}))
	assert.NoError(t, err)
}
