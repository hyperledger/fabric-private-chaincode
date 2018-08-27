# Lab Name
Hyperledger Fabric Secure Chaincode Execution

# Short Description
This lab enables Secure Chaincode Execution using Intel SGX for Hyperledger Fabric.

The transparency and resilience gained from blockchain protocols ensure the integrity of blockchain applications and yet contradicts the goal to keep application state confidential and to maintain privacy for its users.

To remedy this problem, this project uses Trusted Execution Environments (TEEs), in particular Intel Software Guard Extensions (SGX), to protect the privacy of chaincode data and computation from potentially untrusted peers.

Intel SGX is the most prominent TEE today and available with commodity CPUs. It establishes trusted execution contexts called enclaves on a CPU, which isolate data and programs from the host operating system in hardware and ensure that outputs are correct.

This lab provides a framework to develop and execute Fabric chaincode within an enclave.  Furthermore, Fabric extensions for chaincode enclave registration and transaction verification are provided.

# Scope of Lab
This lab proposes an architecture to enable Secure Chaincode Execution using Intel SGX for Hyperledger Fabric.  We provide an initial proof-of-concept implementation of the proposed architecture. The main goal of this lab is to discuss and refine the proposed architecture involving the Hyperledger community.

# Initial Committers
- https://github.com/mbrandenburger Marcus Brandenburger (bur@zurich.ibm.com)
- https://github.com/cca88 Christian Cachin (cca@zurich.ibm.com)
- https://github.com/rrkapitz Rüdiger Kapitza (kapitza@ibr.cs.tu-bs.de)
- https://github.com/ale-linux Alessandro Sorniotti (aso@zurich.ibm.com)

# Sponsor
https://github.com/mastersingh24 Gari Singh (garis@us.ibm.com)
