# Architecture design

The main [README](../README.md) gives an high-level overview of the FPC architecture. Here we provide additional details.

More detailed architectural information and overview of the protocols can be found in the [Fabric Private Chaincode RFC](https://github.com/hyperledger/fabric-rfcs/blob/main/text/0000-fabric-private-chaincode-1.0.md).

The full detailed operation of FPC is documented in a series of UML
Sequence Diagrams. Note that FPC version 1.x corresponds to `FPC Lite`
in documents and code.

Specifically:

- The `fpc-lifecycle-v2`([puml](design/fabric-v2%2B/fpc-lifecycle-v2.puml)) diagram describes the normal lifecycle of a chaincode in FPC, focusing in particular on those elements that change in FPC vs. regular Fabric.
- The `fpc-registration`([puml](design/fabric-v2%2B/fpc-registration.puml)) diagram describes how an FPC Chaincode Enclave is created on a Peer and registered in the FPC Registry, including the Remote Attestation process.
- The `fpc-key-dist`([puml](design/fabric-v2%2B/fpc-key-dist.puml)) diagram describes the process by which chaincode-unique cryptographic keys are created and distributed among enclaves running identical chaincodes. Note that in the current version of FPC, key generation is performed, but the key distribution protocol has not yet been implemented.
- The `fpc-cc-invocation`([puml](design/fabric-v2%2B/fpc-cc-invocation.puml)) diagram illustrates the invocation process at the beginning of the chaincode lifecycle in detail, focusing on the cryptographic operations between the Client and Peer leading up to submission of a transaction for Ordering.
- The `fpc-cc-execution`([puml](design/fabric-v2%2B/fpc-cc-execution.puml)) diagram provides further detail of the execution phase of an FPC chaincode, focusing in particular on the `getState` and `putState` interactions with the Ledger.
- The `fpc-validation`([puml](design/fabric-v2%2B/fpc-validation.puml)) diagram describes the FPC-specific process of validation.
- The `fpc-components`([puml](design/fabric-v2%2B/fpc-components.puml)) diagram shows the important data structures of FPC components and messages exchanged between components.
- The detailed message definitions can be found as [protobufs](protos/fpc).
- The [interfaces document](design/fabric-v2%2B/interfaces.md) defines the interfaces exposed by the FPC components and their internal state.

Additional Google documents provide details on FPC 1.0:

- The [FPC for Health use case](https://docs.google.com/document/d/1jbiOY6Eq7OLpM_s3nb-4X4AJXROgfRHOrNLQDLxVnsc/) describes how FPC 1.0 enables a health care use case.
  The document also gives more details on the FPC 1.0-enabled application domains and related constraints. Lastly, it provides a security analysis why these constraints are sufficient for security.
- The [FPC externalized endorsement validation](https://docs.google.com/document/d/1RSrOfI9nh3d_DxT5CydvCg9lVNsZ9a30XcgC07in1BY/) describes the FPC 1.0 enclave endorsement validation mechanism.

