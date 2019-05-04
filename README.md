# Hyperledger Fabric Secure Chaincode Execution

This lab enables Secure Chaincode Execution using Intel SGX for Hyperledger
Fabric.

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

This lab provides a framework to develop and execute Fabric chaincode within
an enclave.  Furthermore, Fabric extensions for chaincode enclave registration
and transaction verification are provided.

This lab proposes an architecture to enable Secure Chaincode Execution using
Intel SGX for Hyperledger Fabric as presented and published in the paper:

* Marcus Brandenburger, Christian Cachin, Rüdiger Kapitza, Alessandro
  Sorniotti: Blockchain and Trusted Computing: Problems, Pitfalls, and a
  Solution for Hyperledger Fabric. https://arxiv.org/abs/1805.08541

We provide an initial proof-of-concept implementation of the proposed
architecture. Note that the code provided in this repository is prototype code
and not meant for production use! The main goal of this lab is to discuss and
refine the proposed architecture involving the Hyperledger community.

# Architecture and components

This lab extends a Fabric peer with the following components: A chaincode
enclave that executes a particular chaincode and a ledger enclave that enables
all chaincode enclaves to verify the blockchain state integrity; all run
inside SGX. In the untrusted part of the peer, an enclave registry maintains
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
   encryption, attestation, and secure blockchain state access. This component
   is devided into two subcomponents: ``ecc_enclave`` contains the code
   running inside an enclave and ``ecc`` contains a wrapper chaincode that
   invokes the enclave.

1. *Ledger enclave:* The ledger enclave maintains the ledger in an enclave in
   the form of integrity-specific metadata representing the most recent
   blockchain state at the peer. It performs the same validation steps as the
   peer when a new block arrives, but additionally generates a cryptographic
   hash of each key-value pair of the blockchain state and stores it within
   the enclave. The ledger enclave exposes an interface to the chaincode
   enclave for accessing the integrity-specific metadata. This is used to
   verify the correctness of the data retrieved from the blockchain
   state. Like the chaincode enclave, the ledger enclave is divided into two
   subcomponents: ``tlcc`` and ``tlcc_enclave``.

1. *Enclave registry:* The enclave registry is a chaincode that runs outside
   SGX and maintains a list of all existing chaincode enclaves in the
   network. It performs attestation with the chaincode enclave and stores the
   attestation result on the blockchain. The attestation demonstrates that a
   specific chaincode executes in an actual enclave. This enables the peers
   and the clients to inspect the attestation of a chaincode enclave before
   invoking chaincode operations or committing state changes. The enclave
   registry (``ercc``) comes with a custom validation plugin (ercc-vscc).

1. *Enclave transaction validator:* The enclave transaction validator
   (``ecc/vscc``) complements the peer’s validation system and is responsible
   for validating transactions produced by a chaincode enclave. In particular,
   the enclave transaction validator checks that a transaction contains a
   valid signature issued by a registered chaincode enclave. If the validation
   is successful, it marks the transactions as valid and hands it over to the
   ledger enclave, which crosschecks the decision before it finally commits
   the transaction to the ledger.

# Getting started

The following steps guide you through the build phase and configuration, for
deploying and running an example chaincode.

## Requirements

* CMake 3.5.1 or higher
* Go 1.11.x or higher
* Hyperledger Fabric v1.4.1 https://github.com/hyperledger/fabric

    If you are new to Fabric, we recommend the Fabric documentation as your starting point. You should start with
    [installing](https://hyperledger-fabric.readthedocs.io/en/release-1.4/prereqs.html) Fabric dependencies and setting
    up your [development environment](https://hyperledger-fabric.readthedocs.io/en/release-1.4/dev-setup/build.html).

* SGX SDK v2.4 or v2.5 for Linux https://github.com/intel/linux-sgx
* SSL for SGX SDK v2.4.1 https://github.com/intel/intel-sgx-ssl (we recommend using OpenSSL 1.1.0j)
* (for hardware-mode SGX build) credentials for IAS, read [here](fabric#intel-attestation-service-ias)

### SGX SDK and SSL

The the chaincode envlave and the trusted ledger enclave require SGX SSL.  See
README in project root for more details. Intel SGX SSL
https://github.com/intel/intel-sgx-ssl

After installing the SGX SDK and SGX SSL double check that ``SGX_SDK`` and
``SGX_SSL`` variables are set correctly in your env. In particular,
the makefiles are configured with meaningful defaults for variables
like ``SGX_SDK``, ``SGX_ARCH``, ``SGX_MODE``, ``SGX_BUILD`` or
``SGX_SSL``.  However, if you have non-standard values for install
location or would like build with debug and/or simulator mode, you can
override the default values by defining a corresponding environment
variable, e.g., `export SGX_MODE=SIM SGX_BUILD=DEBUG
SGX_SSL=/my/sgx/ssl/install/`.


## Working from behind a proxy

The current code should work behind a proxy assuming
  * you have defined the corresponding environment variables (i.e.,
  `http_proxy`, `https_proxy` and, potentially, `no_proxy`) properly
  defined, and
  * docker (daemon & client) is properly set up for proxies as
    outlined in the Docker documentation for [clients](https://docs.docker.com/network/proxy/) and the [daemon](https://docs.docker.com/config/daemon/systemd/#httphttps-proxy).
If you run Ubuntu 18.04, make sure you run docker 18.09 or later. Otherwise you will run into problems with DNS resolution inside the container.

Another problem you might encounter when running the sample scripts
insofar that some '0.0.0.0' in ``fabric/sgxconfig/core.yaml`` used by
clients -- e.g., the peer CLI using the ``address: 0.0.0.0:7051`` config
as part of the ``peer`` section -- result in the client being unable
to find the server. The likely error you will see is
 ``err: rpc error: code = Unavailable desc = transport is closing``.
In that case, you will have to replace the '0.0.0.0' with a concrete
ip address such as '127.0.0.1'.

## Time to get the code

Checkout the code and make sure it is on your ``GOPATH``.
(Important: we assume in this documentation and default configuration
that your GOPATH has a _single_ root-directoy!)

    $ git clone https://github.com/hyperledger-labs/fabric-secure-chaincode.git $GOPATH/src/hyperledger-labs/fabric-secure-chaincode

## Prepare your Fabric

First we need to enable a Fabric peer to excute chaincode using our
project. In [fabric](fabric) you find a details described how to patch fabric
and build the peer.

## Custom chaincode environment docker image

In [utils/fabric-ccenv-sgx](utils/fabric-ccenv-sgx) you can find instructions
to create a custom fabric-ccenv docker image that is required to execute a
chaincode within an enclave.

## Build the chaincode enclave and ledger enclave

Next build the chaincode enclave [ecc_enclave](ecc_enclave) and the ledger
enclave [tlcc_enclave](tlcc_enclave). Follow the instructions in the
corresponding directories.

In the next step we need to integrate the enclave code into a Fabric
chaincode.  Please follow the instructions in [ecc](ecc) for the chaincode
enclave and [tlcc](tlcc) for the ledger enclave.

In order to run and deploy a chaincode enclave we need to build the enclave
registry. See [ercc](ercc).

Now we have all components we need to run the example auction chaincode in an enclave.


# References

- Marcus Brandenburger, Christian Cachin, Rüdiger Kapitza, Alessandro
  Sorniotti: Blockchain and Trusted Computing: Problems, Pitfalls, and a
  Solution for Hyperledger Fabric. https://arxiv.org/abs/1805.08541

# Project Status
Hyperledger Fabric Secure Chaincode operates as a Hyperledger Labs project.
This code is provided solely to demonstrate basic Fabric Secure Chaincode
mechanisms and to facilitate collaboration to refine the project architecture
and define minimum viable product requirements. The code provided in this
repository is prototype code and not intended for production use.

# Initial Committers
- [Marcus Brandenburger](https://github.com/mbrandenburger) (bur@zurich.ibm.com)
- [Christian Cachin](https://github.com/cca88) (cca@zurich.ibm.com)
- [Rüdiger Kapitza](https://github.com/rrkapitz) (kapitza@ibr.cs.tu-bs.de)
- [Alessandro Sorniotti](https://github.com/ale-linux) (aso@zurich.ibm.com)

# Sponsor
[Gari Singh](https://github.com/mastersingh24) (garis@us.ibm.com)

# License
Hyperledger Fabric Secure Chaincode Execution source code files are made
available under the Apache License, Version 2.0 (Apache-2.0), located in the
[LICENSE file](LICENSE).
