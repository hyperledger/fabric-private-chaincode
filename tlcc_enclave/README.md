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

We use *nanopb*, a lightweight implementation of Protocol Buffers, inside the ledger enclave to parse blocks of
transactions. Install nanopb by following the instruction below. For more detailed information consult the official
nanopb documentation http://github.com/nanopb/nanopb.

    $ git clone https://github.com/nanopb/nanopb.git ~/nanopb
    $ cd ~/nanopb/
    $ git checkout nanopb-0.3.9.2
    $ cd generator/proto && make


Next copy `pb.h`, ``pb_encode.*``, ``pb_decode.*`` and ``pb_common.*`` to
``common/protobuf/`` directory in the root folder.

    $ mkdir -p common/protobuf
    $ cp ~/nanopb/pb* common/protobuf 

Now we can generate the proto files by using ``generate_protos.sh``. Check that
you export the following variables pointing to Fabric and nanopb.

    export FABRIC_PATH=/path-to/fabric/
    export NANOPB_PATH=/path-to/nanopb/

and run it.

    $ ./generate_protos.sh

## Build

We use cmake to build tlcc_enclave.

    $ mkdir build 
    $ cd build
    $ cmake ../.
    $ make

## Test

    $ make test

## Deploy

    $ make deploy

## Debugging

Run gdb

    $ make
    $ LD_LIBRARY_PATH=$LD_LIBRARY_PATH:./ sgx-gdb test_runner
    > enable sgx_emmt
    > r
Note that OPENSSL sometimes complains, here you can just continue debugging.
