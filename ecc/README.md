<!---
Licensed under Creative Commons Attribution 4.0 International License
https://creativecommons.org/licenses/by/4.0/
--->
# Chaincode wrapper (ecc)

Before your continue here make sure you have build ``ecc_enclave`` before.
We refer to [ecc_enclave/README.md](../ecc_enclave). Otherwise, the build
might fail with the message `ecc_enclave build does not exist!`.

This is a go chaincode that is used to invoke the enclave. The chaincode logic
is implemented in C++ as enclave code and is loaded by by the go chaincode as
C library (``ecc/enclave/lib/enclave.signed.so``).  For more details on the
chaincode implementation see [ecc_enclave](../ecc_enclave).


The following steps guide you through the build phase. Make sure this project is on your `$GOPATH`.

First, build the chaincode and the validation plugin

    $ make

Note that when you are running the tests, you will observe some error messages
produced by the enclave ``ERROR [Enclave] VIOLATIONoh oh! cmac does not
match!``. Don't worry this is expected as the test does create a proper cmac on
the state data that is consumed by the chaincode when `getState` is called.

## Chaincode docker image

Fabric uses docker to deploy and run chaincodes in isolated containers. Normally, the peer creates the docker image
automatically when a chaincode is installed.  In particular, it fetches the source code, builds the chaincode
executable, and copies them into a new docker images based on the Fabric chaincode environment (`fabric-ccenv`) base
image.  Note that, since the peer is lazy, the docker image is not recreated if it already exists. The image name
comprise of the peer name, the chaincode name, and a hash.

However, we use a custom chaincode environment docker image that has SGX-support enabled.  As the default chaincode
deployment process does not support embedded libraries, which we need for the enclave chaincode, we will shortcut the
normal installation by overriding an existing chaincode image.

In detail, first install some `dummy` chaincode, which we are going to override, and check if the corresponding docker
image has been created successfully. For example:

    $ peer chaincode install -n ecc -v 0 -p github.com/hyperledger/fabric/examples/chaincode/go/example02/cmd 
    $ docker images
    REPOSITORY
    TAG                 IMAGE ID
    dev-jdoe-ecc-0-8bdbb434df41902eb2d2b2e2f10f6b0504b63f56eb98582f307c11a15fc14eb7
    latest              0c18434ae5e3

You can see that the peer has successfully created the docker image comprising the `dummy` chaincode.

Next, just run ``make docker`` to build and override the existing docker image with our private chaincode. To verify that
the image contains our enclave code, let's have a look inside the image and see if we can see an enclave folder.

    $ make docker
    $ docker run -i -t --entrypoint ls dev-jdoe-ecc-0-8bdbb434df41902eb2d2b2e2f10f6b0504b63f56eb98582f307c11a15fc14eb7:latest

You can also define the peer name and the chaincode name manually.

    $ make docker DOCKER_IMAGE=my-peername-ecc-0

For debugging you can also start the docker image.

    $ make docker-run

When developing a new chaincode it is useful to also wipe out the old docker image.

    $ make docker-clean 
