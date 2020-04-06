/*
* Copyright 2019 Intel Corporation
*
* SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"flag"
	"fmt"

	"github.com/hyperledger/fabric/core/container/dockercontroller"
)

func main() {
	netId := flag.String("net-id", "dev", "peer->networkId as specified in core.yaml")
	peerId := flag.String("peer-id", "jdoe", "peer->Id as specified in core.yaml")
	ccName := flag.String("cc-name", "ecc", "name of CC")
	ccVersion := flag.String("cc-version", "0", "version of CC")

	flag.Parse()

	vm := &dockercontroller.DockerVM{NetworkID: *netId, PeerID: *peerId}
	// chaincode id consists of name and version, see https://github.com/hyperledger/fabric/blob/c491d69962966db1f0231496ae6cab457d8a247d/core/scc/scc.go#L24
	ccid := *ccName + ":" + *ccVersion
	name, _ := vm.GetVMNameForDocker(ccid)
	fmt.Println(name)
}
