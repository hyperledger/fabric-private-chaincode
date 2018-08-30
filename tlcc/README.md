# Trusted Ledger Chaincode (tlcc)

Before your continue here make sure you have build tllib before. We
refer to tllib/README.md

First build tlcc as system chaincode plugin. What is that? See
https://hyperledger-fabric.readthedocs.io/en/latest/systemchaincode.html

    $make

This build `tlcc.so`. Using this plugin when running the peer inside
Docker most problby will not work out-of-the-bock, thus, not supported
right now.

## Integrate with fabric

Add tlcc as system chaincode plugin to your `core.yaml`. Example:

```
chaincode:
    system:
        tlcc: enable

    systemPlugins:
        - enabled: true
        name: tlcc
        path: /your/file/system/sgx-cc/tlcc/tlcc.so
        invokableExternal: true
        invokableCC2CC: true
```

## Start the peer

Make sure `LD_LIBRARY_PATH` points to the enclave lib.

    $LD_LIBRARY_PATH=/your/file/system/sgx-cc/tlcc/enclave/lib bin/node start

## Join the channel

This prototype currently supports a single channel. Start with using
`configtxgen` to create a new channel and let your peer join it. Next,
call tlcc (mis)using query operation to join the channel. See example
below.

    $bin/peer channel create -o localhost:7050 -c mychannel -f mychannel.tx
    $bin/peer channel join -b mychannel.block
    $bin/peer chaincode query -n tlcc -c '{"Args": ["JOIN_CHANNEL", "mychannel"]}' -C mychannel

Your trusted ledger should be up and running now.


