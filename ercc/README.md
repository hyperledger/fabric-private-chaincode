# Enclave Registry (ercc)

The enclave registry is a chaincode that runs outside SGX and maintains a list
of all existing chaincode enclaves in the network. It performs attestation
with the chaincode enclave and stores the attestation result on the
blockchain. The attestation demonstrates that a specific chaincode executes
in an actual enclave. This enables the peers and the clients to inspect the
attestation of a chaincode enclave before invoking chaincode operations or
committing state changes.

The enclave registry is implemented as a normal chaincode and comes with a
custom validation plugin. Before you can install ercc at a peer you have to
build the vscc plugin by running the following.

    $ make

