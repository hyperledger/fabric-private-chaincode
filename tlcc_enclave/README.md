<!---
Licensed under Creative Commons Attribution 4.0 International License
https://creativecommons.org/licenses/by/4.0/
--->
# Trusted ledger enclave (tlcc_enclave)

The ledger enclave maintains the ledger in an enclave in the form of
integrity-specific metadata representing the most recent blockchain state. It
performs the same validation steps as the peer when a new block arrives, but
additionally generates a cryptographic hash of each key-value pair of the
blockchain state and stores it within the enclave. The ledger enclave exposes
an interface to the chaincode enclave for accessing the integrity-specific
metadata. This is used to verify the correctness of the data retrieved from
the blockchain state.

## Protos

We use nanopb for proto buffers inside the trusted ledger enclave.
We can generate the proto files by using ``make protos``.  Please check that
you export the environment variables `NANOPB_PATH`. Fabric proto files
are located in `protos/fabric` and fetched from https://github.com/hyperledger/fabric-protos.

## Build

We use cmake to build tlcc_enclave. For convince, you can just invoke ``make``
to trigger the cmake build.

    $ make

## Test

    $ make test

## Debugging

Run gdb

    $ make
    $ LD_LIBRARY_PATH=$LD_LIBRARY_PATH:./ sgx-gdb test_runner
    > enable sgx_emmt
    > r

Note that OPENSSL sometimes complains, here you can just continue debugging.

## Cleaning

    $ make clean
