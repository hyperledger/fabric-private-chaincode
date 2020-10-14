<!---
Licensed under Creative Commons Attribution 4.0 International License
https://creativecommons.org/licenses/by/4.0/
--->
# Enclave Registry (ercc)

The enclave registry is a chaincode that runs outside SGX and maintains a list
of all existing chaincode enclaves in the network. It performs attestation
with the chaincode enclave and stores the attestation result on the
blockchain. The attestation demonstrates that a specific chaincode executes
in an actual enclave. This enables the peers and the clients to inspect the
attestation of a chaincode enclave before invoking chaincode operations or
committing state changes.

The enclave registry is implemented as a normal go chaincode. However, since
we are using our c/c++ based attestation API to verification, we are using
the external builder functionality of Fabric.

In order to install the enclave registry chaincode at a peer, make sure that
fpc externalBuilder is available. 

```
    externalBuilders:
        # FPC Addition 0: external builder for fpc-c chaincode
        - path: $FPC_PATH/fabric/externalBuilder
          name: fpc-c
          propagateEnvironment:
              - FPC_HOSTING_MODE
              - FABRIC_LOGGING_SPEC
              - ftp_proxy
              - http_proxy
              - https_proxy
              - no_proxy
```

To build and package the enclave registry you can run the following:
```
    make package
```

This command will build and create the `ercc.tar.gz` package that can be installed on the peers.