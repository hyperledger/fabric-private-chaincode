# Setup your Development Environment (Option 2)

In this document we explain how to setup your FPC development environment by installing all required software components on your machine. We refer to this setup as `Option 2`. An alternative setup with docker you can find in [Option 1](setup-option1.md).


## Requirements

Make sure that you have the following required dependencies installed:
* Linux (OS) (we recommend Ubuntu 22.04, see [list](https://github.com/intel/linux-sgx#prerequisites) supported OS)

* CMake v3.5.1 or higher

* [Go](https://golang.org/) 1.25.x or higher

* Docker 18.09 (or higher) and docker-compose 1.25.x (or higher)
  Note that version from Ubuntu 18.04 is not recent enough!  To upgrade, install a recent version following the instructions from [docker.com](https://docs.docker.com/compose/install/), e.g., for version 1.25.4 execute	
  ```bash	
  sudo curl -L "https://github.com/docker/compose/releases/download/1.25.4/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose	
  sudo chmod +x /usr/local/bin/docker-compose
  ```

  To install docker-componse 1.25.4 from [docker.com](https://docs.docker.com/compose/install/), execute
  ```bash
  sudo curl -L "https://github.com/docker/compose/releases/download/1.25.4/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
  sudo chmod +x /usr/local/bin/docker-compose
  ``` 

* yq v4.x
  You can install `yq` via `go get`.
  ```bash
    go get github.com/mikefarah/yq/v4
  ```

* Protocol Buffers
    - Protocol Buffers 3.0.x needed for the Intel SGX SDK
    - Protocol Buffers 3.11.x or higher and [Nanopb](http://github.com/nanopb/nanopb) 0.4.7

* SGX PSW & SDK v2.22 for [Linux](https://01.org/intel-software-guard-extensions/downloads)
  (alternatively, you could also install it from the [source](https://github.com/intel/linux-sgx)

* Credentials for Intel Attestation Service, read [here](../README.md#intel-attestation-service-ias) (for hardware-mode SGX)

* [Intel Software Guard Extensions SSL](https://github.com/intel/intel-sgx-ssl)
  (we recommend using tag `3.0_Rev2` OpenSSL `3.0.12`)

* Hyperledger [Fabric](https://github.com/hyperledger/fabric/tree/v2.5.9) v2.5.9

* Clang-format 6.x or higher

* jq

* hex (for Ubuntu, found in package basez)

* A recent version of [PlantUML](http://plantuml.com/), including Graphviz, for building documentation. See [Documentation](../README.md#building-documentation) for our recommendations on installing. The version available in common package repositories may be out of date.

## Intel SGX SDK and SSL

Fabric Private Chaincode requires the Intel [SGX SDK](https://github.com/intel/linux-sgx) and
[SGX SSL](https://github.com/intel/intel-sgx-ssl) to build the main components of our framework and to develop and build
your first private chaincode.

Install the Intel SGX software stack for Linux by following the
official [documentation](https://github.com/intel/linux-sgx). Please make sure that you use the
SDK version as denoted above in the list of requirements.

For SGX SSL, just follow the instructions on the [corresponding
github page](https://github.com/intel/intel-sgx-ssl). In case you are
building for simulation mode only and do not have HW support, you
might also want to make sure that [simulation mode is set](https://github.com/intel/intel-sgx-ssl#available-make-flags)
when building and installing it.

Once you have installed the SGX SDK and SSL for SGX SDK please double check that `SGX_SDK` and `SGX_SSL` variables
are set correctly in your environment.


## Protocol Buffers

We use *nanopb*, a lightweight implementation of Protocol Buffers, inside the enclaves to parse blocks of
transactions. Install nanopb by following the instruction below. For this you need a working Google Protocol Buffers
compiler with python bindings (e.g. via `apt-get install protobuf-compiler python3-protobuf libprotobuf-dev`).
For more detailed information consult the official nanopb documentation http://github.com/nanopb/nanopb.
```bash
export NANOPB_PATH=/path-to/install/nanopb/
git clone https://github.com/nanopb/nanopb.git $NANOPB_PATH
cd $NANOPB_PATH
git checkout nanopb-0.4.7
cd generator/proto && make
```

Make sure that you set `$NANOPB_PATH` as it is needed to build Fabric Private Chaincode.

Moreover, in order to build Fabric protobufs we also require a newer Protobuf compiler than what is provided as standard Ubuntu package and is used to build the
Intel SGX SDK. For this reason you will have to download and install another version and use it together with Nanopb. Do not install the new protobuf, though, such that it is not found in your standard PATH but instead define the `PROTOC_CMD`, either as environment variable or via `config.override.mk` to point to the new `protoc` binary
```bash
wget https://github.com/protocolbuffers/protobuf/releases/download/v22.3/protoc-22.3-linux-x86_64.zip
unzip protoc-22.3-linux-x86_64.zip -d /usr/local/proto3
export PROTOC_CMD=/usr/local/proto3/bin/protoc
```

## Hyperledger Fabric

Our project fetches the latest supported Fabric binaries during the build process automatically.
However, if you want to use your own Fabric binaries, please checkout Fabric 2.5.9 release using the following commands:
```bash
export FABRIC_PATH=$GOPATH/src/github.com/hyperledger/fabric
git clone https://github.com/hyperledger/fabric.git $FABRIC_PATH
cd $FABRIC_PATH; git checkout tags/v2.5.9
```

Note that Fabric Private Chaincode may not work with the Fabric `main` branch.
Therefore, make sure you use the Fabric `v2.5.9` tag.
Make sure the source of Fabric is in your `$GOPATH`.
