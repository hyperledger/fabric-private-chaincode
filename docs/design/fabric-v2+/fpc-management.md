# FPC Management

This document defines the management APIs of FPC.
Such definition builds on the design diagrams describing the enclave creation, 
the enclave registration and the chaincode key generation.

[//]: # (## Enclave and Chaincode APIs)

[//]: # (### Create Chaincode Enclave)

[//]: # (### Register Chaincode Enclave)

[//]: # (### Generate FPC Chaincode Keys)

## Admin Commands


### Create Chaincode Enclave
This command results in the creation of a new chaincode enclave,
which generates its enclave-specific cryptographic keys and produces a hardware-based attestation.

```peer lifecycle chaincode createenclave -n <chaincode id>```

A successful command returns the base64-encoded string of the enclave's Credentials (see components diagram).

```Credentials: <base64-encoded string>```

### Register Chaincode Enclave

This command registers the enclave's credentials on the Enclave Registry (chaincode).

```peer lifecycle chaincode registereenclave <base64-encoded Credentials structure>```

A successful command returns `0`.

### Generate FPC Chaincode Keys

This command makes an (already created) enclave generate chaincode-specific cryptographic keys,
and registers them on the Enclave Registry (chaincode).

```peer lifecycle chaincode generatekeys -n <chaincode id>```

### Key Distribution

Not supported in the initial version of FPC.
