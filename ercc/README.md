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

The enclave registry is implemented as a normal chaincode and comes with a
custom validation plugin. Additionally, we use a chaincode decorator to supply
peer specific attestation metadata to the ercc chaincode when performing
attestation. Before you can install and use ercc at a peer, you have to build the
custom-vscc and decorator plugin. See more details in [/plugins/](../plugins/). 

As Fabric creates a docker image for every installed chaincode, it sometimes
could be useful to delete the ercc docker image as follows.  In particular,
in Fabric, the peer implements a lazy-build strategy to reduce unnecessary work.
That is, when you perform `peer install chaincode` for a chaincode that already
exists (in form of the docker image), the peer does not re-create the docker image.
There are two ways to update a chaincode (i.e., `ercc`). The first is to specify a
new version number whenever the chaincode is installed and use it for subsequent 
invocations. The second approach is to just delete the chaincode docker image and
then re-install it. You can use the following command. 


    $ make docker-clean
