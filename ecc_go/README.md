<!---
Licensed under Creative Commons Attribution 4.0 International License
https://creativecommons.org/licenses/by/4.0/
--->
# Go Chaincode Support for Fabric Private Chaincode

*Note - Go Chaincode support is currently under development and should be considered experimental.*

## Overview

This directory contains the components to enable Go Chaincode support for Fabric Private Chaincode (FPC).
This feature relies on the [Ego project](https://www.ego.dev/) to build and execute go application with Intel SGX.

In particular, it contains:
- FPC Go Library to be used with your Go Chaincode.
- Building and packaging utilities

### FPC Go Library

We aim to support Go Chaincode without the need to refactor existing code but still benefit from the security properties added by FPC.
However, we currently support only a limited Chaincode API with common functionality to enable a broad set of applications.
We refer to [shim.go](chaincode/enclave_go/shim.go) for the full list of supported functions.
Note that calling unsupported shim functions, currently results in a `panic`.

To use FPC, you simply need to wrap your chaincode with the FPC Go Library. Here an example:
```go
package main

import (
	"github.com/hyperledger/fabric-chaincode-go/shim"
	fpc "github.com/hyperledger/fabric-private-chaincode/ecc_go/chaincode"
	"github.com/your-project/project/chaincode"
)

func main() {

	... 
	
	// create private chaincode
	privateChaincode := fpc.NewPrivateChaincode(&chaincode.YourChaincode{})

	// start chaincode as a service
	server := &shim.ChaincodeServer{
		CC: privateChaincode,
	}

	if err := server.Start(); err != nil {
		panic(err)
	}
}
```

### Building and packaging

In contrast to traditional Fabric Go Chaincode, FPC uses the ego compiler to build the chaincode and then package it in a docker image.
To simplify this process, we provide you a `Dockerfile` and a `build.mk` which you can use in your project.
Here an example of a `Makefile`:
```Makefile
include $(FPC_PATH)/ecc_go/build.mk
CC_NAME ?= your-chaincode-name
```

Your make file now comes with standard build targets, such as, `build`, `test`, and `clean`.
See `build.mk` for a full list of available build targets.

## Installation

### Install Ego inside dev environment

This installation assumes a working FPC dev environment.
You can find all setup information in the getting started section of our [README.md](../README.md#setup-your-development-environment).

If you are **not** using the FPC dev docker container, you need to install the ego compiler manually.
Install ego by running the following:
```bash
wget -qO- https://download.01.org/intel-sgx/sgx_repo/ubuntu/intel-sgx-deb.key | apt-key add
add-apt-repository "deb [arch=amd64] https://download.01.org/intel-sgx/sgx_repo/ubuntu `lsb_release -cs` main"
wget https://github.com/edgelesssys/ego/releases/download/v1.0.0/ego_1.0.0_amd64.deb
apt install ./ego_1.0.0_amd64.deb build-essential libssl-dev
```
You can find more information about ego installation on the official [documentation](https://docs.edgeless.systems/ego/#/getting-started/install).

## Examples

So see FPC Go Chaincode Support in action, we provide a few examples in our repository.

### Simple Asset Tutorial

We provide a quick getting started tutorial that walks you through the process to write, build, deploy, and run FPC Go Chaincode.
You can find the tutorial in [samples/chaincode/simple-asset-go](../samples/chaincode/simple-asset-go).

### Auction

We provide a sample auction [samples/chaincode/auction-go](../samples/chaincode/auction-go).
You can run it using the integration test suite as follows:
```bash
cd $FPC_PATH/integration/go_chaincode/auction
make
```

### KV-Test

Another example is provided [samples/chaincode/kv-test-go](../samples/chaincode/kv-test-go).
You can run it using the integration test suite as follows:
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
