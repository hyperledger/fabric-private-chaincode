# FPC Go Simple Asset Tutorial

*Note - Go Chaincode support is currently under development and should be considered experimental.*

This tutorial shows how to create, build, install and test go chaincode using the Fabric Private Chaincode (FPC) framework.
Here we focus on the development of FPC Chaincode in go. There exists a companion [hello world tutorial](../helloworld) illustrating the use of FPC with C++ chaincode.

This tutorial illustrates a simple usecase where a FPC chaincode is used to store a single asset, `asset1` in the ledger and then retrieve the latest value of `asset1`.  Here are the steps to accomplish this:

TODO UPDATE

* Develop chaincode
* Launch Fabric network
* Install and instantiate chaincode on the peer
* Invoke transactions (`storeAsset` and `retrieveAsset`)
    * by using the Peer CLI and
    * by using the FPC Client SDK for Go
* Shut down the network

## Prerequisites
This tutorial presumes that you have installed FPC on your `$GOPATH` as described in the FPC [README.md](../../../README.md#requirements) and `$FPC_PATH` is set accordingly.
Additionally, you have already installed the extension to FPC go chaincode.
See the installation steps [here](../../../ecc_go/README.md#install).  

We also assume that you are familiar with Fabric chaincode development in go.
Most of the steps in this tutorial following the normal Fabric chaincode development process, however, there are a few differences we will highlight here.

## Writing Go Chaincode
Go to `$FPC_PATH/samples/chaincode/simple-asset-go` and create our project structure.
```bash
cd $FPC_PATH/samples/chaincode/simple-asset-go
mkdir chaincode
touch chaincode/chaincode.go
touch main.go
```

The `chaincode` directory will contain our chaincode logic. Here we just have a single `chaincode.go`.
The `main.go` contains the starting point of the chaincode.

Let's first focus on the chaincode logic. Add the following code to `chaincode/chaincode.go`:
```go
package chaincode

import (
  "fmt"

  "github.com/hyperledger/fabric-chaincode-go/shim"
  pb "github.com/hyperledger/fabric-protos-go/peer"
)

type SimpleAsset struct {
}

func (t *SimpleAsset) Init(stub shim.ChaincodeStubInterface) pb.Response {
  return shim.Success(nil)
}

func (t *SimpleAsset) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
  switch f, _ := stub.GetFunctionAndParameters(); f {
  case "store":
    return storeAsset(stub)
  case "retrieve":
    return retrieveAsset(stub)
  }

  return shim.Error("unknown function")
}

func storeAsset(stub shim.ChaincodeStubInterface) pb.Response {
  _, args := stub.GetFunctionAndParameters()

  if len(args) < 2 {
    return shim.Error("not enough arguments")
  }

  assetName, value := args[0], args[1]

  if err := stub.PutState(assetName, []byte(value)); err != nil {
    return shim.Error("something went wrong")
  }

  return shim.Success([]byte("OK"))
}

func retrieveAsset(stub shim.ChaincodeStubInterface) pb.Response {
  _, args := stub.GetFunctionAndParameters()

  if len(args) < 1 {
    return shim.Error("not enough arguments")
  }

  assetName := args[0]

  value, err := stub.GetState(assetName)
  if err != nil {
    return shim.Error("something went wrong")
  }

  if len(value) == 0 {
    shim.Success([]byte("NOT FOUND"))
  }

  return shim.Success([]byte(fmt.Sprintf("%s:%s", assetName, value)))
}

```

You can see that this code implements the `Init` and `Invoke` methods of a chaincode. 
The chaincode supports two types of transactions: `store` and `retrieve`, which are implemented by `storeAsset` and `retrieveAsset` methods respectively.

Let's first focus on the `store` transaction, which simply saves the value of an asset, implemented by the `storeAsset` method.
We extract the `assetName` and the `value` from chaincode invocation parameters and use the `PutState` function to store them.
Similarly, let us add the logic for the `retrieve` transaction, which reads the value of an asset by calling `GetState` and returns it.

So far the code presented here is not different from traditional go chaincode. All the FPC specific protect mechanisms
are handled by the FPC framework transparently.

To complete the code, we need to add some logic to instantiate our private chaincode and start it.
To do so, we use `fpc.NewPrivateChaincode(&chaincode.AssetExample{})`.
Add the following code to our `$FPC_PATH/samples/chaincode/simple-asset-go/main.go`:
```go
package main

import (
	"os"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	fpc "github.com/hyperledger/fabric-private-chaincode/ecc_go/chaincode"
	"github.com/hyperledger/fabric-private-chaincode/samples/chaincode/simple-asset-go/chaincode"
)

func main() {
	ccid := os.Getenv("CHAINCODE_PKG_ID")
	addr := os.Getenv("CHAINCODE_SERVER_ADDRESS")

	// create private chaincode
	privateChaincode := fpc.NewPrivateChaincode(&chaincode.SimpleAsset{})

	// start chaincode as a service
	server := &shim.ChaincodeServer{
		CCID:    ccid,
		Address: addr,
		CC:      privateChaincode,
		TLSProps: shim.TLSProperties{
			Disabled: true, // just for testing good enough
		},
	}

	if err := server.Start(); err != nil {
		panic(err)
	}
}
```

## Building FPC Go Chaincode 

Create a `Makefile` (i.e., `touch $FPC_PATH/samples/chaincode/simple-asset-go/Makefile`) with the following content: 
```Makefile
TOP = ../../..
include $(TOP)/ecc_go/build.mk

CC_NAME ?= simple-asset
```

Please make sure that in the file above the variable `TOP` points to the FPC root directory (i.e., `$FPC_PATH`).

In `$FPC_PATH/samples/chaincode/simple-asset-go` directory, to build the chaincode and package it as docker image, execute:
```bash
make
```

After building, you can check your local docker registry that the `fpc/fpc-simple-asset-go` image exists using
```bash
docker images | grep simple-asset
```


## Time to test!

Next step is to test the chaincode by invoking transactions, for which you need a basic Fabric network with a channel.
We will use the test network provided in `$FPC_PATH/samples/deployment/test-network`.
To invoke the chaincode, we will use the `simple-cli` application in `$FPC_PATH/samples/application/simple-cli-go`

### Enclave Registry

To run FPC chaincode we need to prepare the docker images for the FPC Enclave Registry (ERCC).
In case you have not yet created them, run `make -C $FPC_PATH/ercc build docker`.

### Prepare the test network

We already provide a detailed tutorial how to use FPC with the test network in [$FPC_PATH/samples/deployment/test-network](../../deployment/test-network).
However, for completeness let's go through the required steps you need to run once.

```bash
cd $FPC_PATH/samples/deployment/test-network
git clone https://github.com/hyperledger/fabric-samples
cd $FPC_PATH/samples/deployment/test-network/fabric-samples
# no we pick a specific version here to have stable experience :)
git checkout -b "works" 98028c7
curl -sSL https://bit.ly/2ysbOFE | bash -s -- 2.3.3 1.4.9 -s
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
export CC_ID=simple-asset-go
export CC_PATH="$FPC_PATH/samples/chaincode/simple-asset-go/"
export CC_VER=$(cat "$FPC_PATH/samples/chaincode/simple-asset-go/mrenclave")
cd $FPC_PATH/samples/deployment/test-network
./installFPC.sh
```

Note that the `installFPC.sh` script returns an export statement you need now.
Copy it to your terminal and continue with running `make ercc-ecc-start`.
You should see now four containers running (i.e., `simple-asset.peer0.org1`, `simple-asset.peer0.org2`, `ercc.peer0.org1`, and `ercc.peer0.org2`). 

### Invoke simple asset

```bash

# make fpcclient
cd $FPC_PATH/samples/application/simple-cli-go
make

# export fpcclient settings
export CC_NAME=simple-asset
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

# init our enclave
./fpcclient init $CORE_PEER_ID

# interact with the FPC Chaincode
./fpcclient invoke store diamond 10000
./fpcclient query retrieve diamond

```

Congratulations! You have successfully created a FPC chaincode with go and invoked it using our simple cli.
