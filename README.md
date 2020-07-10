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

This lab provides a framework to develop and execute Fabric chaincode within
an enclave.  It allows to write chaincode applications where the data is
encrypted on the ledger and can only be accessed in clear by authorized
parties. Furthermore, Fabric extensions for chaincode enclave registration
and transaction verification are provided.

This lab proposes an architecture to enable private chaincode execution using
Intel SGX for Hyperledger Fabric as presented and published in the paper:

* Marcus Brandenburger, Christian Cachin, Rüdiger Kapitza, Alessandro
  Sorniotti: Blockchain and Trusted Computing: Problems, Pitfalls, and a
  Solution for Hyperledger Fabric. https://arxiv.org/abs/1805.08541

We provide an initial proof-of-concept implementation of the proposed
architecture. Note that the code provided in this repository is prototype code
and not meant for production use! The main goal of this lab is to discuss and
refine the proposed architecture involving the Hyperledger community.

For up to date information about our community meeting schedule, past
presentations, and info on how to contact us please refer to our
[wiki page](https://wiki.hyperledger.org/display/fabric/Hyperledger+Fabric+Private+Chaincode).

## Architecture and components

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
   registry (``ercc``) comes with a custom validation plugin (``ercc-vscc``).

1. *Enclave transaction validator:* The enclave transaction validator
   (``ecc/vscc``) complements the peer’s validation system and is responsible
   for validating transactions produced by a chaincode enclave. In particular,
   the enclave transaction validator checks that a transaction contains a
   valid signature issued by a registered chaincode enclave. If the validation
   is successful, it marks the transactions as valid and hands it over to the
   ledger enclave, which crosschecks the decision before it finally commits
   the transaction to the ledger.

More detailed architectural information can be found in the [Fabric Private Chaincode RFC](https://github.com/mbrandenburger/fabric-rfcs/blob/fpc-rfc/text/0000-fabric-private-chaincode.md) <!-- TODO: fix-me with pointer to version in fabric-rfcs once RFC is accepted ... -->

The full detailed operation of FPC is documented in a series of UML
Sequence Diagrams. Specifically:

- The `fpc-lifecycle-v2`([puml](docs/design/fabric-v2%2B/fpc-lifecycle-v2.puml)/[png](docs/design/fabric-v2%2B/fpc-lifecycle-v2.png)) diagram describes the normal lifecycle of a chaincode in FPC, focusing in particular on those elements that change in FPC vs. regular Fabric.
- The `fpc-registration`([puml](docs/design/fabric-v2%2B/fpc-registration.puml)/[png](/design/fabric-v2%2B/fpc-registration.png)) diagram describes how an FPC Chaincode Enclave is created on a Peer and registered in the FPC Registry, including the Remote Attestation process.
- The `fpc-key-dist`([puml](docs/design/fabric-v2%2B/fpc-key-dist.puml),[png](docs/design/fabric-v2%2B/fpc-key-dist.ng))) diagram describes the process by which chaincode-unique cryptographic keys are created and distributed among enclaves running identical chaincodes. Note that in the current version of FPC, key generation is performed, but the key distribution protocol has not yet been implemented.
- The `fpc-cc-invocation`([puml](docs/design/fabric-v2%2B/fpc-cc-invocation.puml)/[png](docs/design/fabric-v2%2B/fpc-cc-invocation.png)) diagram illustrates the invocation process at the beginning of the chaincode lifecycle in detail, focusing on the cryptographic operations between the Client and Peer leading up to submission of a transaction for Ordering.
- The `fpc-cc-execution`([puml](docs/design/fabric-v2%2B/fpc-cc-execution.puml)/[png](docs/design/fabric-v2%2B/fpc-cc-execution.png)) diagram provides further detail of the execution phase of an FPC chaincode, focusing in particular on the `getState` and `putState` interactions with the Ledger and verification of state with the Ledger Enclave.
- The `fpc-validation`([puml](docs/design/fabric-v2%2B/fpc-validation.puml)/[png](docs/design/fabric-v2%2B/fpc-validation.png)) diagram describes the FPC-specific process of validation and revalidation using the Ledger Enclave.
- The `fpc-components`([puml](docs/design/fabric-v2%2B/fpc-components.puml)/[png](docs/design/fabric-v2%2B/fpc-components.png)) diagram shows the important data structures of FPC components and messages exchanged between components.

Note: for the `png`-files, you will have to build the documentation as described [below](#building-documentation).



## Releases

- [Tech Preview - July 2, 2020](https://github.com/hyperledger-labs/fabric-private-chaincode/tree/concept-release-2.0)

- [Concept Release - March 2, 2020](https://github.com/hyperledger-labs/fabric-private-chaincode/tree/concept-release-1.0)

*WARNING: This project is in continous development and the `master`
 branch will not always be stable. Unless you want to actively
 contribute to the project itself, we advice you to use one of above releases*



## Getting started

The following steps guide you through the build phase and configuration, for
deploying and running an example private chaincode.

We assume that you are familiar with building Fabric manually; otherwise we highly recommend to spend some time to build
Fabric and run a simple network with a few peers and a ordering service. We recommend the Fabric documentation as your
starting point. You should start with
[installing](https://hyperledger-fabric.readthedocs.io/en/release-2.0/prereqs.html) Fabric dependencies and setting up
your [development environment](https://hyperledger-fabric.readthedocs.io/en/release-2.0/dev-setup/build.html).

Moreover, we assume that you are familiar with the Intel SGX SDK.

### Intel SGX

To run Fabric Private Chaincode in secure mode, you need an SGX-enabled
hardware as well corresponding OS support.  However, even if you don't
have SGX hardware available, you still can run FPC in simulation mode by
setting `SGX_MODE=SIM` in your environment.

Note that the simulation mode is for developing purpose only and does
not provide any security guarantees.

Notice: by default the project builds in hardware-mode SGX, ``SGX_MODE=HW`` as defined in `<absolute-project-path>/fabric-private-chaincode/config.mk` and you can
explicitly opt for building in simulation-mode SGX, ``SGX_MODE=SIM``. In order to set non-default values for install
location, or for building in simulation-mode SGX, you can create the file `<absolute-project-path>/fabric-private-chaincode/config.override.mk` and override the default
values by defining the corresponding environment variable.

If you run SGX in __simulation mode only__, you can skip below
sections and jump right away to [Setup your Development
Environment](#setup-your-development-environment).

Note that you can always come back here when you want a setup with SGX
hardware-mode later after having tested with simulation mode.


#### Install SGX Operation System Support

Install the Intel SGX software stack for Linux (including the SGX
driver and the SGX Platform Software (PSW)) by following the official
[documentation](https://github.com/intel/linux-sgx).


#### Register with Intel Attestation Service (IAS)

We currently support EPID-based attestation and  use the Intel's
Attestation Service to perform attestation with chaincode enclaves.

What you need:

* a Service Provider ID (SPID)
* the (primary) api-key associated with your SPID

In order to use Intel's Attestation Service (IAS), you need to register
with Intel. On the [IAS EPID registration page](https://api.portal.trustedservices.intel.com/EPID-attestation)
you can find more details on how to register and obtain your SPID plus corresponding api-key.

We currently support both `linkable` and `unlinkable` signatures for the attestation.
The type of attestation used is selected based on the `FPC_ATTESTATION_TYPE` environment variable:
`epid_unlinkable` for unlinkable or `epid_linkable` for linkable signatures. If you 
do not define that environment variable, the chosen attestation method is `epid_unlinkable`.
Note that a mismatch between your IAS credentials and the linkable setting
will result in an (HTTP) error '400' visible in the log-files when the
code tries to verify the attestation. (Another cause for such error '400'
could a mismatch between provided SPID and api key as specified below).

Place your ias api key and your SPID in the ``ias`` folder as follows:

    echo 'YOUR_API_KEY' > ${GOPATH}/src/github.com/hyperledger-labs/fabric-private-chaincode/config/ias/api_key.txt
    echo 'YOUR_SPID' > ${GOPATH}/src/github.com/hyperledger-labs/fabric-private-chaincode/config/ias/spid.txt

### Setup your Development Environment

There are 2 different ways to develop Fabric Private Chaincode. Using our preconfigured Docker container development environment or setting up your local system with all required software dependencies to build and develop chaincode locally. 

### Option 1: Using the Docker-based FPC Development Environment

As standard Fabric, we require docker to run chaincode.  We recommend
to set privileges to manage docker as a non-root user. See the
official docker [documentation](https://docs.docker.com/install/linux/linux-postinstall/)
for more details.

We provide a docker image containing the FPC development environment. This will enable you to get a quick start to
get FPC running.

First make sure your host has
* A running Docker daemon compatible with docker provided by Ubuntu
  18.04, currently `Docker version 18.09`.  It also should use
  `/var/run/docker.sock` as socket to interact with the daemon (or you
  will have to override in `<absolute-project-path>/fabric-private-chaincode/config.override.mk` the default definition in make of `DOCKER_DAEMON_SOCKET`)
* GNU make

A few notes:
* if you run behind a proxy, you will have to configure the proxy,
  e.g., for docker (`~/.docker/config.json`). See
  [Working from behind a proxy](#working-from-behind-a-proxy) below for more information.
* if your local host is SGX enabled, i.e., there is a device `/dev/sgx` or
  `/dev/isgx` and your PSW daemon listens to `/var/run/aesmd`, then the docker image will be sgx-enabled and your settings from `./config/ias` will be used. You will have to manually set `SGX_MODE=HW` before building anything to use HW mode.
* if you want additional apt packages in your container image, add to the `<absolute-project-path>/fabric-private-chaincode/config.override.mk` file in the fabric-private-chaincode directory. In that file, define `DOCKER_DEV_IMAGE_APT_ADD__PKGS` with a
  list of packages you want. They will then be automatically added to the docker image
* due to the way the peer's port for chaincode connection is managed,
  you will be able to run only a single FPC development container on a
  particular host.

To build the docker image, run

    $ cd utils/docker; make dev

and then use it with

    $ cd utils/docker; make run

This will open a shell inside the FPC development container, with
environment variables like GOPATH appropriately defined and all
dependencies like fabric built, ready to build and run FPC.

Note that by default the dev container mounts your local cloned FPC project as a volume to `/project/src/github.com/hyperledger-labs/fabric-private-chaincode` within the docker container.
This allows you to edit the content of the repository using your favorite editor in your system and the changes inside the docker container. Additionally, you are also not loosing changes inside the container when you reboot or the container gets stopped for other reasons.

Optional: to do a clean build do the following within the container
```
<docker-root>:project/src/github.com/hyperledger-labs/fabric-private-chaincode# make clean
<docker-root>:project/src/github.com/hyperledger-labs/fabric-private-chaincode# make build
```
Now you are ready to start development.  Go to the [Develop Your First Private Chaincode
](#your-first-private-chaincode) section.

### Option 2: Setting up your system to do local development

#### Requirements

Make sure that you have the following required dependencies installed:

* Linux (OS) (we recommend Ubuntu 18.04, see [list](https://github.com/intel/linux-sgx#prerequisites) supported OS)

* CMake v3.5.1 or higher

* [Go](https://golang.org/) 1.14.1 or higher

* Docker 18.x

* Protocol Buffers 
    - Protocol Buffers 3.0.x needed for the Intel SGX SDK
    - Protocol Buffers 3.11.x or higher and [Nanopb](http://github.com/nanopb/nanopb) 0.3.9.2

* SGX SDK v2.6 for [Linux](https://github.com/intel/linux-sgx)

* Credentials for Intel Attestation Service, read [here](#intel-attestation-service-ias) (for hardware-mode SGX)

* [SSL](https://github.com/intel/intel-sgx-ssl)  for SGX SDK v2.4.1 (we recommend using OpenSSL 1.1.0j) 

* Hyperledger [Fabric](https://github.com/hyperledger/fabric/tree/v2.1.1) v2.1.1

* Clang-format 6.x or higher

* jq
* hex (for ubuntu, found in package basez)

* A recent version of [PlantUML](http://plantuml.com/), including Graphviz, for building documentation. See [Documentation](#building-documentation) for our recommendations on installing. The version available in common package repositories may be out of date.

#### Intel SGX SDK and SSL

Fabric Private Chaincode requires the Intel [SGX SDK](https://github.com/intel/linux-sgx) and
[SGX SSL](https://github.com/intel/intel-sgx-ssl) to build the main components of our framework and to develop and build
your first private chaincode.     

Install the Intel SGX software stack for Linux by following the
official [documentation](https://github.com/intel/linux-sgx). Please make sure that you use the
SDK version as denoted above in the list of requirements. 

For SGX SSL, just follow the instructions on the [corresponding
github page](https://github.com/intel/intel-sgx-Sal). In case you are
building for simulation mode only and do not have HW support, you
might also want to make sure that [simulation mode is set](https://github.com/intel/intel-sgx-ssl#available-make-flags)
when building and installing it.

Once you have installed the SGX SDK and SSL for SGX SDK please double check that ``SGX_SDK`` and ``SGX_SSL`` variables
are set correctly in your environment. 


#### Protocol Buffers

We use *nanopb*, a lightweight implementation of Protocol Buffers, inside the ledger enclave to parse blocks of 
transactions. Install nanopb by following the instruction below. For this you need a working Google Protocol Buffers
compiler with python bindings (e.g. via `apt-get install protobuf-compiler python-protobuf libprotobuf-dev`). 
For more detailed information consult the official nanopb documentation http://github.com/nanopb/nanopb. 

    $ export NANOPB_PATH=/path-to/install/nanopb/
    $ git clone https://github.com/nanopb/nanopb.git ${NANOPB_PATH}
    $ cd ${NANOPB_PATH}
    $ git checkout nanopb-0.3.9.2
    $ cd generator/proto && make

Make sure that you set `$NANOPB_PATH` as it is needed to build Fabric Private Chaincode.

Moreover, in order to build Fabric protobufs we also require a newer Protobuf compiler than what is provided as standard ubuntu package and is used to build the
Intel SGX SDK. For this reason you will have to download and install another version and use it together with Nanopb. Do not install the new protobuf, though, such that it is not found in your standard PATH but instead define the `PROTOC_CMD`, either as environment variable or via `config.override.mk` to point to the new `protoc` binary

    $ wget https://github.com/protocolbuffers/protobuf/releases/download/v3.11.4/protoc-3.11.4-linux-x86_64.zip
    $ unzip protoc-3.11.4-linux-x86_64.zip -d /usr/local/proto3
    $ export PROTOC_CMD=/usr/local/proto3/bin/protoc


#### Clone Fabric Private Chaincode

Clone the code and make sure it is on your `$GOPATH`. (Important: we assume in this documentation and default configuration that your `$GOPATH` has a _single_ root-directoy!)

    $ git clone --recursive https://github.com/hyperledger-labs/fabric-private-chaincode.git $GOPATH/src/github.com/hyperledger-labs/fabric-private-chaincode

#### Prepare Fabric

Fabric Private Chaincode support requires to re-build the Fabric peer. 

Checkout Fabric 2.1.1 release using the following commands:

    $ export FABRIC_PATH=${GOPATH}/src/github.com/hyperledger/fabric
    $ git clone https://github.com/hyperledger/fabric.git $FABRIC_PATH
    $ cd $FABRIC_PATH; git checkout tags/v2.1.1
    
Note that Fabric Private Chaincode currently does not work with the Fabric `master` branch, therefore make sure you use the Fabric
`v2.1.1` tag. This is important as our build script applies some patches
to the fabric peer to enable FPC support. Make sure the source of Fabric is in your ``$GOPATH``.

### Build Fabric Private Chaincode

Once you have your development environment up and running (i.e., using our docker-based setup or install all dependencies on your machine) you can build FPC and start developing your own FPC application.

```
$ cd $GOPATH/src/github.com/hyperledger-labs/fabric-private-chaincode
$ make
 ```
This will build all required components and run the integration tests.

#### Building individual components

In [utils/fabric-ccenv-sgx/](utils/fabric-ccenv-sgx) you can find instructions
to create a custom fabric-ccenv docker image that is required to execute a
chaincode within an enclave.

The chaincode enclave [ecc_enclave](ecc_enclave) and the ledger
enclave [tlcc_enclave](tlcc_enclave) can be built manually.
Follow the instructions in the corresponding directories.

For the integration of the enclave code into a Fabric
chaincode, please follow the instructions in [ecc/](ecc) for the chaincode
enclave and [tlcc/](tlcc) for the ledger enclave.

In order to run and deploy a chaincode enclave we need to build the enclave
registry. See [ercc/](ercc).

Both, ecc and ercc use custom validation to verify transactions. In [fabric/](fabric) you can find
instructions how to build the FPC-enabled peer that uses our custom validation.

Moreover, we provide a set of integration tests in [integration/](integration) to demonstrate Fabric Private Chaincode
capabilities.

#### Trouble shooting

##### Docker

Building the project requires docker. We do not recommend to run `sudo make`
to resolve issues with mis-configured docker environments as this also changes your `$GOPATH`. Please see hints on
[docker](#docker) installation above.

The makefiles do not ensure that docker files are always rebuild to
match the latest version of the code in the repo.  If you suspect you
have an issue with outdated docker images, you can run `make clobber
build` which forces a rebuild.  It also ensures that all other
downlad, build or test artifacts are scrubbed from your repo and might
help overcoming other problems. Be advised that that the rebuild can
take a fair amount of time.

##### Working from behind a proxy

The current code should work behind a proxy assuming
  * you have defined the corresponding environment variables (i.e.,
    `http_proxy`, `https_proxy` and, potentially, `no_proxy`) properly, and
  * docker (daemon & client) is properly set up for proxies as
    outlined in the Docker documentation for
    [clients](https://docs.docker.com/network/proxy/) and the
    [daemon](https://docs.docker.com/config/daemon/systemd/#httphttps-proxy).
If you run Ubuntu 18.04, make sure you run docker 18.09 or later. Otherwise you will run into problems with DNS resolution inside the container.

You will also require a recent version of docker-compose. In particular, the docker-compose from ubuntu 18.04
(docker-compose 1.17) is _not_ recent enough to understand `~/.docker/config.json` and related proxy options.
To upgrade, install a recent version following the instructions from [docker.com](https://docs.docker.com/compose/install/), e.g.,
for version 1.25.4 execute
```
 sudo curl -L "https://github.com/docker/compose/releases/download/1.25.4/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
 sudo chmod +x /usr/local/bin/docker-compose
```
Furthermore, for docker-compose networks to work properly with proxies, the `noProxy`
variable in your `~/.docker/config.json` should at least contain `127.0.0.1,127.0.1.1,localhost,.org1.example.com,.example.com`.

Another problem you might encounter when running the integration tests 
insofar that some '0.0.0.0' in ``integration/config/core.yaml`` used by
clients -- e.g., the peer CLI using the ``address: 0.0.0.0:7051`` config
as part of the ``peer`` section -- result in the client being unable
to find the server. The likely error you will see is
 ``err: rpc error: code = Unavailable desc = transport is closing``.
In that case, you will have to replace the '0.0.0.0' with a concrete
ip address such as '127.0.0.1'.


##### Environment settings

Our build system requires a few variables to be set in your environment. Missing variables may cause `make` to fail. 
Below you find a summary of all variables which you should carefully check and add to your environment. 

```bash
# Path to your SGX SDK and SGX SSL
export SGX_SDK=/opt/intel/sgxsdk
export SGX_SSL=/opt/intel/sgxssl

# Path to nanopb
export NANOPB_PATH=$HOME/nanopb

# SGX simulation mode
export SGX_MODE=SIM

# SGX simulation mode
export SGX_MODE=HW

# The attestation type is ignored when SGX_MODE=SIM is set.

# IAS attestation (unlinkable)
export FPC_ATTESTATION_TYPE=epid_unlinkable

# IAS attestation (linkable)
export FPC_ATTESTATION_TYPE=epid_linkable

```
##### Clang-format

Some users may experience problems with clang-format. In particular, the error `command not found: clang-format` 
appears even after installing it via `apt-get install clang-format`. See [here](https://askubuntu.com/questions/1034996/vim-clang-format-clang-format-is-not-found)
for how to fix this.  

##### ERCC setup failures

If, e.g., running the integration tests executed when you run `make`,
you get errors of following form:

```
Error: endorsement failure during invoke. response: status:500 message:"Setup failed: Can not register enclave at ercc: Error while retrieving attestation report: IAS returned error: Code 401 Access Denied"
```

In case you run in SGX HW mode, check that your files in `config/ias`
are set properly as explained in [Section Intel Attestation Service
(IAS)](#intel-attestation-service-ias).  Note that if you run
initially in simulation mode and these files do not exist, the build
will create dummy files. In case you switch later to HW mode without
configuring these files correctly for HW mode, this will result in
above error.

### Building Documentation

To build documentation (e.g., images from the PlantUML `.puml` files), you will have to install `java` and download `plantuml.jar`. Either put `plantuml.jar` into
in your `CLASSPATH` environment variable or override `PLANTUML_JAR` or `PLANTUML_CMD` in `config.override.mk`
(see `config.mk` for default definition of the two variables). Additionally, you will need the `dot` program from the
graphviz package (e.g., via `apt-get install graphviz` on ubuntu).

By running the following command you can generate the documentation.

    $ cd docs
    $ make



## Developing with Fabric Private Chaincode

### Your first private chaincode
Create, build and test your first private chaincode with this [tutorial](examples/README.md).

### A Complete Application
For an end-to-end application demonstrating the potential of FPC, check out our [Clock Auction Demo Application](demo/README.md).


## Reference Guides

### Management API

While the management API for Fabric is mostly unchanged, some modifications are needed for FPC to work.
In particular, FPC extends the Fabric's lifecycle API with additional commands to create an FPC enclave and handle the key provisioning.
These are detailed separately in the [FPC Management API document](docs/design/fabric-v2%2B/fpc-management.md)

### FPC Shim

The FPC Shim follows the programming model used in the standard Fabric Go shim and offers a C++ based FPC Shim to FPC chaincode developers. It currently comprises only a subset of the standard Fabric Shim and is complemented in the future.
These details are documented separately in the Shim header file itself: **[ecc_enclave/enclave/shim.h](ecc_enclave/enclave/shim.h)**

### FPC Client-side SDK
FPC related encryption and decryption client-side is transparent to the user, i.e., client-side programming is standard Fabric and agnostic to FPC.


## Getting Help

Found a bug? Need help to fix an issue? You have a great idea for a new feature? Talk to us! You can reach us on
[RocketChat](https://chat.hyperledger.org/) in #fabric-private-chaincode. 

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

-  [Fabric Private Chaincode RFC](https://github.com/mbrandenburger/fabric-rfcs/blob/fpc-rfc/text/0000-fabric-private-chaincode.md) <!-- TODO: fix-me with pointer to version in fabric-rfcs once RFC is accepted ... -->

- Presentation at the Hyperledger Fabric contributor meeting August 21, 2019.
  Slides: https://docs.google.com/presentation/d/1ewl7PcY9t27lScv2O2VaeHMsk13oe5B2MqU-qzDiR80 

## Project Status

Hyperledger Fabric Private Chaincode operates as a Hyperledger Labs project.
This code is provided solely to demonstrate basic Fabric Private Chaincode
mechanisms and to facilitate collaboration to refine the project architecture
and define minimum viable product requirements. The code provided in this
repository is prototype code and not intended for production use.


## Initial Committers

- [Marcus Brandenburger](https://github.com/mbrandenburger) (bur@zurich.ibm.com)
- [Christian Cachin](https://github.com/cca88) (cca@zurich.ibm.com)
- [Rüdiger Kapitza](https://github.com/rrkapitz) (kapitza@ibr.cs.tu-bs.de)
- [Alessandro Sorniotti](https://github.com/ale-linux) (aso@zurich.ibm.com)


## Sponsor

[Gari Singh](https://github.com/mastersingh24) (garis@us.ibm.com)


## License

Hyperledger Fabric Private Chaincode source code files are made
available under the Apache License, Version 2.0 (Apache-2.0), located in the
[LICENSE file](LICENSE).
