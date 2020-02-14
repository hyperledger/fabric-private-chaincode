<!---
Licensed under Creative Commons Attribution 4.0 International License
https://creativecommons.org/licenses/by/4.0/
--->
# Spectrum Auction Chaincode

The code in this folder implements a demo Spectrum Auction by heavily building on [documents](https://wireless.fcc.gov/auctions/1002/resources/fABS_tutorial_final/presentation_html5.html) of the Federal Communications Commission (FCC). *This demo is for demonstration purposes only, and not meant for production use.*

The purpose of this demo chaincode is to showcase the privacy enhancements that the Fabric Private Chaincode project brings to Hyperledger Fabric by using Intel(R) SGX. To run such demo, please see the documents to set up the Hyperledger Fabric Network and the User Interface.
The present document only provides references to the auction specification, and instructions to build the auction chaincode and run its tests.

A [specification document](https://docs.google.com/document/d/1YUF4mzzuybzWk3fbXbTANWO8-tr757BP85qcwE7gQdk) is available for additional information about APIs, input and output messages.

The chaincode implements a simple randomized version of the Assignment phase of the auction. Hence, it does not implement the `submitAssignmentBid` API. In particular, the Assignment phase is terminated immediately, with no bids. Then, the assignment is performed randomly. Finally, the phase, and so the auction itself, is closed.

## Build the code

* below assumes you have set the `FPC_PATH` environment variable to the root folder of the Fabric Private Chaincode project

### Locally
* run `make` to build locally

### Using docker
* Build the Auction Chaincode
```
cd $FPC_PATH/demo/chaincode/fpc
make docker-build
```
**NOTE** If you already built the `hyperledger/fabric-private-chaincode-cc-builder`
image, it will use that image to build the chaincode. Otherwise it will build
the [`cc-builder` image](../../../utils/cc-builder/Dockerfile) and the
[`dev` image](../../../utils/dev/Dockerfile) before building the chaincode.

## Test the code
**NOTE** If you used docker to build the chaincode, make sure you run the
following using the [development docker image](../../../utils/docker/dev/Dockerfile).

The auction chaincode can be conveniently developed and tested by using the [FPC Mock Server](../../client/backend/mock) as follows:
* build the auction chaincode
* follow the instruction in the Mock Server [Readme](../../client/backend/mock/README.md) file to build the server and make sure the server can access the compiled auction artifacts
* run `./test.sh`
