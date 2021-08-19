<!---
Licensed under Creative Commons Attribution 4.0 International License
https://creativecommons.org/licenses/by/4.0/
--->
# Enclave Registry (ercc)

The enclave registry is a chaincode that runs outside SGX and
maintains a list of registered FPC chaincodes and related artifacts,
e.g., associated validated enclaves and public keys. It performs
verification of the chaincode enclave's attestation which demonstrates
that a specific chaincode executes in an actual enclave. This enables
the peers and the clients to inspect the attestation of a chaincode
enclave before invoking chaincode operations or committing state changes.

The enclave registry is implemented as a normal go chaincode. However, since
we are using our c/c++ based attestation API to verification, we are using
the external builder functionality of Fabric.

The enclave registry can be run in two modes, as a normal chaincode
(where the lifecycle of the chaincode is controlled by the peer) and
as chaincode-as-a-service.
See more details below.

## Normal mode

The enclave registry will start in that mode if _neither_ of the environment
variables `CHAINCODE_PKG_ID` and `CHAINCODE_SERVER_ADDRESS` are
defined. 

It also requires that the peer's `core.yaml` defines the FPC CaaS
external builder scripts as follows:
```yaml
...
externalBuilders:
    - path: ${FPC_PATH}/fabric/externalBuilder/chaincode
      name: fpc-c
      propagateEnvironment:
          - FPC_HOSTING_MODE
          - FABRIC_LOGGING_SPEC
          - ftp_proxy
          - http_proxy
          - https_proxy
          - no_proxy
...
```

## Chaincode-as-a-Service mode

The enclave registry will start in CaaS mode if the environment variables
`CHAINCODE_PKG_ID` and `CHAINCODE_SERVER_ADDRESS` are defined.

It also requires that the peer's `core.yaml` defines the FPC CaaS
external builder scripts as follows:
```yaml
...
externalBuilders:
- path: ${FPC_PATH}/fabric/externalBuilder/chaincode_server
  name: fpc-external-launcher
  propagateEnvironment:
    - CORE_PEER_ID
    - FABRIC_LOGGING_SPEC
...
```

