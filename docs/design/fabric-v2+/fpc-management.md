# FPC Management

This document defines the management APIs of FPC as part of the `peer` commandline tool.
Such definition builds on the design diagrams describing the enclave creation, 
the enclave registration and the chaincode key generation.

## Admin Commands

### Extended Fabric v2 Lifecycle Chaincode Commands

These commands extend the original Fabric v2 commands to handle FPC chaincodes.
For a description of the original Fabric v2 commands, check out the [Fabric v2 documentation](https://hyperledger-fabric.readthedocs.io/en/release-2.3/commands/peerlifecycle.html).
In the following, this document describes additional flags and usages.
Any other commands, which does not appear in this list, remains unchanged in FPC.

#### `package`

This commands has a new option `--lang fpc-c` to specify the FPC chaincode type.

#### `approveformyorg`

This command has the following requirements *when it is used for FPC chaincodes*:
* `--version <MRENCLAVE as (upper-case) hexadecimal string>`, the version of the FPC chaincode must contain a string that represent the identity of the enclave. This is the same identity which the trusted hardware computes and attests to. The string can be conveniently found in the `mrenclave` output file of the `generate_mrenclave.sh` script.
* `--signature-policy string`, the endorsement policy must be a valid FPC endorsement policy.
FPC currently supports only a single enclave as endorser, running at a designated peer. See more details below in the [FPC Endorsement Policies](#fpc-endorsement-policies) section.
* `--endorsement-plugin string`, this flag is not supported in FPC.
* `--validation-plugin string`, this flag is not supported in FPC.

#### `checkcommitreadiness`

This command has the same requirements *when it is used for FPC chaincodes* as described above for `approveformyorg`.

#### `commit`

This command has the same requirements *when it is used for FPC chaincodes* as described above for `approveformyorg`.


### Create Chaincode Enclave

```peer lifecycle chaincode initEnclave -n <chaincode id> --peerAddresses <ip addr:port> --sgx-credentials <path>```

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


## Admin Client SDK

In addition to the CLI based admin commands there exists also a FPC management API for Go.

```go
func LifecycleInitEnclave(channelId string, req LifecycleInitEnclaveRequest, options ...resmgmt.RequestOption) (fab.TransactionID, error)
```

See the details of the API in [godoc](https://pkg.go.dev/github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/client/resmgmt)
and an example of its use in  [`integration/client_sdk/go/utils.go`](../../../integration/client_sdk/go/utils.go).

### Key Distribution

Key distribution commands are not supported in the initial version of FPC.
Users must aware that each FPC chaincode runs in a single enclave at a designated peer.


## FPC Endorsement Policies

In FPC, signatures by enclaves are the base of endorsements.
For the initial version of FPC, FPC Lite, an endorsement of a single
endorsing peer is sufficient.
More specifically, the endorsement must stem from an enclave which is properly registered with the enclave registiry (ercc) as part of a successful `initEnclave` call (see [above](#create-chaincode-enclave).
The corresponding *enclave endorsement policy* is called [`designated enclave`](https://docs.google.com/document/d/1RSrOfI9nh3d_DxT5CydvCg9lVNsZ9a30XcgC07in1BY/edit) and is *implicitly defined*.
As an FPC chaincode is protected by the use of Trusted Execution Environments,
from a security perspective,
the integrity guarantees gained from a single enclave are similar to the guarantee provided in a typical Fabric system with multiple peer endorsements.

With the [externalized enclave endorsement validation](https://docs.google.com/document/d/1RSrOfI9nh3d_DxT5CydvCg9lVNsZ9a30XcgC07in1BY/)
of the initial version of FPC,
there is though also separate endorsement required,
namely an endorsement that the enclave endorsement was properly validated.
The policy is called the *validation endorsement policy* and should match the standard organizational trust model in Fabric,
e.g., a majority of involved organizations.


For FPC Lite, the validation endorsement policy is specified through the
[`approveformyorg`](#approveformyorg) and [`commit`](#commit)
commands, while the enclave endorsement policy is implicit,
as mentioned above.
For example, to define require any 2 organizations for a consortium of `Org1`, `Org2` and `Org3` are required for endorsement validation,
you might install fpc chaincode as follows:
```bash
	..
    peer lifecycle chaincode approveformyorg --channelID mychannel name myfpc --signature-policy "OutOf(2, 'Org1.peer', 'Org2.peer', Org3.peer')" ...
	..
    peer lifecycle chaincode commit --channelID mychannel --name myfpc --signature-policy "OutOf(2, 'Org1.peer', 'Org2.peer', Org3.peer')" ...
	..
```
Future releases of FPC might provide support for richer enclave endorsement policies and/or enclave endorsement validation internalized.
Correspondingly, specification of enclave endorsement policies and enclave validation endorsement policies will likely change in the future.

