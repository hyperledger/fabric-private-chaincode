# The simple testing network

This gives an example on how to run your FPC Chaincode with a simple testing network build with the Fabric Smart Client.
The network is specified in [pkg/topology/fabric.go](pkg/topology/fabric.go).


## Preparation

Before we run the network, let's make sure we have built the Enclave Registry (ERCC) Container Images and a FPC Chaincode Container Image.
In this example we use the FPC [kvs-test-go sample](../../../chaincode/kv-test-go) chaincode.

Build ERCC image:
```bash
make -C $FPC_PATH/ercc all docker
```

Build the kv-test-go chaincode:
```bash
export CC_NAME=kv-test-go
export CC_PATH=$FPC_PATH/samples/chaincode/kv-test-go
make -C $CC_PATH build docker
``` 

To run the Fabric network we need the Fabric binaries.
We will use the following:
```bash
make -C $FPC_PATH/fabric
export FAB_BINS=$FPC_PATH/fabric/_internal/bin
```


## Run the network

```bash
make run
# OR
go run . network start --path ./testdata
```
Once the network is up and running, use a second terminal to interact with the network.
Uou can shut down the network using `CTRL+C`.

To clean up the network you can run `make clean` or `go run . network clean --path ./testdata`.


## Interact with your FPC Chaincode

In this sample we use the simple FPC Cli tool to invoke functions of our FPC Chaincode.


### How to use simple-cli-go

You can also use `$FPC_PATH/samples/application/simple-cli-go` instead of the simple-go application.

```bash
# Make fpcclient
cd $FPC_PATH/samples/application/simple-cli-go
make

# Run the following script and source the resulting environment variables
export CC_NAME=kv-test-go
$FPC_PATH/samples/deployment/fabric-smart-client/the-simple-testing-network/env.sh Org1
source Org1.env

# Interact with the FPC Chaincode
export FABRIC_LOGGING_SPEC=warning
./fpcclient invoke put_state hello world
./fpcclient query get_state hello
```

You can switch the user by running `./env.sh` either with `Org1` or `Org2`.


### Hyperledger Explorer

The test network includes a Hyperledger Blockchain explorer that you can use to see what is happening on the network.
By default, you can reach the Blockchain Explorer Dashboard via your browser on `http://localhost:8080` with username `admin` and password `admin`.


### Troubleshooting

Sometimes something goes wrong.
This is a collection of helpful commands to kill dangling network components.
```bash
killall peer
killall orderer
docker ps -a | grep fpc
```

We also provide shortcuts to run and clean the network:
```bash
make clean
make run
```
