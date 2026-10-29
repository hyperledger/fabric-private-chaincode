/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"github.com/hyperledger-labs/fabric-smart-client/integration"
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/cmd"
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/cmd/network"
	fpcnwo "github.com/hyperledger/fabric-private-chaincode/extension/fsc/integration/nwo/fabric/fpc"
)

func main() {
	// FSC no longer ships the FPC NWO extension; register it before the network is generated
	network.StartCMDPostNew = func(ii *integration.Infrastructure) error {
		ii.RegisterPlatformFactory(fpcnwo.NewPlatformFactory())
		return nil
	}

	m := cmd.NewMain("the-simple-testing-network", "0.1")
	mainCmd := m.Cmd()
	mainCmd.AddCommand(network.NewCmd(Fabric()...))
	m.Execute()
}
