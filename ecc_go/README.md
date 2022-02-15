<!---
Licensed under Creative Commons Attribution 4.0 International License
https://creativecommons.org/licenses/by/4.0/
--->
# Go Chaincode Support for Fabric Private Chaincode

*Note - Go Chaincode support is currently under development and should be considered experimental.*

This directory contains the components to enable go support Fabric Private Chaincode.

FPC Go Chaincode Support relies on the [Ego project](https://www.ego.dev/) to build and execute go application with Intel SGX.

## Overview
 TODO
 
In particular, it contains:
- Go library to be included in a go chaincode
- Building and packaging utilities

### Architecture
TODO

## Install

### Install Ego inside dev environment

This installation assumes a working FPC dev environment.
You can find all setup information in the getting started section of our [README.md](../README.md#setup-your-development-environment).

If you are **not** using the FPC dev docker container, you need to install the ego compiler manually.
Install ego by running the following:
```bash
wget -qO- https://download.01.org/intel-sgx/sgx_repo/ubuntu/intel-sgx-deb.key | apt-key add
add-apt-repository "deb [arch=amd64] https://download.01.org/intel-sgx/sgx_repo/ubuntu `lsb_release -cs` main"
wget https://github.com/edgelesssys/ego/releases/download/v0.4.1/ego_0.4.1_amd64.deb
apt install ./ego_0.4.1_amd64.deb build-essential libssl-dev
```
You can find more information about ego installation on the official [documentation](https://docs.edgeless.systems/ego/#/getting-started/install).

### Create ccenv-go

In order to package and run FPC go chaincode we use the `ccenv-go` docker image.
Run the following to create the docker image.

```bash
cd $FPC_PATH/utils/docker/
make ccenv-go
```

Now you have all you need to get started with your first FPC go chaincode. 

## Examples

### Simple Asset Tutorial

We provide a quick getting started tutorial that walks you through the process to write, build, deploy, and run FPC go chaincode.
You can find the tutorial [here](../samples/chaincode/asset-transfer-go).


### Auction

We provide a simple sample auction [here](../samples/chaincode/auction-go). You can run it using the integration test suite as follows:
```bash
cd $FPC_PATH/integration/go_chaincode/auction
make
```


### KV-Test

Another example is provided [here](../samples/chaincode/kv-test-go). You can run it using the integration test suite as follows:
```bash
cd $FPC_PATH/integration/go_chaincode/kv_test/
make
```


## Developer notes

Here provide a collection of useful developer notes which may help you while developing.  


### Kill hanging containers
```bash
docker kill $(docker ps -a -q --filter ancestor=fpc/ercc --filter ancestor=fpc/fpc-auction-go)
docker rm $(docker ps -a -q --filter ancestor=fpc/ercc --filter ancestor=fpc/fpc-auction-go)
```

More to come ...

## TODOs

The following components are not yet implemented.

- [ ] HW Attestation support
- [ ] Fabric contract API
