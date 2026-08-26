# FPC with Fablo - Detailed Setup Guide

This guide provides detailed instructions for setting up and running Fabric Private Chaincode (FPC) with Fablo network management.

## Table of Contents

- [Prerequisites](#prerequisites)
- [System Requirements](#system-requirements)
- [Installation Steps](#installation-steps)
- [Configuration](#configuration)
- [Verification](#verification)
- [Next Steps](#next-steps)

## Prerequisites

### Required Software

1. **Docker** (v20.x or later)
   ```bash
   docker --version
   ```

2. **Docker Compose** (v1.29 or later)
   ```bash
   docker-compose --version
   ```

3. **curl** - Required for downloading Fablo
   ```bash
   curl --version
   ```

4. **Go** (v1.19 or later) - Required for building FPC components
   ```bash
   go version
   ```

5. **Git**
   ```bash
   git --version
   ```

### Intel SGX (Optional for Hardware Mode)

For running in SGX hardware mode, you need:

- Intel SGX-capable CPU
- SGX drivers installed
- SGX Platform Software (PSW) installed
- AESM service running

To check if your system supports SGX:
```bash
cpuid | grep SGX
```

For development and testing, you can use **simulation mode** which doesn't require SGX hardware.

## System Requirements

### Minimum Requirements

- **CPU**: 4 cores (8 cores recommended)
- **RAM**: 8 GB (16 GB recommended)
- **Disk**: 20 GB free space
- **OS**: Linux (Ubuntu 20.04 or later recommended)

### Supported Operating Systems

- Ubuntu 20.04 LTS or later
- Other Linux distributions with Docker support
- macOS (with Docker Desktop) - Simulation mode only
- Windows with WSL2 (with Docker Desktop) - Simulation mode only

## Installation Steps

### Step 1: Clone and Build FPC Repository

1. Clone the FPC repository:
   ```bash
   git clone https://github.com/hyperledger/fabric-private-chaincode.git
   cd fabric-private-chaincode
   ```

2. Set the FPC_PATH environment variable:
   ```bash
   export FPC_PATH=$(pwd)
   echo "export FPC_PATH=$FPC_PATH" >> ~/.bashrc
   ```

3. Build FPC components:
   ```bash
   # Build ERCC (Enclave Registry Chaincode)
   make -C $FPC_PATH/ercc all docker
   
   # Build echo-go chaincode
   make -C $FPC_PATH/samples/chaincode/echo-go build docker
   
   # Build or pull FPC chaincode environment
   make -C $FPC_PATH/utils/docker pull
   # Or build from scratch:
   # make -C $FPC_PATH/utils/docker build
   ```

### Step 2: Install Fablo

Fablo will be automatically downloaded by the setup script, but you can also install it manually:

```bash
# Download Fablo script (recommended method from official repository)
curl -Lf https://github.com/hyperledger-labs/fablo/releases/download/1.2.0/fablo.sh -o ./fablo && chmod +x ./fablo
```

Verify installation:
```bash
./fablo --version
```

For more installation options, see the [official Fablo repository](https://github.com/hyperledger-labs/fablo).

### Step 3: Configure SGX Mode

Choose your SGX mode based on your hardware:

**For Simulation Mode (No SGX hardware required):**
```bash
export SGX_MODE=SIM
echo "export SGX_MODE=SIM" >> ~/.bashrc
```

**For Hardware Mode (Requires SGX-capable CPU):**
```bash
export SGX_MODE=HW
echo "export SGX_MODE=HW" >> ~/.bashrc

# Set SGX credentials path (if using IAS attestation)
export SGX_CREDENTIALS_PATH=$FPC_PATH/config/ias
```

### Step 4: Navigate to Fablo Sample

```bash
cd $FPC_PATH/samples/deployment/fablo
```

### Step 5: Run Setup Script

The setup script will:
- Install Fablo (if not already installed)
- Generate and start the Fabric network
- Deploy ERCC
- Deploy echo-go chaincode
- Initialize the FPC enclave

```bash
./scripts/setup.sh
```

This process takes approximately 5-10 minutes depending on your system.

## Configuration

### Network Configuration

The network is configured in `fablo-config.yaml`:

- **2 Organizations**: Org1 and Org2
- **2 Peers**: One per organization
- **1 Orderer**: Solo orderer
- **1 Channel**: mychannel

To modify the network topology, edit `fablo-config.yaml` and regenerate:
```bash
fablo generate
```

### FPC-Specific Configuration

FPC-specific Docker configurations are in `docker-compose-fpc-override.yaml`:

- SGX device mounting
- AESM service volume
- FPC environment variables

### Chaincode Configuration

The echo-go chaincode is configured in `compose/ecc-compose.yaml`:

- Chaincode as a Service (CaaS) deployment
- SGX device access
- Network connectivity

## Verification

### 1. Check Docker Containers

Verify all containers are running:
```bash
docker ps
```

You should see containers for:
- peer0.org1.example.com
- peer0.org2.example.com
- orderer0.group1.orderer.example.com
- ercc.peer0.org1.example.com
- ercc.peer0.org2.example.com
- echo-go.peer0.org1.example.com
- echo-go.peer0.org2.example.com

### 2. Run Validation Script

```bash
./scripts/validate.sh
```

This script checks:
- Container status
- Network connectivity
- ERCC deployment
- Chaincode deployment
- Enclave registration
- Basic chaincode operations

### 3. Check Logs

View logs for specific containers:
```bash
# Peer logs
docker logs peer0.org1.example.com

# ERCC logs
docker logs ercc.peer0.org1.example.com

# Echo-go chaincode logs
docker logs echo-go.peer0.org1.example.com
```

## Troubleshooting

### Common Issues

1. **Port conflicts**: Ensure ports 7020-7061 and 7030 are available
2. **Docker permissions**: User must be in the docker group
3. **SGX device access**: Check `/dev/sgx_enclave` and `/dev/sgx_provision` exist (HW mode)
4. **Memory issues**: Ensure sufficient RAM is available

For detailed troubleshooting, see [TROUBLESHOOTING.md](TROUBLESHOOTING.md).

## Next Steps

After successful setup:

1. **Source environment variables**:
   ```bash
   source config/fpcclient-env.sh
   ```

2. **Use the FPC client**:
   ```bash
   cd $FPC_PATH/samples/application/simple-cli-go
   ./fpcclient init peer0.org1.example.com
   ./fpcclient invoke "Hello FPC!"
   ./fpcclient query
   ```

3. **Explore the sample**: See [USAGE.md](USAGE.md) for detailed usage examples

4. **Tear down when done**:
   ```bash
   cd $FPC_PATH/samples/deployment/fablo
   ./scripts/teardown.sh
   ```

## Additional Resources

- [FPC Documentation](https://github.com/hyperledger/fabric-private-chaincode)
- [Fablo Documentation](https://github.com/hyperledger-labs/fablo/wiki)
- [Hyperledger Fabric Documentation](https://hyperledger-fabric.readthedocs.io/)
- [Intel SGX Documentation](https://www.intel.com/content/www/us/en/developer/tools/software-guard-extensions/overview.html)

## Support

For issues and questions:
- FPC: [GitHub Issues](https://github.com/hyperledger/fabric-private-chaincode/issues)
- Fablo: [GitHub Issues](https://github.com/hyperledger-labs/fablo/issues)
- Hyperledger Fabric: [Discord](https://discord.gg/hyperledger)