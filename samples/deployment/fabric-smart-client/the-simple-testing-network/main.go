/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/cmd"
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/cmd/network"
)

func main() {
	m := cmd.NewMain("the-simple-testing-network", "0.1")
	mainCmd := m.Cmd()
	mainCmd.AddCommand(network.NewCmd(Fabric()...))
	m.Execute()
}
