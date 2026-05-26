/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package echo

import (
	"github.com/hyperledger-labs/fabric-smart-client/integration"
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/api"
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/fabric"
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/fsc"
	api2 "github.com/hyperledger-labs/fabric-smart-client/pkg/node"
	fpctopo "github.com/hyperledger/fabric-private-chaincode/extension/fsc/integration/nwo/fabric/topology"
	"github.com/hyperledger/fabric-private-chaincode/integration/go_chaincode/echo/views"
)

func Topology(sdk api2.SDK, commType fsc.P2PCommunicationType, replicationOpts *integration.ReplicationOptions) []api.Topology {
	// Create an empty fabric topology
	fabricTopology := fabric.NewDefaultTopology()
	// Note that Idemix is currently not supported by FPC
	// fabricTopology.EnableIdemix()
	// Add two organizations
	fabricTopology.AddOrganizationsByName("Org1", "Org2")
	// Add a standard chaincode
	// fabricTopology.AddNamespaceWithUnanimity("mycc", "Org1", "Org2")
	// Add an FPC by passing chaincode's id and docker image
	// Note: do NOT use AddPostRunInvocation for FPC chaincodes — FSC's InvokeChaincode
	// sends plain-text args which FPC's private chaincode wrapper rejects as "invalid
	// invocation". The FPC extension's own PostRun already handles __initEnclave.
	fpctopo.AddFPC(fabricTopology, "echo", "fpc/fpc-echo-go")

	// Create an empty FSC topology
	fscTopology := fsc.NewTopology()
	fscTopology.P2PCommunicationType = commType
	//fscTopology.SetLogging("debug", "")

	// Alice
	fscTopology.AddNodeByName("alice").
		AddOptions(fabric.WithOrganization("Org2")).
		AddOptions(replicationOpts.For("alice")...).
		RegisterViewFactory("ListProvisionedEnclaves", &views.ListProvisionedEnclavesViewFactory{}).
		RegisterViewFactory("Echo", &views.EchoViewFactory{})

	// Bob
	fscTopology.AddNodeByName("bob").
		AddOptions(fabric.WithOrganization("Org2")).
		AddOptions(replicationOpts.For("bob")...).
		RegisterViewFactory("ListProvisionedEnclaves", &views.ListProvisionedEnclavesViewFactory{}).
		RegisterViewFactory("Echo", &views.EchoViewFactory{})

	// Add Fabric SDK to FSC Nodes
	fscTopology.AddSDK(sdk)

	return []api.Topology{fabricTopology, fscTopology}
}
