# Trusted ledger enclave (tlcc_enclave)

The ledger enclave maintains the ledger in an enclave in the form of
integrity-specific metadata representing the most recent blockchain state. It
performs the same validation steps as the peer when a new block arrives, but
additionally generates a cryptographic hash of each key-value pair of the
blockchain state and stores it within the enclave. The ledger enclave exposes
an interface to the chaincode enclave for accessing the integrity-specific
metadata. This is used to verify the correctness of the data retrieved from
the blockchain state.

## Start with generating proto parser

We use nanopb, a lightweight implementation of Protocol Buffers.
Install nanopb by following the instruction on http://github.com/nanopb/nanopbopy and copy pb_encode.c, pb_decode.c and pb_common.c to common/protobuf/ directory.

Set fabric path and nanpb path in ``compile_protos.sh`` 

    FABRIC=/path-to/fabric/
    NANOPB_PATH=/path-to/nanopb

and run it.

    $ ./compile_protos.sh

## SGX SSL

The trusted ledger enclave requires SGXSSL. See README in project root
fore more details. Intel SGX SSL https://github.com/intel/intel-sgx-ssl

## Build

We use cmake to build tlcc_enclave.

    $ mkdir build 
    $ cd build
    $ cmake ../.
    $ make

## Test

    make test

## Debugging

Run gdb

    $ make
    $ LD_LIBRARY_PATH=$LD_LIBRARY_PATH:./ sgx-gdb test_runner
    > enable sgx_emmt
    > r
Note that OPENSSL sometimes complains, here you can just continue debugging.
