/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package simulation

import (
	"fmt"

	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/api"
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/fabric"
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/fabric/topology"
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/fsc"
)

const (
	ChaincodeName      = "kv-test-no-sgx"
	ChaincodeImageName = "fpc/kv-test-no-sgx"
	ChaincodeImageTag  = "latest"
)

func Topology() []api.Topology {
	chaincodeImageName := fmt.Sprintf("%s:%s", ChaincodeImageName, ChaincodeImageTag)
	var fpcOptions []func(chaincode *topology.ChannelChaincode)

	fabricTopology := fabric.NewDefaultTopology()
	fabricTopology.AddOrganizationsByName("Org1")
	fabricTopology.AddFPC(ChaincodeName, chaincodeImageName, fpcOptions...)
	fabricTopology.SetLogging("fpc=debug:grpc=error:comm.grpc=error:gossip=warning:info", "")
	fscTopology := fsc.NewTopology()

	// client
	clientNode := fscTopology.AddNodeByName("client")
	clientNode.AddOptions(fabric.WithOrganization("Org1"))
	clientNode.RegisterViewFactory("invoke", &ClientViewFactory{})

	return []api.Topology{fabricTopology, fscTopology}
}
