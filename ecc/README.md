# Chaincode wrapper (ecc)

This is a go chaincode that is used to invoke the enclave. The chaincode logic
is implemented in C++ as enclave code and is loaded by by the go chaincode as
C library (``ecc/enclave/lib/enclave.signed.so``).  For more details on the
chaincode implementation see ecc_encalve/.

## Getting started

The following steps guide you through the build phase. Make sure this project is on your GOPATH.

First, build the chaincode and the validation plugin

    $ make

Next, build the chaincode docker image that is used by a fabric peer
to run our chaincode.  Normally, the peer creates the docker image
automatically when a new chaincode is installed.  In particular, it
fetches the source code, builds the chaincode binary, and copies them
into a new docker images based on fabric-ccenv.  Note that, since the
peer is lazy, the docker image is only created when the chaincode is
installed and it is not already existing.  The image name comprise of
the peer name, the chaincode name, and a hash.

However, we use a custom chaincode environment docker image that has
SGX-support enabled.  In order to tell a peer to use our SGX chaincode
image, we need to override an existing chaincode image.

For example: ``dev-jdoe-ecc-0-8bdbb434df41902eb2d2b2e2f10f6b0504b63f56eb98582f307c11a15fc14eb7``

Therefore, first install some chaincode, which we are going to override,
and check if the corresponding docker image has been created
successfully.

    $ peer chaincode install -n ecc -v 0 -p github.com/hyperledger/fabric/examples/chaincode/go/example02/cmd 
    $ docker images
    REPOSITORY
    TAG                 IMAGE ID
    dev-jdoe-ercc-0-a5a84629692f2ed6e111c44bd91e8c3e0906deb39d9e16f7acd5aefc51303184
    latest              7a5ea0677404
    dev-jdoe-ecc-0-8bdbb434df41902eb2d2b2e2f10f6b0504b63f56eb98582f307c11a15fc14eb7
    latest              0c18434ae5e3

Next, just run ```make docker`` to override the existing docker image with
our SGX chaincode. To verify that the image contains our enclave
code, let's have a look inside the image and see if we can see an
enclave folder.

    $ make docker
    $ docker run -i -t --entrypoint ls dev-jdoe-ecc-0-8bdbb434df41902eb2d2b2e2f10f6b0504b63f56eb98582f307c11a15fc14eb7:latest
    chaincode  chaintool  enclave  node  npm  npx  protoc-gen-go

You can define the peer and the chaincode name also manually.

    $ make docker DOCKER_IMAGE=my-peername-ecc-0

For debugging you can also start the docker image.

    $ make docker-run
