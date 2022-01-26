/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package kv

import (
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/api"
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/fabric"
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/fsc"
)

const (
	ChaincodeName      = "kv-test"
	ChaincodeImageName = "fpc/fpc-kv-test-go:main"
)

//const ChaincodeImageName = "fpc/auction"

func Topology() []api.Topology {
	fabricTopology := fabric.NewDefaultTopology()
	fabricTopology.AddOrganizationsByName("Org1")
	fabricTopology.AddFPC(ChaincodeName, ChaincodeImageName)
	fabricTopology.SetLogging("fpc=debug:grpc=error:comm.grpc=error:gossip=warning:info", "")
	fscTopology := fsc.NewTopology()

	// client
	clientNode := fscTopology.AddNodeByName("client")
	clientNode.AddOptions(fabric.WithOrganization("Org1"))
	clientNode.RegisterViewFactory("invoke", &ClientViewFactory{})

	return []api.Topology{fabricTopology, fscTopology}
}
