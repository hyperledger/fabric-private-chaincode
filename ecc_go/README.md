<!---
Licensed under Creative Commons Attribution 4.0 International License
https://creativecommons.org/licenses/by/4.0/
--->
# Chaincode wrapper (ecc)

Before your continue here make sure you have build ``ecc_enclave`` before.
We refer to [ecc_enclave/README.md](../ecc_enclave). Otherwise, the build
might fail with the message `ecc_enclave build does not exist!`.

This is a go chaincode that is used to invoke the enclave (and hence
FPC chaincode). The chaincode logic is implemented in C++ as enclave
code and is loaded by by the go chaincode as C library
(``ecc/enclave/lib/enclave.signed.so``).  For more details on the 
chaincode implementation see [ecc_enclave](../ecc_enclave).

The FPC chaincode can be run in two modes, as a normal chaincode
(where the lifecycle of the chaincode is controlled by the peer) and
as chaincode-as-a-service.
See more details below.

## Normal mode

The chaincode will start in that mode if _neither_ of the environment
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

The chaincode will start in CaaS mode if the environment variables
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

