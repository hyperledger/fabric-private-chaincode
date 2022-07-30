/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package simulation

import (
	"testing"

	"github.com/hyperledger-labs/fabric-smart-client/integration"
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/common"
	"github.com/stretchr/testify/assert"
)

func TestFlow(t *testing.T) {

	// setup fabric network
	ii, err := integration.Generate(23000, false, Topology()...)
	assert.NoError(t, err)
	ii.Start()
	defer ii.Stop()

	_, err = ii.Client("client").CallView("invoke", common.JSONMarshall(&Client{
		CID:      ChaincodeName,
		Function: "put_state",
		Args:     []string{"echo-0", "echo-0"},
	}))
	assert.NoError(t, err)

	_, err = ii.Client("client").CallView("invoke", common.JSONMarshall(&Client{
		CID:      ChaincodeName,
		Function: "get_state",
		Args:     []string{"echo-0"},
	}))
	assert.NoError(t, err)

	_, err = ii.Client("client").CallView("invoke", common.JSONMarshall(&Client{
		CID:      ChaincodeName,
		Function: "del_state",
		Args:     []string{"echo-0"},
	}))
	assert.NoError(t, err)
}
