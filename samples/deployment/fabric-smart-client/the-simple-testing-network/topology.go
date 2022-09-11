/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/api"
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/fabric"
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/fabric/topology"
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/fsc"
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/monitoring"
)

const (
	defaultChaincodeName      = "kv-test-go"
	defaultChaincodeImageName = "fpc/fpc-kv-test-go"
	defaultChaincodeImageTag  = "latest"
	defaultChaincodeMRENCLAVE = "fakeMRENCLAVE"
	defaultLoggingSpec        = "info"
)

func Fabric() []api.Topology {
	config := setup()

	fabricTopology := fabric.NewDefaultTopology()
	fabricTopology.AddOrganizationsByName("Org1", "Org2")
	fabricTopology.SetLogging(config.loggingSpec, "")

	// in this example we use the FPC kv-test-go chaincode
	// we just need to set the docker images
	fabricTopology.EnableFPC()
	fabricTopology.AddFPC(config.chaincodeName, config.chaincodeImage, config.fpcOptions...)

	// bring hyperledger explorer into the game
	// you can reach it http://localhost:8080 with admin:admin
	monitoringTopology := monitoring.NewTopology()
	monitoringTopology.EnableHyperledgerExplorer()

	return []api.Topology{fabricTopology, fsc.NewTopology(), monitoringTopology}
}

type config struct {
	loggingSpec    string
	chaincodeName  string
	chaincodeImage string
	fpcOptions     []func(chaincode *topology.ChannelChaincode)
}

// setup prepares a config helper struct, containing some additional configuration that can be injected via environment variables
func setup() *config {
	config := &config{}

	// export FABRIC_LOGGING_SPECS=info
	config.loggingSpec = os.Getenv("FABRIC_LOGGING_SPEC")
	if len(config.loggingSpec) == 0 {
		config.loggingSpec = defaultLoggingSpec
	}

	// export CC_NAME=kv-test-go
	config.chaincodeName = os.Getenv("CC_NAME")
	if len(config.chaincodeName) == 0 {
		config.chaincodeName = defaultChaincodeName
	}

	// export FPC_CHAINCODE_IMAGE=fpc/fpc-kv-test-go:latest
	config.chaincodeImage = os.Getenv("FPC_CHAINCODE_IMAGE")
	if len(config.chaincodeImage) == 0 {
		config.chaincodeImage = fmt.Sprintf("%s:%s", defaultChaincodeImageName, defaultChaincodeImageTag)
	}

	// get mrenclave
	mrenclave := os.Getenv("FPC_MRENCLAVE")
	if len(mrenclave) == 0 {
		mrenclave = defaultChaincodeMRENCLAVE
	}
	config.fpcOptions = append(config.fpcOptions, topology.WithMREnclave(mrenclave))

	// check if we are running in SGX HW mode
	// export SGX_MODE=SIM
	if strings.ToUpper(os.Getenv("SGX_MODE")) == "HW" {
		sgxDevicePath := DetectSgxDevicePath()
		config.fpcOptions = append(config.fpcOptions, topology.WithSGXDevicesPaths(sgxDevicePath))
	}

	return config
}

func DetectSgxDevicePath() []string {
	possiblePaths := []string{"/dev/isgx", "/dev/sgx/enclave"}
	for _, p := range possiblePaths {
		if _, err := os.Stat(p); err != nil {
			continue
		} else {
			// first found path returns
			return []string{p}
		}
	}

	panic("no sgx device path found")
}

func ReadMrenclaveFromFile(path string) string {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(fmt.Errorf("cannot read mrenclave from %s", path))
	}

	mrenclave := strings.TrimSpace(string(data))
	if len(mrenclave) == 0 {
		panic(fmt.Errorf("mrenclave file empty"))
	}

	return mrenclave
}
