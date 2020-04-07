<!---
Licensed under Creative Commons Attribution 4.0 International License
https://creativecommons.org/licenses/by/4.0/
--->
# Trusted Ledger Chaincode (tlcc)

Before your continue here make sure you have built ``tlcc_enclave``.
We refer to [tlcc_enclave/README.md](../tlcc_enclave).

## Integrate with Fabric

TLCC is integrated into a fabric peer as system chaincode and thus must be built into the peer binary.
We provide a peer target in [fabric/](../fabric) that builds the peer with tlcc integration.

## Starting the peer

When starting the peer make sure that `LD_LIBRARY_PATH` points to the enclave lib.

    $ LD_LIBRARY_PATH=$FPC_PATH/tlcc/enclave/lib build/bin/peer node start

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
