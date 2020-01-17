<!---
Licensed under Creative Commons Attribution 4.0 International License
https://creativecommons.org/licenses/by/4.0/
--->
# Spectrum Auction Chaincode

The code in this folder implements a demo Spectrum Auction by heavily building on [documents](https://wireless.fcc.gov/auctions/1002/resources/fABS_tutorial_final/presentation_html5.html) of the Federal Communications Commission (FCC). *This demo is for demonstration purposes only, and not meant for production use.*

The purpose of this demo chaincode is to showcase the privacy enhancements that the Fabric Private Chaincode project brings to Hyperledger Fabric by using Intel(R) SGX. To run such demo, please see the documents to set up the Hyperledger Fabric Network and the User Interface.
The present document only provides references to the auction specification, and instructions to build the auction chaincode and run its tests.

A [specification document](https://docs.google.com/document/d/1YUF4mzzuybzWk3fbXbTANWO8-tr757BP85qcwE7gQdk) is available for additional information about APIs, input and output messages.


## Build the code

* make sure the the `FPC_PATH` environment variable is set to the root folder of the Fabric Private Chaincode project
* run `make`

## Test the code

The auction chaincode can be conveniently developed and tested by using the [FPC Mock Server](../../client/backend/mock) as follows:
* build the auction chaincode
* follow the instruction in the Mock Server [Readme](../../client/backend/mock/README.md) file to build the server and make sure the server can access the compiled auction artifacts
* run `./test.sh`

