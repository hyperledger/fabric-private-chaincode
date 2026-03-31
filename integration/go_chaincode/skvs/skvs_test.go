/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package kv

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

	// 1. Initialize Secret Keeper:
	// ./fpcclient invoke InitSecretKeeper
	_, err = ii.Client("alice").CallView("invoke", common.JSONMarshall(&Client{
		CID:      ChaincodeName,
		Function: "InitSecretKeeper",
		Args:     []string{},
	}))
	assert.NoError(t, err)

	// 2. Reveal the secret as Alice:
	// ./fpcclient query RevealSecret Alice
	_, err = ii.Client("alice").CallView("invoke", common.JSONMarshall(&Client{
		CID:      ChaincodeName,
		Function: "RevealSecret",
		Args:     []string{"Alice"},
	}))
	assert.NoError(t, err)

	// 3. Change the secret as Bob:
	// ./fpcclient invoke LockSecret Bob NewSecret
	_, err = ii.Client("alice").CallView("invoke", common.JSONMarshall(&Client{
		CID:      ChaincodeName,
		Function: "LockSecret",
		Args:     []string{"Bob", "NewSecret"},
	}))
	assert.NoError(t, err)

	// 4. Attempt to reveal the secret as Alice (now updated):
	// ./fpcclient query revealSecret Alice
	_, err = ii.Client("alice").CallView("invoke", common.JSONMarshall(&Client{
		CID:      ChaincodeName,
		Function: "RevealSecret",
		Args:     []string{"Alice"},
	}))
	assert.NoError(t, err)

	// 5. Remove Bob's access as Alice:
	// ./fpcclient invoke removeUser Alice Bob
	_, err = ii.Client("alice").CallView("invoke", common.JSONMarshall(&Client{
		CID:      ChaincodeName,
		Function: "RemoveUser",
		Args:     []string{"Alice", "Bob"},
	}))
	assert.NoError(t, err)

	// 6. Attempt to reveal the secret as Bob (should fail):
	// ./fpcclient query revealSecret Bob // (will failed)
	_, err = ii.Client("alice").CallView("invoke", common.JSONMarshall(&Client{
		CID:      ChaincodeName,
		Function: "RevealSecret",
		Args:     []string{"Bob"},
	}))
	assert.Error(t, err)

	// 7. Re-add Bob to the authorization list as Alice:
	// ./fpcclient invoke addUser Alice Bob
	_, err = ii.Client("alice").CallView("invoke", common.JSONMarshall(&Client{
		CID:      ChaincodeName,
		Function: "AddUser",
		Args:     []string{"Alice", "Bob"},
	}))
	assert.NoError(t, err)

	// 8. Bob can now reveal the secret successfully:
	// ./fpcclient query revealSecret Bob // (will success)
	_, err = ii.Client("alice").CallView("invoke", common.JSONMarshall(&Client{
		CID:      ChaincodeName,
		Function: "RevealSecret",
		Args:     []string{"Bob"},
	}))
	assert.NoError(t, err)

	/*
		TODO:
			instead of using Alice/Bob as a parameter during invoke/query
			we use the signer identity
		NEED to change secret-keeper.go implementation.

		Before:
			cd $FPC_PATH/samples/application/simple-cli-go
			make
			source $FPC_PATH/samples/chaincode/secret-keeper-go/details.env
			$FPC_PATH/samples/deployment/fabric-smart-client/the-simple-testing-network/env.sh Org1
			source Org1.env
		After:
			cd $FPC_PATH/samples/application/simple-cli-go
			make
			source $FPC_PATH/samples/chaincode/secret-keeper-go/details.env
			$FPC_PATH/samples/deployment/fabric-smart-client/the-simple-testing-network/env.sh Org1
			source Org1.env

			(Another Terminal)
			$FPC_PATH/samples/deployment/fabric-smart-client/the-simple-testing-network/env.sh Org2
			source Org2.env

		make ECC_MAIN_FILES=cmd/skvs/main.go with_go env docker
	*/
}
