# The simple testing network

This gives an example how to run your FPC Chaincode with a simple testing network build with the Fabric Smart Client.
The network is specified in [pkg/topology/fabric.go](pkg/topology/fabric.go).

In this demo we use the demo fpc-echo chaincode.

## Fabric network

### Prep

Let's make sure we have built the Enclave Registry Container Images and a FPC Chaincode Container Image.
In this example we will run the [echo sample](../../../chaincode/echo). 

Build ERCC image
```bash
make -C $FPC_PATH/ercc all docker
```

Build FPC-Echo image
```bash
export CC_NAME=echo
export CC_PATH=$FPC_PATH/samples/chaincode/echo
make -C $CC_PATH
make -C $FPC_PATH/ecc DOCKER_IMAGE=fpc/fpc-${CC_NAME}$${HW_EXTENSION} DOCKER_ENCLAVE_SO_PATH=$CC_PATH/_build/lib all docker
``` 

Download fabric binaries:

```bash
curl -sSL https://bit.ly/2ysbOFE | bash -s -- 2.2.7 1.4.9 -s -d
export FAB_BINS=$(pwd)/bin
```

### Build

```bash
go build -o tstn
```

### Run the network

```bash
./tstn network start --path ./testdata
```
Use a second terminal to interact with the network

## Interact with your FPC Chaincode

In this sample we use the simple FPC Cli tool to invoke functions of our FPC Chaincode.

### How to use simple-cli-go

You can also use `$FPC_PATH/samples/application/simple-cli-go` instead of the simple-go application.

```bash
# Make fpcclient
cd $FPC_PATH/samples/application/simple-cli-go
make

# Run the following script and copy and paste the export statements back into your terminal
$FPC_PATH/samples/deployment/fabric-smart-client/the-simple-testing-network/env.sh Org1

# Interact with the FPC Chaincode
./fpcclient invoke foo
./fpcclient query foo
```

You can switch the user by running `./env.sh` either with `Org1` or `Org2`.


## Launch FPC Chaincode manually
TODO

### Build and package

TODO
- Build the FPC chaincode as docker image
- Create chaincode `package.tar.gz` to install at a peer

### Install

TODO

### Launch with Docker

TODO
- Spawn FPC chaincode container 
