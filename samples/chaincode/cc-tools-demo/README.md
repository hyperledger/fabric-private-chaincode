# CC-Tools-Demo Tutorial

This tutorial shows how to build, install and test a Go Chaincode developed using the [CC-Tools]() framework and integrating it with the Fabric Private Chaincode (FPC) framework.

This tutorial illustrates a simple use case where we follow the [cc-tools-demo]() chaincode which is based on standard fabric and then convert it to an FPC chaincode achieving FPC security capabilities.

This tutorial is based on the [FPC with CC-Tools integration project]() and all our design choices are explained here in the [design document]()
Here are the steps to accomplish this:

* Clone and copy the cc-tools-demo chaincode
* Edit the chaincode to became an FPC chaincode instead of normal fabric
* Build your FPC Go Chaincode
* Launch a Fabric network
* Install and instantiate your chaincode
* Invoke transactions by using the FPC simple-cli

## Prerequisites

This tutorial presumes that you have installed FPC on your `$GOPATH` as described in the FPC [README.md](../../../README.md#clone-fabric-private-chaincode) and `$FPC_PATH` is set accordingly.

We also need a working FPC development environment. As described in the "Setup your Development Environment" Section of the FPC [README.md](../../../README.md#setup-your-development-environment), you can use our docker-based dev environment (Option 1) or setup your local development environment (Option 2).
We recommend using the docker-based development environment and continue this tutorial within the dev container terminal.

Moreover, within your FPC development you have already installed the FPC Go Chaincode Support components.
See the installation steps in [ecc_go/README.md](../../../ecc_go/README.md#installation).

We also assume that you are familiar with Fabric chaincode development in go.
Most of the steps in this tutorial follow the normal Fabric chaincode development process, however, there are a few differences that we will highlight here.

## Clone and copy the cc-tools-demo chaincode

Clone the [cc-tools-demo]() repository and copy the [chaincode]() folder. Then paste it in the root directory for cc-tools-demo here.

```bash
cd ~
git clone https://github.com/hyperledger-labs/cc-tools-demo.git
cp ~/cc-tools-demo/chaincode $FPC_PATH/samples/chaincode/cc-tools-demo
cd $FPC_PATH/samples/chaincode/cc-tools-demo
```

The chaincode code structure is different than normal chaincode as it's using the cc-tools-framework

## Edit the chaincode to became an FPC chaincode instead of normal fabric

Go to `$FPC_PATH/samples/chaincode/cc-tools-demo/main.go` and create the project structure.

```bash
cd $FPC_PATH/samples/chaincode/cc-tools-demo
```

The code presented here is not different from traditional Go Chaincode developed with CC-Tools framework. All the FPC specific protect mechanisms are handled by the FPC framework transparently.

The `main.go` contains the starting point of the chaincode.

To use FPC, we need to add some logic to instantiate our private chaincode and start it.
To do so, we use `fpc.NewPrivateChaincode(&chaincode.AssetExample{})`.
Add the following code to `$FPC_PATH/samples/chaincode/cc-tools-demo/main.go` in functions `main()` and `runCCaaS()`:

```go
package main

import (
	fpc "github.com/hyperledger/fabric-private-chaincode/ecc_go/chaincode"
)

func main() {

	//*CC-Tools Specific code (DO NOT OVERWRITE)*//

	if os.Getenv("RUN_CCAAS") == "true" {
		err = runCCaaS()
	} else {
		if os.Getenv("FPC_ENABLED") == "true" {
			err = shim.Start(fpc.NewPrivateChaincode(new(CCDemo)))
		} else {
			err = shim.Start(new(CCDemo))
		}
	}

	//*CC-Tools Specific code (DO NOT OVERWRITE)*//

}

func runCCaaS() error {
	address := os.Getenv("CHAINCODE_SERVER_ADDRESS")
	ccid := os.Getenv("CHAINCODE_PKG_ID")

	//*CC-Tools Specific code (DO NOT OVERWRITE)*//

	var cc shim.Chaincode
	if os.Getenv("FPC_ENABLED") == "true" {
		cc = fpc.NewPrivateChaincode(new(CCDemo))
	} else {
		cc = new(CCDemo)
	}

	//*CC-Tools Specific code (DO NOT OVERWRITE)*//

}

```

## Building FPC Go Chaincode

Create a `Makefile` (i.e., `touch $FPC_PATH/samples/chaincode/cc-tools-demo/Makefile`) with the following content:

```Makefile
TOP = ../../..
include $(TOP)/ecc_go/build.mk

CC_NAME ?= fpc-cc-tools-demo
```

Please make sure that in the file above the variable `TOP` points to the FPC root directory (i.e., `$FPC_PATH`).

In `$FPC_PATH/samples/chaincode/cc-tools-demo` directory, to build the chaincode and package it as docker image, execute:

```bash
make
```

After building, you can check that the `fpc/fpc-cc-tools-demo` image exists in your local docker registry using:

```bash
docker images | grep fpc-cc-tools-demo
```

## Time to test!

Next step is to test the chaincode by invoking transactions, for which you need a basic Fabric network with a channel.
We will use the test network provided in [`$FPC_PATH/samples/deployment/test-network`](../../deployment/test-network).
To invoke the chaincode, we will use the `simple-cli` application in [`$FPC_PATH/samples/application/simple-cli-go`](../../pplication/simple-cli-go).

### Enclave Registry

To run any FPC chaincode we need to prepare the docker images for the FPC Enclave Registry (ERCC).
In case you have not yet created them, run `make -C $FPC_PATH/ercc build docker`.

### Prepare the test network

We already provide a detailed tutorial how to use FPC with the test network in [`$FPC_PATH/samples/deployment/test-network`](../../deployment/test-network).
However, for completeness, let's go through the required steps once again.

```bash
cd $FPC_PATH/samples/deployment/test-network
./setup.sh
```

### Start the test network

Now we are ready to launch the fabric test network and install the FPC chaincode on it.
We begin with setting up the network with a single channel `mychannel`.

```bash
cd $FPC_PATH/samples/deployment/test-network/fabric-samples/test-network
./network.sh up -ca
./network.sh createChannel -c mychannel
```

Once the network is up and running, we install the simple asset chaincode and the FPC Enclave Registry.
We provide a small shell script to make this task a bit easier.

```bash
export CC_ID=cc-tools-demo
export CC_PATH="$FPC_PATH/samples/chaincode/cc-tools-demo/"
export CC_VER=$(cat "$FPC_PATH/samples/chaincode/cc-tools-demo/mrenclave")
cd $FPC_PATH/samples/deployment/test-network
./installFPC.sh
```

Note that the `installFPC.sh` script returns an export statement you need to copy and paste in the terminal.
This sets environment variables for the package IDs for each chaincode container.
Continue by running:

```bash
make ercc-ecc-start
```

You should see now four containers running (i.e., `cc-tools-demo.peer0.org1`, `cc-tools-demo.peer0.org2`, `ercc.peer0.org1`, and `ercc.peer0.org2`).

### Invoke simple asset

Open a new terminal and connect to the `fpc-development-go-support` container by running

```bash
docker exec -it fpc-development-go-support /bin/bash
```

```bash
# prepare connections profile
cd $FPC_PATH/samples/deployment/test-network
./update-connection.sh

# # update the connection profile for external clients outside the fpc dev environment
cd $FPC_PATH/samples/deployment/test-network
./update-external-connection.sh

# make fpcclient
cd $FPC_PATH/samples/application/simple-cli-go
make

# export fpcclient settings
export CC_NAME=cc-tools-demo
export CHANNEL_NAME=mychannel
export CORE_PEER_ADDRESS=localhost:7051
export CORE_PEER_ID=peer0.org1.example.com
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_MSPCONFIGPATH=$FPC_PATH/samples/deployment/test-network/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_TLS_CERT_FILE=$FPC_PATH/samples/deployment/test-network/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/server.crt
export CORE_PEER_TLS_ENABLED="true"
export CORE_PEER_TLS_KEY_FILE=$FPC_PATH/samples/deployment/test-network/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/server.key
export CORE_PEER_TLS_ROOTCERT_FILE=$FPC_PATH/samples/deployment/test-network/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export ORDERER_CA=$FPC_PATH/samples/deployment/test-network/fabric-samples/test-network/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
export GATEWAY_CONFIG=$FPC_PATH/samples/deployment/test-network/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/connection-org1.yaml
export FPC_ENABLED=true
export RUN_CCAAS=true

# init our enclave
./fpcclient init $CORE_PEER_ID

# interact with the FPC-CC-Tools Chaincode. The getSchema is a built-in cc-tools transaction
./fpcclient invoke getSchema

```

Congratulations! You have successfully created a FPC chaincode with go and invoked it using our simple cli.
