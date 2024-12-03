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

* This tutorial presumes that you have installed FPC as described in the FPC [README.md](../../../README.md#clone-fabric-private-chaincode) and `$FPC_PATH` is set accordingly.
* We need a working FPC development environment. As described in the "Setup your Development Environment" Section of the FPC [README.md](../../../README.md#setup-your-development-environment), you can use our docker-based dev environment (Option 1) or setup your local development environment (Option 2).
  We recommend using the docker-based development environment and continue this tutorial within the dev container terminal.
* Moreover, within your FPC development you have already installed the FPC Go Chaincode Support components.
  See the installation steps in [ecc_go/README.md](../../../ecc_go/README.md#installation).
* We assume that you are familiar with Fabric chaincode development in go.
  Most of the steps in this tutorial follow the normal Fabric chaincode development process, however, there are a few differences that we will highlight here.
* Also, since the tutorial is on the integration between cc-tools and FPC, we expect you to have a grasp knowledge of [cc-tools](https://github.com/hyperledger-labs/cc-tools) framework and that you've at least tried to run the [cc-tools-demo](https://github.com/hyperledger-labs/cc-tools-demo) once by yourself on a Fabric network

## Clone the cc-tools-demo chaincode

We need to clone the chaincode folder from the [cc-tools-demo](https://github.com/hyperledger-labs/cc-tools-demo) repository here.
Run the following script inside the dev environment:

```bash
$FPC_PATH/samples/chaincode/cc-tools-demo/setup.sh
```

**Note**: You might encounter permission errors if you run this outside the FPC dev container. In that case you may want to use `sudo`.

## Edit the chaincode to become an FPC chaincode instead of normal fabric

The chaincode code structure is different than normal chaincode as it's using the cc-tools framework.

Go to `$FPC_PATH/samples/chaincode/cc-tools-demo/main.go` and view the project structure.

```bash
cd $FPC_PATH/samples/chaincode/cc-tools-demo
```

The code presented here is not different from traditional Go Chaincode developed with CC-Tools framework. All the FPC specific protection mechanisms are handled by the FPC framework transparently.

The `main.go` contains the starting point of the chaincode.

To use FPC, we need to add some logic to instantiate our private chaincode and start it.
To do so, we use `shim.Start(fpc.NewPrivateChaincode(new(CCDemo)))` and since we're already in the FPC repo we will point to the local FPC package.

Go to the `go.mod` file in the `cc-tools-demo` chaincode folder and add the following replace line before the `require( )` block:

```go
replace github.com/hyperledger/fabric-private-chaincode => ../../../ 
```

CC-tools-demo chaincode has its own packages that are needed, so we run `go mod tidy` in the `cc-tools-demo` folder.

Then, edit the chaincode in the `$FPC_PATH/samples/chaincode/cc-tools-demo/main.go` file:

For import block, add the fpc package

```go
import (
	///Keep everything as is
	fpc "github.com/hyperledger/fabric-private-chaincode/ecc_go/chaincode"
)
```

In the `main()` function, replace the following `if` statement:

```go
if os.Getenv("RUN_CCAAS") == "true" {
	err = runCCaaS()
} else {
	err = shim.Start(new(CCDemo))
}

```

With this:

```go
if os.Getenv("RUN_CCAAS") == "true" {
	err = runCCaaS()
} else {
	if os.Getenv("FPC_ENABLED") == "true" {
		err = shim.Start(fpc.NewPrivateChaincode(new(CCDemo)))
	} else {
		err = shim.Start(new(CCDemo))
	}
}
```

In the `runCCaaS()` function, replace the following line:

```go
ccid := os.Getenv("CHAINCODE_ID")
```

With this

```go
ccid := os.Getenv("CHAINCODE_PKG_ID")
```

And replace this

```go
server := &shim.ChaincodeServer{
	CCID:     ccid,
	Address:  address,
	CC:       new(CCDemo),
	TLSProps: *tlsProps,
}
```

With this

```go

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

```

## Building FPC Go Chaincode

First, to update the go dependencies run (inside the dev environment):

```bash
cd $FPC_PATH/samples/chaincode/cc-tools-demo
go get github.com/hyperledger/fabric-private-chaincode
go mod tidy
go get
```

Create a `Makefile` (i.e., `touch $FPC_PATH/samples/chaincode/cc-tools-demo/Makefile`) with the following content:

```Makefile
TOP = ../../..
include $(TOP)/ecc_go/build.mk

CC_NAME ?= fpc-cc-tools-demo

EGO_CONFIG_FILE = $(FPC_PATH)/samples/chaincode/cc-tools-demo/ccToolsDemoEnclave.json
ECC_MAIN_FILES=$(FPC_PATH)/samples/chaincode/cc-tools-demo

```

Please make sure that in the file above the variable `TOP` points to the FPC root directory (i.e., `$FPC_PATH`) as it uses the `$FPC_PATH/ecc_go/build.mk` file.

In `$FPC_PATH/samples/chaincode/cc-tools-demo` directory, to build the chaincode and package it as docker image, execute:

```bash
make
```

**Note**: For those who have arm-based computers, you should use the method recommended in the [main readme](https://github.com/hyperledger/fabric-private-chaincode?tab=readme-ov-file#fpc-playground-for-non-sgx-environments) for building both the cc-tools-demo chaincode as well as ercc.

**Note**: this command runs inside the FPC dev environment and not your local host.

**Note**: If you faced this error:

```bash
/project/pkg/mod/github.com/hyperledger-labs/cc-tools@v1.0.1/mock/mockstub.go:146:22: cannot use stub (variable of type *MockStub) as shim.ChaincodeStubInterface value in argument to stub.cc.Init: *MockStub does not implement shim.ChaincodeStubInterface (missing method PurgePrivateData)
```

This is because there is a minor difference between the `ChaincodeStubInterface` used in the cc-tools `Mockstub` as it's missing the `PurgePrivateData` method.
To solve this, run the following to download all used packages

```bash
cd $FPC_PATH/samples/chaincode/cc-tools-demo
go mod vendor
```

Edit the file of the error `$FPC_PATH/samples/chaincode/cc-tools-demo/vendor/github.com/hyperledger-labs/cc-tools/mock/mockstub.go` and add the missing method there:

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

### Set the needed env vars in the docker-compose file and start the chaincode

From the code above, we need to set two env variables for the chaincode application to work and use FPC and chaincode-as-a-service (CCAAS).

```yaml
- RUN_CCAAS=true
- FPC_ENABLED=true
```

To achieve this, we created extra configurations for the start command at `$FPC_PATH/samples/chaincode/cc-tools-demo/cc-tools-demo-compose.yaml`.

Continue by running:

```bash
export EXTRA_COMPOSE_FILE="$FPC_PATH/samples/chaincode/cc-tools-demo/cc-tools-demo-compose.yaml"
make ercc-ecc-start
```

You should see now four containers running (i.e., `cc-tools-demo.peer0.org1`, `cc-tools-demo.peer0.org2`, `ercc.peer0.org1`, and `ercc.peer0.org2`).

### Invoke simple getSchema transaction

Open a new terminal and connect to the `fpc-development-main` container by running

```bash
docker exec -it fpc-development-main /bin/bash
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

**Note**: In cc-tools-demo, most of these transactions set permissions to filter which orgs are allowed to invoke it or not. The current organization used in this script is "Org1MSP". Also, beware that org names are case sensitive

## Next Step

CC-tools-demo also provides a unique API server called CCAPI that is able to communicate with the peers and execute transactions through a REST API. We integrated this either in the [CCAPI tutorial](../../application/ccapi/).
