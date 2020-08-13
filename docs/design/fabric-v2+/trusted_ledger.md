# Trusted Ledger (TLCC)

The purpose of TLCC is to establish a trusted view of the ledger (inside an enclave)
and enable ECC to verify ledger state receiveed from the (untrusted) peer.

## Requirements

- ECC can query integrity metadata to validate ledger state received from the peer.
- TLCC must be synronized with peer's ledger state
- Secure (authenticated) channel between ECC and TLCC to query metadata;
- ECC must perform queries on a snapshot during a single transaction invocation
- TLCC must provide identity validation (orderer, msp, etc...)
- TLCC must provide channel metadata (chaincode definition)
- Maintaining FPC chaincode state only is sufficient
- Validate "normal" Enclave Registry (ERCC) transactions; however, ERCC comes with "hard-coded" endorsement policy



## Approach

Restrict Fabric functionality

- Reduces validation logic to a minimum
- Minimizes catch-up play with Fabric validation changes
- Relax the notion of endorsement policy; 
- Simplify chaincode versioning 
- No support for custom endorsing/validation plugins for non-FPC chaincodes
- Support for default MSP only
- Any transactions/blocks with unsupported features are ignored/aborted

Defined process to:
- keep redundant validation code (in stock peer and TLCC) in sync
- keep code relation traceable and changes trackable
- consistent naming and code structure as much as possible
- bonus: automated code-changes notification; notifies whenever Fabric code changes and FPC must be updates.

### Non-features


## Design

Access control
- check org is "writer"

MSP:
- Implement X509-based MSP

Endorsing:
- Phase1: Designated enclave only
- Phase2: Any enclave that runs a particular FPC chaincode

Versioning:
- single autonomously monotonously increasing version number

Validation:
- Introduce FPC transaction type (similar to introduce FPC namespace) and create
a dedicated FPC tx processor; (removes the need of custom validation plugins and interference with existing Fabric validation logic; and also gives more freedom to FPC validation logic as no it not longer bound to the structure and format of endorsement transaction).
- Support for (subset of) endorsing policies
    - Lifecycle (a majority of Orgs) (See [Link](https://hyperledger-fabric.readthedocs.io/en/release-2.2/chaincode_lifecycle.html#install-and-define-a-chaincode))
    - ERCC (same as lifecycle)
    - FPC Chaincode (see above)
- Parse chaincode definitions
- Config blocks are validated by the ordering service ([see code](https://github.com/hyperledger/fabric/blob/f27912f2f419c3b35d2c1df120f19585815eceb0/orderer/common/msgprocessor/standardchannel.go#L131))
- Validate Signatures:
    - block signatures
    - FPC endorsement signatures
    - ERCC endorsement signatures

## Development plan

### short term
See approach above
### mid/long term
Re-use Fabric code components inside trusted ledger enclave. This requires further development on go-support for SGX. Although some PoC based on graphene for go are already available but seems not be stable yet.



### QA process

## Implementation

- nanopb to parse fabric protos (alternatively we could try to use real proto for c)
- data state : levldb
    - pros:
        + snapshots;
        + batchWrites
        + c++ implementation open source
    - cons:
        + persistence needs syscalls
