# Reference Guides

Here you will find the FPC reference guides, including the management API, the FPC Shim, and the FPC Client SDK.

## Management API

While the management API for Fabric is mostly unchanged, some modifications are needed for FPC to work.
In particular, FPC extends the Fabric's lifecycle API with additional commands to create an FPC enclave and handle the key provisioning.
These are detailed separately in the **[FPC Management API document](design/fabric-v2%2B/fpc-management.md)**

## FPC Shim

The FPC Shim follows the programming model used in the standard Fabric Go shim and offers a C++ based FPC Shim to FPC chaincode developers. It currently comprises only a subset of the standard Fabric Shim and is complemented in the future.
These details are documented separately in the Shim header file itself: **[`../ecc_enclave/enclave/shim.h`](../ecc_enclave/enclave/shim.h)**

*Important*: The initial version of FPC, FPC 1.0 (aka FPC Lite), has a
few constraints in applicability and programming model.  Hence, study carefully the
[section discussing this in the FPC RFC](https://github.com/hyperledger/fabric-rfcs/blob/main/text/0000-fabric-private-chaincode-1.0.md#fpc-10-application-domain)
and the comments at the top of [`shim.h`](../ecc_enclave/enclave/shim.h)
before designing, implementing and deploying an FPC-based solution.
<!-- could also mention
	[FPC for Health use  case](https://docs.google.com/document/d/1jbiOY6Eq7OLpM_s3nb-4X4AJXROgfRHOrNLQDLxVnsc/)
-->


## FPC Client SDK

In order to interact with a FPC chaincode you can use the FPC Client SDK for Go or use the Peer CLI tool provided with FPC.
Both make FPC related client-side encryption and decryption transparent to the user, i.e., client-side programming is mostly standard Fabric and agnostic to FPC.

The FPC Client SDK for Go is located in [../client_sdk/go](../client_sdk/go). See also [Godocs](https://pkg.go.dev/github.com/hyperledger/fabric-private-chaincode/client_sdk/go/).

For the command-line invocations, use the **`$FPC_PATH/fabric/bin/peer.sh`** wrapper script. We refer to our integration tests for usage examples.
