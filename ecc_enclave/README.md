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

We provide an example chaincode that implements a simple auction.

## Build

    $ mkdir build
    $ cd build
    $ cmake ../.
    $ make

## Deploy and packaging

After successfully building the chaincode enclave you need to copy the build
ouput to the ecc project by just calling the following. Deploy copys
``enclave.signed.so``, ``libsgxcc.so``, and ``mrenclave`` to
``ecc/enclave/lib`` and the the header folder to `enc/enclave/include`.

    $ make deploy
