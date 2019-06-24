/*
* Copyright Intel Corp. 2019 All Rights Reserved.
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

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
