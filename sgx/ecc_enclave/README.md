<!---
Licensed under Creative Commons Attribution 4.0 International License
https://creativecommons.org/licenses/by/4.0/
--->
# Chaincode Enclave (ecc_enclave)

The chaincode enclave executes one particular chaincode, and thereby isolates
it from the peer and from other chaincodes. ECC acts as intermediary between
the chaincode in the enclave and the peer. The chaincode enclave exposes the
Fabric chaincode interface and extends it with additional support for state
encryption, attestation, and secure blockchain state access.
