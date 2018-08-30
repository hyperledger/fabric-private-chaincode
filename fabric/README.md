# SGX support for a Fabric peer

To enable SGX support for a Fabric peer take a fresh copy of Fabric and apply
our patch.

    $ git clone https://github.com/hyperledger/fabric.git
    $ git checkout release-1.2
    $ git apply path-to-this-patch/sgx_support.patch

When building the peer make sure you enable the plugin feature.

    $ GO_TAGS=pluginsenabled EXPERIMENTAL=false DOCKER_DYNAMIC_LINK=true make

To make your life easier we have prepared an example configuration and an
auction demo. You can copy sgxconfig to your fabric directory and modify the
sgx section in core.yaml accordingly.

To run the demo you can use the scripts in sgxconfig/demo. Make sure that the
scripts point to your fabric-secure-chaincode directory. Note that for better
demonstration transaction arguments are in clear.

Before you start the ordering service and the peer you must create a channel
using the create_channel.sh script.  Next, start the ordering service and the
peer in two separate terminals using the corresponding scripts.  In a third
terminal, you can you run the auction demo with run_sgx_auction.sh.
