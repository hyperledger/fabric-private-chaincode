// Copyright Intel Corp. 2019 All Rights Reserved.
//
//  SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"fmt"

	"github.com/hyperledger/fabric/core/container/ccintf"
	"github.com/hyperledger/fabric/core/container/dockercontroller"
)

func main() {
	netId := flag.String("net-id", "dev", "peer->networkId as specified in core.yaml")
	peerId := flag.String("peer-id", "jdoe", "peer->Id as specified in core.yaml")
	ccName := flag.String("cc-name", "ecc", "name of CC")
	ccVersion := flag.String("cc-version", "0", "version of CC")

	flag.Parse()

	vm := &dockercontroller.DockerVM{NetworkID: *netId, PeerID: *peerId}
	ccid := ccintf.CCID{Name: *ccName, Version: *ccVersion}
	name, _ := vm.GetVMNameForDocker(ccid)
	fmt.Println(name)
}
