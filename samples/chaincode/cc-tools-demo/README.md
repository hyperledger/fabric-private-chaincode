# CC-Tools-Demo Tutorial

This tutorial shows how to build, install and test a Go Chaincode developed using the [CC-Tools](https://github.com/hyperledger-labs/cc-tools) framework and integrating it with the Fabric Private Chaincode (FPC) framework.

This tutorial illustrates a simple use case where we follow the [cc-tools-demo](https://github.com/hyperledger-labs/cc-tools-demo) chaincode which is based on standard Fabric and then convert it to an FPC chaincode achieving FPC security capabilities.

This tutorial is based on the [FPC with CC-Tools integration project](https://lf-hyperledger.atlassian.net/wiki/spaces/INTERN/pages/21954957/Hyperledger+Fabric+CC-Tools+Support+for+Fabric+Private+Chaincode) and all our design choices are explained here in the [design document](https://github.com/hyperledger/fabric-private-chaincode/tree/main/docs/design/integrate-with-cc-tools).
Here are the steps to accomplish this:

* Clone the cc-tools-demo chaincode
* Modify the chaincode to use FPC
* Build your FPC CC-tools-demo chaincode
* Launch a Fabric network
* Install and instantiate your chaincode
* Invoke transactions by using the FPC simple-cli

## Prerequisites

* This tutorial presumes that you have installed FPC on your `$GOPATH` as described in the FPC [README.md](../../../README.md#clone-fabric-private-chaincode) and `$FPC_PATH` is set accordingly.

* We need a working FPC development environment. As described in the "Setup your Development Environment" Section of the FPC [README.md](../../../README.md#setup-your-development-environment), you can use our docker-based dev environment (Option 1) or setup your local development environment (Option 2).
We recommend using the docker-based development environment and continue this tutorial within the dev container terminal.

* Moreover, within your FPC development you have already installed the FPC Go Chaincode Support components.
See the installation steps in [ecc_go/README.md](../../../ecc_go/README.md#installation).

* We assume that you are familiar with Fabric chaincode development in go.
Most of the steps in this tutorial follow the normal Fabric chaincode development process, however, there are a few differences that we will highlight here.

* Also, since the tutorial is on the integration between cc-tools and FPC, we expect you to have a grasp knowledge of [cc-tools](https://github.com/hyperledger-labs/cc-tools) framework and that you've at least tried to run the [cc-tools-demo](https://github.com/hyperledger-labs/cc-tools-demo) once by yourself on a Fabric network

## Clone the cc-tools-demo chaincode

We need to clone the chaincode folder from the [cc-tools-demo](https://github.com/hyperledger-labs/cc-tools-demo) repository here.

```bash
export ccToolsDemoPath=$FPC_PATH/samples/chaincode/cc-tools-demo
git clone -n --no-checkout --depth=1 --filter=tree:0 https://github.com/hyperledger-labs/cc-tools-demo.git "$ccToolsDemoPath/chaincode"
cd "$ccToolsDemoPath/chaincode" || { echo "$ccToolsDemoPath/chaincode does not exist" ; exit 1; }
git sparse-checkout set --no-cone chaincode/*
git checkout
mv chaincode/* $ccToolsDemoPath
cd $ccToolsDemoPath
sudo rm -r $ccToolsDemoPath/chaincode
```

The chaincode code structure is different than normal chaincode as it's using the cc-tools framework.

## Edit the chaincode to became an FPC chaincode instead of normal fabric

Go to `$FPC_PATH/samples/chaincode/cc-tools-demo/main.go` and create the project structure.

```bash
cd $FPC_PATH/samples/chaincode/cc-tools-demo
```

The code presented here is not different from traditional Go Chaincode developed with CC-Tools framework. All the FPC specific protection mechanisms are handled by the FPC framework transparently.

The `main.go` contains the starting point of the chaincode.

To use FPC, we need to add some logic to instantiate our private chaincode and start it.
To do so, we use `shim.Start(fpc.NewPrivateChaincode(new(CCDemo)))` and since we're already in the FPC repo, remove the `go.mod` and `go.sum` files so we use the chaincode as part of the FPC package.

CC-tools-demo chaincode has its own packages that are needed, so we run `go get` in the `cc-tools-demo` folder.

Then, add the following code to `$FPC_PATH/samples/chaincode/cc-tools-demo/main.go` in functions `main()` and `runCCaaS()`:

```go
package main

import (
	"github.com/hyperledger/fabric-private-chaincode/samples/chaincode/cc-tools-demo/assettypes"
	"github.com/hyperledger/fabric-private-chaincode/samples/chaincode/cc-tools-demo/datatypes"
	"github.com/hyperledger/fabric-private-chaincode/samples/chaincode/cc-tools-demo/header"

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
	ccid := os.Getenv("CHAINCODE_PKG_ID") //This needs to be replaced

	//*CC-Tools Specific code (DO NOT OVERWRITE)*//

	var cc shim.Chaincode
	if os.Getenv("FPC_ENABLED") == "true" {
		cc = fpc.NewPrivateChaincode(new(CCDemo))
	} else {
		cc = new(CCDemo)
	}


	server := &shim.ChaincodeServer{
		CCID:     ccid,
		Address:  address,
		CC:       cc,
		TLSProps: *tlsProps,
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

EGO_CONFIG_FILE ?= $(FPC_PATH)/ecc_go/ccToolsDemoEnclave.json

```

Please make sure that in the file above the variable `TOP` points to the FPC root directory (i.e., `$FPC_PATH`) as it uses the `$FPC_PATH/ecc_go/build.mk` file.

**Note**: In our case, we need to change the build command in the `$FPC_PATH/ecc_go/build.mk` file at the `ecc` target to be `ego-go build $(GOTAGS) -o $(ECC_BINARY)` instead of `ego-go build $(GOTAGS) -o $(ECC_BINARY) main.go`

In `$FPC_PATH/samples/chaincode/cc-tools-demo` directory, to build the chaincode and package it as docker image, execute:

```bash
make
```

Note: this command runs inside the FPC dev environment and not your local host.

**Note**: If you faced this error:

```bash
/project/pkg/mod/github.com/hyperledger-labs/cc-tools@v1.0.1/mock/mockstub.go:146:22: cannot use stub (variable of type *MockStub) as shim.ChaincodeStubInterface value in argument to stub.cc.Init: *MockStub does not implement shim.ChaincodeStubInterface (missing method PurgePrivateData)
```

This is because there is a minor difference between the `ChaincodeStubInterface` used in the cc-tools `Mockstub` as it's missing the `PurgePrivateData` method.
To solve this, run `go mod vendor` in the `$FPC_PATH/samples/chaincode/cc-tools-demo` root directory to download all used packages and go to the file of the error to add the missing method there.

```bash
nano $FPC_PATH/vendor/github.com/hyperledger-labs/cc-tools/mock/mockstub.go
```

add the following function

```go
 	// PurgePrivateData ...
 	func (stub *MockStub) PurgePrivateData(collection, key string) error {
 		return errors.New("Not Implemented")
 	}

```

After building again, you can check that the `fpc/fpc-cc-tools-demo` image exists in your local docker registry using:

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

Now we are ready to launch the Fabric test network and install the FPC chaincode on it.
We begin with setting up the network with a single channel `mychannel`.

```bash
cd $FPC_PATH/samples/deployment/test-network/fabric-samples/test-network
./network.sh up -ca
./network.sh createChannel -c mychannel
```


### Set the needed env vars in the docker-compose file

From the code above, we need to set two env variables for the chaincode application to work and use FPC and chaincode-as-a-service (CCAAS). One way to do this is to go to `$FPC_PATH/samples/deployment/test-network/docker-compose.yml` and edit both `ecc.peer0.org1.example.com` and `ecc.peer0.org2.example.com` environment block to have

```yaml
- RUN_CCAAS=true
- FPC_ENABLED=true
```

### Install the chaincode

Once the network is up and running, we install the cc-tools-demo chaincode and the FPC Enclave Registry.
We provide a small shell script to make this task a bit easier.

```bash
export CC_ID=cc-tools-demo
export CC_PATH="$FPC_PATH/samples/chaincode/cc-tools-demo/"
export CC_VER=$(cat "$FPC_PATH/samples/chaincode/cc-tools-demo/mrenclave")
cd $FPC_PATH/samples/deployment/test-network
./installFPC.sh
```

Continue by running:

```bash
make ercc-ecc-start
```

You should see now four containers running (i.e., `cc-tools-demo.peer0.org1`, `cc-tools-demo.peer0.org2`, `ercc.peer0.org1`, and `ercc.peer0.org2`).

### Invoke simple getSchema transaction

Open a new terminal and connect to the `fpc-development-go-support` container by running

```bash
docker exec -it fpc-development-go-support /bin/bash
```

```bash
# prepare connections profile
cd $FPC_PATH/samples/deployment/test-network
./update-connection.sh

# # update the connection profile for external clients outside the FPC dev environment
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

Congratulations! You have successfully created an FPC chaincode with go using cc-tools and invoked it using our simple cli.

Now you can test all your work again by running the [test](./testTutorial.sh) script

## Next Step

CC-tools-demo also provides a unique API server called CCAPI that is able to communicate with the peers and execute transactions through a REST API. We integrated this either in the [CCAPI tutorial](../../application/ccapi/)