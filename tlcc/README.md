# Trusted Ledger Chaincode (tlcc)

Before your continue here make sure you have build ``tlcc_enclave`` before.
We refer to [tlcc_enclave/README.md](../tlcc_enclave).

First we build tlcc as system chaincode plugin. What is that? See
https://hyperledger-fabric.readthedocs.io/en/latest/systemchaincode.html

    $ make

This builds `tlcc.so`. Using this plugin when running the peer inside
Docker most probably will not work out-of-the-box, thus, not supported
right now.

## Integrate with Fabric

Add `tlcc` as system chaincode plugin to your `core.yaml`. Example:

```
chaincode:
    system:
        tlcc: enable

    systemPlugins:
        - enabled: true
        name: tlcc
        path: /path-to/fabric-private-chaincode/tlcc/tlcc.so
        invokableExternal: true
        invokableCC2CC: true
```

## Starting the peer

When starting the peer make sure that `LD_LIBRARY_PATH` points to the enclave lib.

    $ LD_LIBRARY_PATH=/path-to/fabric-private-chaincode/tlcc/enclave/lib .build/bin/node start

## Join the channel

This prototype currently supports a single channel only. Start with using
`configtxgen` to create a new channel and let your peer join it. Next,
call tlcc (mis)using query operation to join the channel. See example
below.

    $ bin/peer channel create -o localhost:7050 -c mychannel -f mychannel.tx
    $ bin/peer channel join -b mychannel.block
    $ bin/peer chaincode query -n tlcc -c '{"Args": ["JOIN_CHANNEL", "mychannel"]}' -C mychannel

Your trusted ledger should be up and running now.


## Demo

We have prepared an auction demo script available in [fabric/sgxconfig](../fabric/sgxconfig/demo).
See `start_peer.sh` and `run_sgx_auction.sh` as an example.