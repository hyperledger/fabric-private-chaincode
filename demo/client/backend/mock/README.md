# FPC Mock Server

This mock server allows to run a FPC chaincode without a Fabric peer and exposes the chaincode interface via REST.

The goal of this tool is to simplify FPC chaincode development and enabled quick testing of FPC chaincodes without
deploying on Fabric network. 

## Start with fpc chaincode

The mock server uses ECC to load and interact with an FPC chaincode. First make sure you have build FPC.  

    $ cd $FPC_PATH
    $ make clean && make
    $ cd $FPC_PATH/ecc
    $ make sym-links

Next build your FPC chaincode.
    
    $ cd $FPC_PATH/examples/YOUR_CHAINCODE
    $ make

We are almost there ...
   
    $ cd $FPC_PATH/demo/client/backend/mock
    $ export LD_LIBRARY_PATH=${LD_LIBRARY_PATH:+"${LD_LIBRARY_PATH}:"}${FPC_PATH}/ecc_enclave/_build/lib
    $ ln -s $FPC_PATH/examples/YOUR_CHAINCODE/_build/ enclave

Finally, we can build the mock server running our FPC chaincode
      
    $ make run-fpc

### Additional notes

When running inside FPC dev docker container, please make sure that you run `make godeps` in `$FPC_PATH` before starting the mock server.

### Building manually

In case you want to build and start the mock server by hand you can do that easily by following these instructions. 

    $ go build -tags fpc
    $ ./mock

By default the mock server starts on port 3000. If you need to start the server on a different port you can can do it by running the following command:

    $ ./mock --port=8080

## Usage

Once the server is up and running it can receive invoke and query requests and forward them to the chaincode.

Example:

    curl -H "Content-Type: application/json" -H "x-user:fake-user" -X POST -d '{"tx":"someChaincodeFunction","args":["arg1", "arg2"]}' http://localhost:3000/api/cc/invoke
    curl -H "Content-Type: application/json" -H "x-user:fake-user" -X POST -d '{"tx":"anotherChaincodeFunction","args":["arg1", "arg2"]}' http://localhost:3000/api/cc/query

### Logging

In order to get FPC Chaincode Debug logging you need to compile the chaincode with `SGX_BUILD=DEBUG`.

Good luck!
