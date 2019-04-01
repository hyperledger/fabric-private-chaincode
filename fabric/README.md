# SGX support for a Fabric peer

To enable SGX support for a Fabric peer start with a fresh copy of Fabric and
apply our patch. https://github.com/hyperledger/fabric

We assume that you are familiar with building Fabric manually; otherwise we highly
recommend to spend some time to build Fabric and run a simple network with a
few peers and a ordering service.

## Patch and Build

Clone fabric and checkout the 1.4 release.

    $ git clone https://github.com/hyperledger/fabric.git
    $ git checkout release-1.4
    $ git apply path-to-this-patch/sgx_support.patch

When building the peer make sure fabric is your ``GOPATH`` and you enable the
plugin feature. Otherwise our custom validation plugins can not be loaded.

    $ GO_TAGS=pluginsenabled make peer

To make your life easier we have prepared an example configuration and an
auction demo. You can copy ``sgxconfig`` to your fabric directory and modify
the sgx section in ``core.yaml`` accordingly. In particular, grep for all
``/path-to/fabric-secure-chaincode/`` and replace with the correct path.  Our
example config contains the MSP for a simple consortium and a bunch of scripts
to run the auction demo.

### Intel Attestation Service (IAS)

The requirements are:
* your certificate registered with IAS
* the private key associated to your certificate
* your Service Provider ID (SPID)

In order to use Intel's Attestation Service (IAS) you need to register
with Intel. [Here](https://software.intel.com/en-us/articles/code-sample-intel-software-guard-extensions-remote-attestation-end-to-end-example)
you can find more details on how to obtain a signed client certificate,
registering it and get a SPID.

Place your client certificate and your SPID in the ``ias`` folder.

    cp client.crt /path-to/fabric/sgxconfig/ias/client.crt
    cp client.key /path-to/fabric/sgxconfig/ias/client.key
    echo 'YOURSPID' | xxd -r -p > /path-to/fabric/sgxconfig/ias/spid.txt

We currently make use of `unlinkable signatures` for the attestation, thus, when registering with the IAS please choose
unlinkable signatures.  In the case you prefer linkable attestation,
e.g., because you already have linkable IAS EPID credentials, change
in line 217 of [../ecc_enclave/sgxcclib/sgxcclib.c](../ecc_enclave/sgxcclib/sgxcclib.c)
the constant `SGX_UNLINKABLE_SIGNATURE` to `SGX_LINKABLE_SIGNATURE`,
re-compile and re-deplay [ecc_enclave](../ecc_enclave#build) and [ecc](../ecc#getting-started),
configure your IAS settings as above with your linkable credentials and run the auction example as follows.

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

Note that when you run ``run_sgx_auction.sh`` the first time, you may
see the following error:

    ../.build/bin/peer chaincode instantiate -o localhost:7050 -C mychannel -n ecc -v 0 -c '{"args":["init"]}' -V ecc-vscc
    Error: could not assemble transaction, err Proposal response was not successful, error code 500, msg transaction returned with failure:
    Incorrect number of arguments. Expecting 4

Don't worry, that is OK! :) The short answer to resolve this is to just
rebuild ecc. Go to ``path-to/fabric-secure-chaincode/ecc`` and run
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
