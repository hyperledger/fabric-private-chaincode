# Fabric Smart Client

Fabric Private Chaincode integrates with the [Fabric Smart Client](https://github.com/hyperledger-labs/fabric-smart-client) to build complex distributed-applications while simplify the development and testing.
In particular, with Fabric Smart Client, it becomes easy to build and test prototypes and proof-of-concepts without the need to deploy a Fabric network. 
That is, developers can focus on the development of the FPC chaincode.

The Fabric Smart Client integrates the FPC Client API to interact with a FPC Chaincode and protect the invocation arguments.
Moreover, the deployment process of a FPC Chaincode is integrated in the test network suite provided by Fabric Smart Client.
The FPC developer just packages the FPC Chaincode as a docker image and points to it in the Fabric network definition. 

## Getting started

The Fabric Smart Client (FSC) repository contains a neat tutorial how FSC can be used with FPC.
It shows how to setup a Fabric network, deploy the FPC echo chaincode (see [/samples/chaincode/echo](../../chaincode/echo)), and invoke it from a FSC view.
You can find the tutorial [here](https://github.com/hyperledger-labs/fabric-smart-client/tree/main/integration/fabric/fpc/echo).

## More advanced example

In [/samples/demos/irb](../../demos/irb) we give a more advanced example of how the Fabric Smart Client is used to build a complex demo with Fabric Private Chaincode. 