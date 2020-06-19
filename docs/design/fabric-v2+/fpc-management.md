# FPC Management

This document defines the management APIs of FPC.
Such definition builds on the design diagrams describing the enclave creation, 
the enclave registration and the chaincode key generation.

[//]: # (## Enclave and Chaincode APIs)

[//]: # (### Create Chaincode Enclave)

[//]: # (### Register Chaincode Enclave)

[//]: # (### Generate FPC Chaincode Keys)

## Admin Commands

### Extended Fabric v2 Lifecycle Chaincode Commands

These commands extend the original Fabric v2 commands to handle FPC chaincodes.
For a description of the original Fabric v2 commands, check out the Fabric v2 documentation.
In the following, this document describes additional flags and usages.
Any other commands, which does not appear in this list, remains unchanged in FPC.

#### Package

This commands has a new option `--lang fpc-c` to specify the FPC chaincode type.

#### Approveformyorg

This command has the following requirements *when it is used for FPC chaincodes*:
* `--version <MRENCLAVE>`, the version of the FPC chaincode must contain a string that represent the identity of the enclave. This is the same identity which the trusted hardware computes and attests to.
* `--signature-policy string`, the endorsement policy must be a valid FPC endorsement policy.
FPC currently supports only a single enclave as endorser, running at a designated peer.
* `--endorsement-plugin string`, this flag is not supported in FPC.
* `--validation-plugin string`, this flag is not supported in FPC.

#### Checkcommitreadiness

This command has the following requirements *when it is used for FPC chaincodes*:
* `--version <MRENCLAVE>`, the version of the FPC chaincode must contain a string that represent the identity of the enclave. This is the same identity which the trusted hardware computes and attests to.
* `--signature-policy string`, the endorsement policy must be a valid FPC endorsement policy.
FPC currently supports only a single enclave as endorser, running at a designated peer.
* `--endorsement-plugin string`, this flag is not supported in FPC.
* `--validation-plugin string`, this flag is not supported in FPC.

#### Commit

This command has the following requirements *when it is used for FPC chaincodes*:
* `--version <MRENCLAVE>`, the version of the FPC chaincode must contain a string that represent the identity of the enclave. This is the same identity which the trusted hardware computes and attests to.
* `--signature-policy string`, the endorsement policy must be a valid FPC endorsement policy.
FPC currently supports only a single enclave as endorser, running at a designated peer.
* `--endorsement-plugin string`, this flag is not supported in FPC.
* `--validation-plugin string`, this flag is not supported in FPC.


### Create Chaincode Enclave

This command performs the following operations:
* the creation of a new chaincode enclave,
which generates its enclave-specific cryptographic keys and produces a hardware-based attestation;
* the registration of the enclave's credentials on the Enclave Registry (chaincode);
* the generation within the enclave of chaincode-specific cryptographic keys,
which are then registered on the Enclave Registry (chaincode).

A successful command execution returns `0`,
indicating that the chaincode enclave is ready to endorse transaction proposals.

```peer lifecycle chaincode createenclave -n <chaincode id>```

### Key Distribution

Key distribution commands are not supported in the initial version of FPC.
Users must aware that each FPC chaincode runs in a single enclave at a designated peer.

