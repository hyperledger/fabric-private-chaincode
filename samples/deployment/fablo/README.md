# FPC with Fablo - Deployment Sample

This sample demonstrates how to deploy and run Fabric Private Chaincode (FPC) using [Fablo](https://github.com/hyperledger-labs/fablo) for network management. It uses the `echo-go` chaincode and the `simple-cli-go` client application.

## Overview

[Fablo](https://github.com/hyperledger-labs/fablo) is a tool that simplifies the process of setting up Hyperledger Fabric networks. This sample shows how to:

1. Configure a Fablo network with FPC-specific requirements (SGX support)
2. Deploy the FPC Enclave Registry Chaincode (ERCC)
3. Deploy the echo-go FPC chaincode
4. Interact with the chaincode using the simple-cli-go client

## Prerequisites

### Required Software

- **Docker** (v20.x or later)
- **Docker Compose** (v1.29 or later)
- **Fablo** (v1.2.0 or later) - Will be downloaded by setup script
- **FPC** - This repository must be built (see main [README](../../../README.md))
- **Intel SGX** drivers and PSW (for hardware mode)
  - For simulation mode, SGX hardware is not required
- **curl** - For downloading Fablo

### Environment Setup

Make sure you have built FPC and set the `FPC_PATH` environment variable:

```bash
export FPC_PATH=/path/to/fabric-private-chaincode
```

For SGX hardware mode, also set:

```bash
export SGX_MODE=HW
export SGX_CREDENTIALS_PATH=$FPC_PATH/config/ias
```

For simulation mode (no SGX hardware required):

```bash
export SGX_MODE=SIM
```

## Quick Start

### 1. Build FPC Components

First, build the echo-go chaincode and ERCC:

```bash
# Build echo-go chaincode
make -C $FPC_PATH/samples/chaincode/echo-go build docker

# Build ERCC
make -C $FPC_PATH/ercc all docker

# Build FPC chaincode environment
make -C $FPC_PATH/utils/docker pull
# or build from scratch:
# make -C $FPC_PATH/utils/docker build
```

### 2. Setup and Start Network

Run the setup script to install Fablo (if needed) and start the network:

```bash
cd $FPC_PATH/samples/deployment/fablo
./scripts/setup.sh
```

This script will:
- Install Fablo if not already installed
- Generate the Fabric network using Fablo
- Start the network
- Deploy ERCC
- Deploy the echo-go chaincode
- Initialize the FPC enclave

### 3. Interact with the Chaincode

Use the simple-cli-go client to interact with the chaincode:

```bash
# Source environment variables
source ./config/fpcclient-env.sh

# Initialize the enclave (first time only)
cd $FPC_PATH/samples/application/simple-cli-go
./fpcclient init peer0.org1.example.com

# Invoke the chaincode
./fpcclient invoke "Hello FPC with Fablo!"

# Query the chaincode
./fpcclient query
```

### 4. Shutdown

To stop and clean up the network:

```bash
cd $FPC_PATH/samples/deployment/fablo
./scripts/teardown.sh
```

## Directory Structure

```
fablo/
├── README.md                          # This file
├── fablo-config.json                  # Fablo network configuration
├── fablo-config.yaml                  # Alternative YAML format
├── docker-compose-fpc-override.yaml   # FPC-specific Docker overrides
├── scripts/
│   ├── setup.sh                       # Main setup script
│   ├── deploy-ercc.sh                 # Deploy ERCC
│   ├── deploy-chaincode.sh            # Deploy echo-go chaincode
│   ├── init-enclave.sh                # Initialize FPC enclave
│   ├── teardown.sh                    # Cleanup script
│   └── validate.sh                    # Validation checks
├── config/
│   ├── connection-org1.yaml           # Connection profile for Org1
│   ├── connection-org2.yaml           # Connection profile for Org2
│   └── fpcclient-env.sh               # Environment variables for client
├── compose/
│   ├── ercc-compose.yaml              # ERCC container configuration
│   └── ecc-compose.yaml               # Echo chaincode container configuration
└── docs/
    ├── SETUP.md                       # Detailed setup instructions
    ├── USAGE.md                       # Usage guide
    └── TROUBLESHOOTING.md             # Common issues and solutions
```

## Network Topology

The Fablo configuration creates a network with:

- **2 Organizations**: Org1 and Org2
- **2 Peers**: One peer per organization (with SGX support)
- **1 Orderer**: Solo orderer for simplicity
- **1 Channel**: `mychannel`
- **2 Chaincodes**:
  - `ercc` - FPC Enclave Registry Chaincode
  - `echo-go` - FPC Echo chaincode

## Key Features

### FPC-Specific Configurations

This sample includes several FPC-specific configurations:

1. **SGX Device Mounting**: Peers have access to `/dev/sgx` device
2. **AESM Service**: Volume mount for SGX AESM service
3. **FPC Environment Variables**: Proper configuration for FPC runtime
4. **Chaincode as a Service**: FPC chaincodes run as external services
5. **ERCC First**: ERCC is deployed before application chaincodes

### Differences from Test-Network Sample

Compared to the `test-network` sample, this Fablo-based sample:

- Uses Fablo for network generation (no fabric-samples dependency)
- Provides declarative YAML configuration
- Generates its own crypto materials
- Simplifies network management with Fablo commands
- Demonstrates integration with a different network setup tool

## Documentation

For more detailed information, see:

- [SETUP.md](docs/SETUP.md) - Detailed setup instructions
- [USAGE.md](docs/USAGE.md) - Usage examples and workflows
- [TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md) - Common issues and solutions

## References

- [FPC Documentation](https://github.com/hyperledger/fabric-private-chaincode)
- [Fablo Documentation](https://github.com/hyperledger-labs/fablo/wiki)
- [Hyperledger Fabric Documentation](https://hyperledger-fabric.readthedocs.io/)

## License

This sample is part of the Fabric Private Chaincode project and follows the same license (Apache-2.0).