<!---
Licensed under Creative Commons Attribution 4.0 International License
https://creativecommons.org/licenses/by/4.0/
--->
# SGX support for a Fabric peer

To enable SGX support for a Fabric peer start with a fresh copy of Fabric and
apply our patches. https://github.com/hyperledger/fabric

We assume that you are familiar with building Fabric manually; otherwise we highly
recommend to spend some time to build Fabric and run a simple network with a
few peers and a ordering service. In the examples, we also assume that you have
Fabric and Fabric Private Chaincode in the same directory-tree from $GOPATH.

If you are new to Fabric, we recommend the Fabric documentation as your starting point. You should start with
[installing](https://hyperledger-fabric.readthedocs.io/en/release-1.4/prereqs.html) Fabric dependencies and setting up
your [development environment](https://hyperledger-fabric.readthedocs.io/en/release-1.4/dev-setup/build.html).

## Patch and Build

Clone fabric and checkout the 1.4.1 release.

    $ git clone --branch v1.4.1 https://github.com/hyperledger/fabric.git $GOPATH/src/github.com/hyperledger/fabric
    $ cd $GOPATH/src/github.com/hyperledger/fabric
    $ git am ../../hyperledger-labs/fabric-private-chaincode/fabric/*.patch

When building the peer  make sure fabric is your ``$GOPATH`` and you enable the
plugin feature. Otherwise our custom validation plugins will
(silently!) ignored by the peer, despite the settings in ``core.yaml``.

    $ GO_TAGS=pluginsenabled make peer

For our sample scripts you will also have to build the orderer and configtxgen

	$ make orderer configtxgen

It is ok to also to build everything with ``make`` with absent
(default) target and/or with ``GO_TAGS``.  However, (a) make sure that
peer _always_ is built with GO_TAGS and (b) if you build the default
target be not surprised if unit tests fail ..

To make your life easier we have prepared an example configuration and an
auction demo in ``sgxconfig``.  Our example config contains the MSP
for a simple consortium and a bunch of scripts to run the auction demo.


## Run the Auction

Before you continue here build the main components of Fabric Private Chaincode by going through the section `Custom
chaincode environment docker image` and `Build the chaincode enclave and ledger enclave` in main [README](../README.md)

other components, such as the chaincode
enclave, ledger enclave, etc ...

Also, if you have run it previously _and_ changed ecc or ercc code, you will have to manually remove
the docker images with `(cd ../ecc; make docker-clean; cd ../ercc; make docker-clean)` or you will get
unexpected results running on (partially) stale code!

To run the demo you can use the scripts in
[sgxconfig/demo](sgxconfig/demo). Make sure that the scripts point to your
Fabric Private Chaincode directory. Note that for better demonstration
transaction arguments are in clear.

Before you start the ordering service and the peer you should create a channel
using the ``create_channel.sh`` script.  Next, start the ordering service and
the peer in two separate terminals using the corresponding scripts.  In a
third terminal, you can you run the auction demo with ``run_sgx_auction.sh``.
Please edit ``start_peer.sh`` and point LD_LIBRARY_PATH to the tlcc enclave lib.

Note that when you run ``run_sgx_auction.sh`` the first time, you may
see the following error:

    ....bin/peer chaincode instantiate -o localhost:7050 -C mychannel -n ecc -v 0 -c '{"args":["init"]}' -V ecc-vscc
    Error: could not assemble transaction, err Proposal response was not successful, error code 500, msg transaction returned with failure:
    Incorrect number of arguments. Expecting 4

Don't worry, that is OK! :) The short answer to resolve this is to just
rebuild ecc. Go to ``${GOPATH}/src/github.com/hyperledger-labs/fabric-private-chaincode/ecc`` and run
``make docker``.  You can, then, re-run ``run_sgx_auction.sh`` and the
error is gone.

The long answer is the following: When a new chaincode is installed, the
Fabric peer takes care of building the corresponding docker image that
is used to execute the chaincode.  As we need a custom SGX-enabled
environment to execute our chaincode inside an enclave, we need to tell
the peer to use our custom docker image.

* Terminal 1

        $ cd fabric/sgxconfig
        $ ./demo/create_channel.sh
        $ ./demo/start_orderer.sh

* Terminal 2

        $ cd fabric/sgxconfig
        $ ./demo/start_peer.sh

* Terminal 3

        $ cd fabric/sgxconfig
        $ ./demo/run_sgx_auction.sh
