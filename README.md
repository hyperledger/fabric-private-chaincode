<!---
Licensed under Creative Commons Attribution 4.0 International License
https://creativecommons.org/licenses/by/4.0/
--->
# Hyperledger Fabric Private Chaincode

Hyperledger Fabric Private Chaincode (FPC) enables the execution of chaincodes
using Intel SGX for Hyperledger Fabric.

The transparency and resilience gained from blockchain protocols ensure the
integrity of blockchain applications and yet contradicts the goal to keep
application state confidential and to maintain privacy for its users.

To remedy this problem, this project uses Trusted Execution Environments
(TEEs), in particular Intel Software Guard Extensions (SGX), to protect the
privacy of chaincode data and computation from potentially untrusted peers.

Intel SGX is the most prominent TEE today and available with commodity
CPUs. It establishes trusted execution contexts called enclaves on a CPU,
which isolate data and programs from the host operating system in hardware and
ensure that outputs are correct.

This project provides a framework to develop and execute Fabric chaincode within
an enclave.  It allows to write chaincode applications where the data is
encrypted on the ledger and can only be accessed in clear by authorized
parties. Furthermore, Fabric extensions for chaincode enclave registration
and transaction verification are provided.

Fabric Private Chaicode is based on the work in the paper:

* Marcus Brandenburger, Christian Cachin, Rüdiger Kapitza, Alessandro
  Sorniotti: Blockchain and Trusted Computing: Problems, Pitfalls, and a
  Solution for Hyperledger Fabric. https://arxiv.org/abs/1805.08541

This project was accepted via a Hyperledger Fabric [RFC](https://github.com/hyperledger/fabric-rfcs/blob/main/text/0000-fabric-private-chaincode-1.0.md) and is now under development.
We provide an initial proof-of-concept implementation of the proposed
architecture. Note that the code provided in this repository is still prototype code
and not yet meant for production use!

For up to date information about our community meeting schedule, past
presentations, and info on how to contact us please refer to our
[wiki page](https://wiki.hyperledger.org/display/fabric/Hyperledger+Fabric+Private+Chaincode).

## Architecture and components

### Overview

This project extends a Fabric peer with the following components: A chaincode
enclave that executes a particular chaincode, running inside SGX.
In the untrusted part of the peer, an enclave registry maintains
the identities of all chaincode enclaves and an enclave transaction validator
that is responsible for validating transactions executed by a chaincode
enclave before committing them to the ledger.

The following diagram shows the architecture:

![Architecture](docs/images/arch.png)

The system consists of the following components:

1. *Chaincode enclave:* The chaincode enclave executes one particular
   chaincode, and thereby isolates it from the peer and from other
   chaincodes. A chaincode library acts as intermediary between the chaincode
   in the enclave and the peer. The chaincode enclave exposes the Fabric
   chaincode interface and extends it with additional support for state
   encryption, attestation, and secure blockchain state access. This
   code is executed inside an Intel SGX enclave.

1. *Enclave Endorsement validation:* The enclave endorsement validation
   complements the peer’s validation system and is responsible for
   validating transactions produced by a chaincode enclave. In
   particular, the validator checks that a transaction contains a
   valid signature issued by a registered chaincode enclave. Iff the
   validation is successful, it causes the state-updates of the
   transaction to be committed to the ledger. This code is a normal Fabric
   transaction, i.e., executed and endorsed on multiple peers as
   required by the organization trust.

1. *FPC Chaincode Pkg:*
   This component bundles together the chaincode enclave and the enclave endorsement validation logic into a fabric chaincode.
   It also includes a shim component which 
   (a) proxies the chaincode enclave shim functionality, e.g., access to ledger, to the fabric peer, and
   (b) dispatches FPC flows to either the chaincode enclave (via `__invoke` queries) or to the enclave endorsement validation logic (via `__endorse` transactions).

1. *Enclave registry:* The enclave registry (`ercc`) is a chaincode that runs outside
   SGX and maintains a list of all existing chaincode enclaves in the
   network. It performs attestation with the chaincode enclave and stores the
   attestation result on the blockchain. The attestation demonstrates that a
   specific chaincode executes in an actual enclave. This enables the peers
   and the clients to inspect the attestation of a chaincode enclave before
   invoking chaincode operations or committing state changes.

More design information can be found [here](docs/architecture-design.md)

### Source organization

- [`client_sdk`](client_sdk/go/): The FPC Go Client SDK
- [`cmake`](cmake/): CMake build rules shared across the project
- [`common`](common/): Shared C/C++ code
- [`config`](config/): SGX configuration
- [`docs`](docs/): Documentation and design documents
- [`ecc_enclave`](ecc_enclave/): C/C++ code for chaincode enclave
    (including the trusted code running inside an enclave)
- [`ecc`](ecc/): Go code for FPC chaincode package, including
    dispatcher and (high-level code for) enclave endorsement validation.
- [`ecc_go`](ecc_go/): Go code for FPC Go Chaincode Support
- [`ercc`](ercc/): Go code for Enclave Registry Chaincode
- [`samples`](samples/): FPC Samples
- [`fabric`](fabric/): FPC wrapper for Fabric peer and utilities to
    start and stop a simple Fabric test network with FPC enabled, used
    by integration tests.
- [`integration`](integration/): FPC integration tests.
- [`internal`](internal/): Shared Go code
- [`protos`](protos/): Protobuf definitions
- [`scripts`](scripts/): Scripts used in build process.
- [`utils/docker`](utils/docker): Docker images and their build process.
- [`utils/fabric`](utils/fabric): Various Fabric helpers.


## Releases

For all releases go to the [Github Release Page](https://github.com/hyperledger/fabric-private-chaincode/releases).

*WARNING: This project is in continous development and the `main`
 branch will not always be stable. Unless you want to actively
 contribute to the project itself, we advise you to use the latest release.*



## Getting started

The following steps guide you through the build phase and configuration, for
deploying and running an example private chaincode.

We assume that you are familiar with Hyperledger Fabric; otherwise we recommend the
[Fabric documentation](https://hyperledger-fabric.readthedocs.io/en/latest/getting_started.html)
as your starting point.
Moreover, we assume that you are familiar with the [Intel SGX SDK](https://github.com/intel/linux-sgx).


This README is structure as follows.
We start by [cloning the FPC repository](#clone-fabric-private-chaincode) and explain how to prepare your development environment for FPC in [Setup your FPC Development Environment](#setup-your-development-environment).
In [Build Fabric Private Chaincode](#build-fabric-private-chaincode) we guide you through the building process and elaborate on common issues.
Finally, we give you a starting point for [Developing with Fabric Private Chaincode](#developing-with-fabric-private-chaincode) by introducing the FPC Hello World Tutorial.

### Clone Fabric Private Chaincode

Clone the code and make sure it is on your `$GOPATH`. (Important: we assume in this documentation and default configuration that your `$GOPATH` has a _single_ root-directoy!)
We use `$FPC_PATH` to refer to the Fabric Private Chaincode repository in your filesystem.  
```bash
export FPC_PATH=$GOPATH/src/github.com/hyperledger/fabric-private-chaincode
git clone --recursive https://github.com/hyperledger/fabric-private-chaincode.git $FPC_PATH
```

## Setup your Development Environment

There are two different ways to develop Fabric Private Chaincode. 

### Option 1: Using the Docker-based FPC Development Environment
Using our preconfigured Docker container development environment. [Option 1](docs/setup-option1.md)

### Option 2: Setting up your system to do local development

As an alternative to the Docker-based FPC development environment you can install and manage all necessary software dependencies which are required to compile and run FPC. [Option 2](docs/setup-option2.md) 

## Build Fabric Private Chaincode

Once you have your development environment up and running (i.e., using our docker-based setup or install all dependencies on your machine) you can build FPC and start developing your own FPC application.
Note by default we build FPC with SGX simulation mode. For SGX hardware-mode support please also read the [Intel SGX Attestation Support](#intel-sgx-attestation-support) Section below. 

To build all required FPC components and run the integration tests run the following:
```bash
cd $FPC_PATH
make docker
make
 ```

Besides the default target, there are also following make targets:
- `build`: build all FPC build artifacts
- `docker`: build docker images 
- `test`: run unit and integration tests
- `clean`: remove most build artifacts (but no docker images)
- `clobber`: remove all build artifacts including built docker images
- `checks`: do license and linting checks on source

Also note that the file `config.mk` contains various defaults which
can all be redefined in an optional file `config.override.mk`.

See also [below](#building-documentation) on how to build the documentation.

### Intel SGX Attestation Support

To run Fabric Private Chaincode in hardware mode (secure mode), you need an SGX-enabled
hardware as well corresponding OS support.  However, even if you don't
have SGX hardware available, you still can run FPC in simulation mode by
setting `SGX_MODE=SIM` in your environment.
You can find more details [here](docs/build-sgx.md).

### FPC Playground for non-SGX environments

FPC leverages Intel SGX as the Confidential Computing technology to guard Fabric chaincodes.
Even though the Intel SGX SDK supports a simulation mode, where you can run applications in a simulated enclave, it still requires an x86-based platform to run and compile the enclave code.
Another limitation comes from the fact that the Intel SGX SDK is only available for Linux and Windows.

To overcome these limitations and allow developers to toy around with the FPC API, we provide two ways to getting started with FPC.

1) Using the [Docker-based FPC Development Environment](#setup-your-development-environment) (works well on x86-based platforms on Linux and Mac).
2) FPC builds without SGX SDK dependencies (targets x86/arm-based platforms on Linux and Mac).

We now elaborate on how to build the FPC components without the SGX SDK [here](docs/playground-nonsgx.md).
Note that this is indented for developing purpose only and does not provide any protection at all.

### Troubleshooting

This section elaborate on common issues with building Fabric Private Chaincode that you can read [here](docs/troubleshooting.md).

### Building Documentation

To build documentation (e.g., images from the PlantUML `.puml` files), you will have to install `java` and download `plantuml.jar`. Either put `plantuml.jar` into
in your `CLASSPATH` environment variable or override `PLANTUML_JAR` or `PLANTUML_CMD` in `config.override.mk`
(see `config.mk` for default definition of the two variables). Additionally, you will need the `dot` program from the
graphviz package (e.g., via `apt-get install graphviz` on Ubuntu).

By running the following command you can generate the documentation.
```bash
cd docs
make
```

## Developing with Fabric Private Chaincode

In the [samples](samples) folder you find a few examples how to develop applications using FPC and run them
on a Fabric network.
In particular, [samples/application](samples/application) contains examples of the FPC Client SDK for Go.
In [samples/chaincode](samples/chaincode) we give illustrate the use of the FPC Chaincode API;
and in [samples/deployment](samples/deployment) we show how to deploy and run FPC chaincode on the Fabric-samples test network and with K8s (minikube).

More details about FPC APIs in the [Reference Guides](#reference-guides) Section.

### Your first private chaincode

Create, build and test your first private chaincode with the [Hello World Tutorial](samples/chaincode/helloworld/README.md).

### Developing and deploying on Azure Confidential Computing

We provide a brief [FPC on Azure Tutorial](samples/deployment/azure/FPC_on_Azure.md) with the required steps to set up a confidential computing instance on Azure to develop and test FPC with SGX hardware mode enabled. 


## Reference Guides

You can find more details related to the Management API, FPC Shim and FPC client SDK [here](docs/referenceguides.md).


## Getting Help

Found a bug? Need help to fix an issue? You have a great idea for a new feature? Talk to us! You can reach us on
[Discord](https://discord.gg/hyperledger) in #fabric-private-chaincode.

We also have a weekly meeting every Tuesday at 3 pm GMT on [Zoom](https://zoom.us/my/hyperledger.community.3). Please
see the Hyperledger [community calendar](https://wiki.hyperledger.org/display/HYP/Calendar+of+Public+Meetings) for
details.

## Contributions Welcome

For more information on how to contribute to Fabric Private Chaincode please see our [contribution](CONTRIBUTING.md)
section.

## References

- Marcus Brandenburger, Christian Cachin, Rüdiger Kapitza, Alessandro
  Sorniotti: Blockchain and Trusted Computing: Problems, Pitfalls, and a
  Solution for Hyperledger Fabric. https://arxiv.org/abs/1805.08541

- [Fabric Private Chaincode RFC](https://github.com/hyperledger/fabric-rfcs/blob/main/text/0000-fabric-private-chaincode-1.0.md)

- Presentation at the Hyperledger Fabric contributor meeting
  August 21, 2019.
  Motivation, background and the inital architecture.
  [Slides](https://docs.google.com/presentation/d/1ewl7PcY9t27lScv2O2VaeHMsk13oe5B2MqU-qzDiR80)

- Presentation of at the Hyperledger Fabric contributor meeting
  November 11, 2020.
  The design and rationale for FPC Lite (FPC 1.0).
  [Slides](https://docs.google.com/presentation/d/1KX3_gB70H6PZw5uvYbIPYPOMt8qsh2nLRsGmXEf98Ls/edit#slide=id.ga89b65b885_0_0)


## Project Status

Hyperledger Fabric Private Chaincode was accepted via a Hyperledger Fabric [RFC](https://github.com/hyperledger/fabric-rfcs/blob/main/text/0000-fabric-private-chaincode-1.0.md) and is now under development.
Before, the project operated as a Hyperledger Labs project.
This code is provided solely to demonstrate basic Fabric Private Chaincode
mechanisms and to facilitate collaboration to refine the project architecture
and define minimum viable product requirements. The code provided in this
repository is prototype code and not intended for production use.

## License

Hyperledger Fabric Private Chaincode source code files are made
available under the Apache License, Version 2.0 (Apache-2.0), located in the
[LICENSE file](LICENSE).
