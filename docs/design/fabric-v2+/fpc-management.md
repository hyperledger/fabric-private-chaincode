# FPC Management

This document defines the management APIs of FPC as part of the `peer` commandline tool.
Such definition builds on the design diagrams describing the enclave creation, 
the enclave registration and the chaincode key generation.

## Admin Commands

### Extended Fabric v2 Lifecycle Chaincode Commands

These commands extend the original Fabric v2 commands to handle FPC chaincodes.
For a description of the original Fabric v2 commands, check out the [Fabric v2 documentation](https://hyperledger-fabric.readthedocs.io/en/release-2.1/commands/peerlifecycle.html).
In the following, this document describes additional flags and usages.
Any other commands, which does not appear in this list, remains unchanged in FPC.

#### Package

This commands has a new option `--lang fpc-c` to specify the FPC chaincode type.

#### Approveformyorg

This command has the following requirements *when it is used for FPC chaincodes*:
* `--version <MRENCLAVE as hexadecimal string>`, the version of the FPC chaincode must contain a string that represent the identity of the enclave. This is the same identity which the trusted hardware computes and attests to. The string can be conveniently found in the `mrenclave` output file of the `generate_mrenclave.sh` script.
* `--signature-policy string`, the endorsement policy must be a valid FPC endorsement policy.
FPC currently supports only a single enclave as endorser, running at a designated peer. See more details below in the [FPC Endorsement Policies](#fpc-endorsement-policies) section.
* `--endorsement-plugin string`, this flag is not supported in FPC.
* `--validation-plugin string`, this flag is not supported in FPC.

#### Checkcommitreadiness

This command has the same requirements *when it is used for FPC chaincodes* as described above for `approveformyorg`.

#### Commit

This command has the same requirements *when it is used for FPC chaincodes* as described above for `approveformyorg`.


### Create Chaincode Enclave

```peer lifecycle chaincode createenclave -n <chaincode id>```

This command performs the following operations:
* the creation of a new chaincode enclave,
which generates its enclave-specific cryptographic keys and produces a hardware-based attestation;
* the registration of the enclave's credentials on the Enclave Registry (chaincode);
* the generation within the enclave of chaincode-specific cryptographic keys,
which are then registered on the Enclave Registry (chaincode).

The command requires that the FPC chaincode definition is already committed on the channel.
Therefore, the command must follow the `lifecycle chaincode commit` command.
The admin can preliminarly check the successfully committed definition through
the `lifecycle chaincode querycommitted` command.

A successful command execution returns `0`,
indicating that the chaincode enclave is ready to endorse transaction proposals.


### Key Distribution

Key distribution commands are not supported in the initial version of FPC.
Users must aware that each FPC chaincode runs in a single enclave at a designated peer.


## FPC Endorsement Policies

The initial version of FPC restricts the set of allowed endorsement policies and currently
supports a designated endorser policy only. That is, only a single endorsing peer is
responsible to execute transactions of an FPC chaincode for the consortium.  As an FPC
chaincode is protected by the use of Trusted Execution Technology, from a security
perspective, the integrity guarantees gained from the endorsement model are similar to the guarantee provided in a typical Fabric system with multiple Peer endorsements.
However, the designated endorser policy does not provide resilience in terms of availability and limits a particular chaincode's scalability.
In future releases of FPC, we will address this limitation and relax this restriction by
providing support for rich endorsement policies.

To achieve the designated-peer endorsement policy, the consortium initially selects an organization (e.g., SampleOrg)
that will host the designated endorsing peer.  The corresponding endorsement policy
needs to be set to `OR('SampleOrg.peer')` and approved by the consortium using normal chaincode lifecycle agreement.
Other expressions like `AND` or `OR` with multiple organizations are currently not allowed and
may lead to unexpected behavior.

Note that in order to use `SampleORG.peer`, `NodeOUs.Enable` must be set to `true`, otherwise
`OR('SampleOrg.member')` must be used.

For example:

    peer lifecycle chaincode approveformyorg --channelID mychannel --name myfpc --signature-policy "OR('SampleOrg.peer')" ...

An admin of SampleOrg is then responsible to invoke `lifecycle chaincode createenclave` at
a single peer and thereby determine the designated endorser for the `myfpc` chaincode.  Clients
of the consortium have to make sure that they invoke transactions only at the designed endorser.
