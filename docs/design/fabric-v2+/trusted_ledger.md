# Trusted Ledger (TLCC)

The purpose of TLCC is to establish a trusted view of the ledger (inside an enclave)
and enable ECC to verify ledger state receiveed from the (untrusted) peer.


## Requirements

- ECC must verify all untrusted ledger state inputs from the peer with the help of TLCC
    - Secure (authenticated) channel between ECC and TLCC to query metadata;
    - TLCC must provide integrity metadata for the ledger state
    - TLCC must provide channel metadata (chaincode definition)

- TLCC must maintain a "trusted" view of the ledger
    - TLCC must be syncronized with peer's ledger
    - TLCC must detect transaction read/write conflicts
    - TLCC must be able to perform identity validation (orderer, msp, etc...)
    - TLCC must support concurrent execution of transaction processing and ledger updates
        (ECC tx must perform read/write on a stable view of the ledger state)
    

## Fabric high-level validation 

TODO pseudo code here

Config validation: [source](https://github.com/hyperledger/fabric/blob/f27912f2f419c3b35d2c1df120f19585815eceb0/common/configtx/validator.go#L163)

- check config sequence number increased by 1
- check is authorized Update [source](https://github.com/hyperledger/fabric/blob/f27912f2f419c3b35d2c1df120f19585815eceb0/common/configtx/update.go#L115)
    - verify ReadSet [source](https://github.com/hyperledger/fabric/blob/f27912f2f419c3b35d2c1df120f19585815eceb0/common/configtx/update.go#L18)
    - verify DeltaSet [source](https://github.com/hyperledger/fabric/blob/f27912f2f419c3b35d2c1df120f19585815eceb0/common/configtx/update.go#L68)
        - for each item validate policy [source](https://github.com/hyperledger/fabric/blob/f27912f2f419c3b35d2c1df120f19585815eceb0/common/policies/policy.go#L133)

Lifecycle validation:


Default validation:

- Validating identities that signed the transaction
    - read/write check?
- Verifying the signatures of the endorsers on the transaction
    - can endorse?
- Ensuring the transaction satisfies the endorsement policies of the namespaces of the corresponding chaincodes.


## Approach

Restrict Fabric functionality:

- keep Fabric validation changes in sync with TLCC
- Reduces validation logic to a minimum
    - minimizes catch-up with Fabric updates
- Restrict the notion of endorsement policy; 
- Simplify chaincode versioning (TODO)
    - No chaincode updates for now
- Support for default MSP only
- Maintaining FPC chaincode state only is sufficient
- Validate "normal" Enclave Registry (ERCC) transactions; however, ERCC comes with "hard-coded" endorsement policy


Any transactions/blocks with unsupported features are ignored/aborted

Defined process to develop TLCC:
- keep redundant validation code (in stock peer and TLCC) in sync
- keep code relation traceable and changes trackable
- consistent naming and code structure as much as possible
- bonus: 
    - automated code-changes notification; notifies whenever relevant Fabric code changes and might FPC must be updates.

### Non-features

- No custom MSP (no idemix)
- Authentication filters and decorators
- Multiple Implementations for a Single Chaincode
- Arbitrary endorsement policies
- State-based endorsement
- Custom endorsement / validation plugins for FPC
- Chaincode-to-chaincode invocations (cc2cc)
- Private Collections
- Complex (CouchDB) queries like range queries and the like

## Design

### Chaincode execution support

- secure channel
- validate proposal
    - check org is "write"
    - replay protection
    - consistent view / snapshot during tx processing on integrity metadata / ledger state

### Ledger maintainance

Channel configuration:
- Parse channel definition
    - parse consensus definition
- parse channel reader/writers
- validate block signatures (Note that this check implies correctness as onfig blocks are validated by the ordering service ([see code](https://github.com/hyperledger/fabric/blob/f27912f2f419c3b35d2c1df120f19585815eceb0/orderer/common/msgprocessor/standardchannel.go#L131)))
- Maintain msp metadata for signature validation
- parse Lifecycle policy 
    - validate Lifecycle policy
    We only support: a majority of Org.member (See [Link](https://hyperledger-fabric.readthedocs.io/en/release-2.2/chaincode_lifecycle.html#install-and-define-a-chaincode))

Access control:
- the "reader/writer/admin" access check is performed through endorsement, that is, if a valid endorsement for a transaction exists, this implies that the access check was successfully performed by the endorsers; the consortium (e.g., their admins) must ensure that the endorsement policy for a chaincode does not ruling out the authorization policy.
- TLCC only verifies endorsement policies and thereby implicitly the "writers" check is performed by the endorsers

MSP:
- One Org - One MSP mapping - One root cert
- Implement X509-based MSP
    - Restrict:
        - no intermediate
        - no CRLs
        - only support member role
        - any certificate will match to role member

Endorsing:
- Phase1: Designated enclave only
- Phase2: Any enclave that runs a particular FPC chaincode

Versioning:
- single autonomously monotonously increasing version number (TBD)

Non-FPC Validation:
- Transaction submitter identity validation
    - submitter satisfies channel's writes policy
- ERCC endorsement signatures verification
- ERCC endorsement policy validation
    - we only support (explicityly) Majority endorsement policy
    - Restrict to ERCC namespace only

FPC Validation:

- Support for (subset of) endorsing policies
    - FPC Chaincode (see above)
- Parse chaincode definitions
- Transaction submitter identity validation
    - submitter satisfies channel's writes policy
- FPC endorsement policy validation
    - Phase 1: Support only OR with a single ORG.member
    - Phase 2: Support only OR with multi ORG.member
- FPC endorsement signatures


## Development plan

### short term
See approach above

We restrict the supported endorsement policies for lifecycle, ERCC, and FPC chaincodes.

### mid/long term
- Introduce FPC transaction type (similar to introduce FPC namespace) and create
a dedicated FPC tx processor; (removes the need of custom validation plugins and interference with existing Fabric validation logic; and also gives more freedom to FPC validation logic as no it not longer bound to the structure and format of endorsement transaction).

Re-use Fabric code components inside trusted ledger enclave. This requires further development on go-support for SGX. Although some PoC based on graphene for go are already available but seems not be stable yet.

We may extend support for more enhanced endorsement policies in the future.

### QA process

## Implementation

- nanopb to parse fabric protos (alternatively we could try to use real proto for c)
- data state : leveldb
    - pros:
        + snapshots;
        + batchWrites
        + c++ implementation open source
    - cons:
        + persistence needs syscalls
