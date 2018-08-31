# SGX support for a Fabric peer

To enable SGX support for a Fabric peer start with a fresh copy of Fabric and
apply our patch. https://github.com/hyperledger/fabric

We assume that you are familiar with build Fabric manually otherwise we
recommend to spend some time to build Fabric and run a simple network with a
few peers and a ordering service.

## Patch and Build

Clone fabric and checkout the 1.2 release.

    $ git clone https://github.com/hyperledger/fabric.git
    $ git checkout release-1.2
    $ git apply path-to-this-patch/sgx_support.patch

When building the peer make sure fabric is your ``GOPATH`` and you enable the
plugin feature. Otherwise our custom validation plugins can not be loaded.

    $ GO_TAGS=pluginsenabled EXPERIMENTAL=false DOCKER_DYNAMIC_LINK=true make peer

To make your life easier we have prepared an example configuration and an
auction demo. You can copy ``sgxconfig`` to your fabric directory and modify
the sgx section in ``core.yaml`` accordingly. In particular, grep for all
``/path-to/fabric-secure-chaincode/`` and replace with the correct path.  Our
example config contains the MSP for a simple consortium and a bunch of scripts
to run the auction demo.

### IAS

In order to use Intel's Attestation Service (IAS) you can register
[here](https://software.intel.com/en-us/sgx).  Place your client certificate
and your SPID in the ``ias`` folder.

## Run the Auction

Before you continue here build the other components, such as the chaincode
enclave, ledger enclave, etc ...

To run the demo you can use the scripts in
[sgxconfig/demo](sgxconfig/demo). Make sure that the scripts point to your
fabric-secure-chaincode directory. Note that for better demonstration
transaction arguments are in clear. 

Before you start the ordering service and the peer you should create a channel
using the ``create_channel.sh`` script.  Next, start the ordering service and
the peer in two separate terminals using the corresponding scripts.  In a
third terminal, you can you run the auction demo with ``run_sgx_auction.sh``.
Please edit ``start_peer.sh`` and point LD_LIBRARY_PATH to the tlcc enclave lib.

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
