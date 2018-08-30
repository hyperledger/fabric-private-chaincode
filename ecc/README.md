# Chaincode wrapper (ecc)

This is a go chaincode that is used to invoke the enclave. The chaincode logic
is implemented in C++ as enclave code and is loaded by by the go chaincode as
C library (``CCEnclave/lib/enclave.signed.so``).  For more details on the
chaincode implementation see ecc_encalve/.

## Getting started

The following steps guide you through the build phase. Make sure this project is on your GOPATH.

First, build the chaincode and the validiation plugin

    $make
    
Next, build the chaincode docker image. Note that normally the fabric peer
itself creates the docker image when a new chaincode is installed. However, we
use a custom chaincode environment docker image that has SGX support enabled.
You can define the peer name and the chaincode name using PEER_NAME and
CC_NAME.

    $make docker

For debugging you can also start the docker image.

    $make docker-run
