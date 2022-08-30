/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package kv

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/api"
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/fabric"
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/fabric/topology"
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/fsc"
	"github.com/hyperledger/fabric-private-chaincode/integration/go_chaincode/utils"
)

const (
	ChaincodeName      = "kv-test"
	ChaincodeImageName = "fpc/fpc-kv-test-go"
	ChaincodeImageTag  = "latest"
)

func Topology() []api.Topology {
	chaincodeImageName := fmt.Sprintf("%s:%s", ChaincodeImageName, ChaincodeImageTag)
	var fpcOptions []func(chaincode *topology.ChannelChaincode)

	if strings.ToUpper(os.Getenv("SGX_MODE")) == "HW" {
		chaincodeImageName = fmt.Sprintf("%s-hw:%s", ChaincodeImageName, ChaincodeImageTag)
		fpcOptions = append(fpcOptions, topology.WithSGXMode("HW"))

		mrenclave, err := utils.ReadMrenclaveFromFile("mrenclave")
		if err != nil {
			panic(errors.Wrapf(err, "cannot get mrenclave"))
		}
		fpcOptions = append(fpcOptions, topology.WithMREnclave(mrenclave))

		sgxDevicePath, err := utils.DetectSgxDevicePath()
		if err != nil {
			panic(errors.Wrapf(err, "SGX HW mode set but now sgx device found"))
		}
		fpcOptions = append(fpcOptions, topology.WithSGXDevicesPaths(sgxDevicePath))
	}

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
