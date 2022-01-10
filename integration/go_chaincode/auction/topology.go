/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package auction

import (
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/api"
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/fabric"
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/fsc"
	"github.com/hyperledger/fabric-private-chaincode/integration/go_chaincode/auction/views/auctioneer"
	"github.com/hyperledger/fabric-private-chaincode/integration/go_chaincode/auction/views/bidder"
)

const ChaincodeName = "auction"

const ChaincodeImageName = "fpc/fpc-auction-go:main"

//const ChaincodeImageName = "fpc/auction"

func Topology() []api.Topology {
	fabricTopology := fabric.NewDefaultTopology()
	fabricTopology.AddOrganizationsByName("Org1", "Org2", "Org3")
	fabricTopology.AddFPC(ChaincodeName, ChaincodeImageName)
	fabricTopology.SetLogging("fpc=debug:grpc=error:comm.grpc=error:gossip=warning:info", "")
	fscTopology := fsc.NewTopology()

	// alice (auctioneer)
	aliceNode := fscTopology.AddNodeByName("alice")
	aliceNode.AddOptions(fabric.WithOrganization("Org1"))
	aliceNode.RegisterViewFactory("init", &auctioneer.InitViewFactory{})
	aliceNode.RegisterViewFactory("create", &auctioneer.CreateViewFactory{})
	aliceNode.RegisterViewFactory("close", &auctioneer.CloseViewFactory{})

	// bob (bidder)
	bobNode := fscTopology.AddNodeByName("bob")
	bobNode.AddOptions(fabric.WithOrganization("Org2"))
	bobNode.RegisterViewFactory("submit", &bidder.SubmitViewFactory{})

	// charly (bidder)
	charlyNode := fscTopology.AddNodeByName("charly")
	charlyNode.AddOptions(fabric.WithOrganization("Org3"))
	charlyNode.RegisterViewFactory("submit", &bidder.SubmitViewFactory{})

	return []api.Topology{fabricTopology, fscTopology}
}
