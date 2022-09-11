# Hyperledger Global Forum 2022 - FPC Live Demo

This demo shows how to port an existing chaincode from the [fabric samples](https://github.com/hyperledger/fabric-samples) repository to a FPC chaincode and run it with the Fabric Smart Client Integration test suite.

Let's dive straight into action.
We will use the [fabric-samples/asset-transfer-basic](https://github.com/hyperledger/fabric-samples/tree/main/asset-transfer-basic/chaincode-go) chaincode.

### Write the chaincode

Modify the `main.go` template by adding the code to instantiate the smart contract as Private Chaincode.

```go
assetChaincode, _ := contractapi.NewChaincode(&chaincode.SmartContract{})
chaincode := fpc.NewPrivateChaincode(assetChaincode)
```

Make sure you import:
```go
"github.com/hyperledger/fabric-contract-api-go/contractapi"
fpc "github.com/hyperledger/fabric-private-chaincode/ecc_go/chaincode"
"github.com/hyperledger/fabric-samples/asset-transfer-basic/chaincode-go/chaincode"
```


### Build chaincode

In order to build the chaincode, we will use the tools provided by FPC.
Note that you can use the FPC docker-based development environment.

Set `CC_NAME ?= fpc-basic-asset-transfer` inside the `Makefile`.

```bash
make
docker images | grep fpc-basic
cat details.env
```

### Start a Fabric network

Go to `$FPC_PATH/samples/deployment/fabric-smart-client/the-simple-testing-network`.
Review the `topology.go`, which defines a simple Fabric network with two organizations.
Note that we enable FPC by setting the chaincode name and the docker image. 

```bash
export FAB_BINS=$(pwd)/bin
source $FPC_PATH/samples/demos/hlgf22/details.env
make run
```

Once the network is started we can open Hyperledger Explorer in a browser at `http://localhost:8080/` with username `admin` and password `admin`.
We can see that the `fpc-basic-asset-transfer` chaincode is installed.
Moreover, we see the FPC Enclave Registry chaincode is installed as well.

### Invoke the chaincode

Finally, we will interact with the deployed FPC Chaincode using the FPC [simple-cli-go](https://github.com/hyperledger/fabric-private-chaincode/tree/go-support-preview/samples/application/simple-cli-go).

Go to `$FPC_PATH/samples/application/simple-cli-go` and build the app. The CLI app uses the FPC Client SDK to make the interaction with the FPC chaincode as easy as possible and hiding all the complexity to protect the transaction arguments.

```bash
cd $FPC_PATH/samples/application/simple-cli-go
make

# configure cli environment
source $FPC_PATH/samples/demos/hlgf22/details.env
$FPC_PATH/samples/deployment/fabric-smart-client/the-simple-testing-network/env.sh Org1
source Org1.env

./fpcclient invoke createAsset 101 green 23 Marcus 9999
./fpcclient invoke readAsset 101
```
