# FPC with Fablo - Troubleshooting Guide

This guide helps you diagnose and resolve common issues when using FPC with Fablo.

## Table of Contents

- [General Troubleshooting Steps](#general-troubleshooting-steps)
- [Setup Issues](#setup-issues)
- [Network Issues](#network-issues)
- [Chaincode Issues](#chaincode-issues)
- [Enclave Issues](#enclave-issues)
- [Client Issues](#client-issues)
- [Performance Issues](#performance-issues)
- [SGX-Specific Issues](#sgx-specific-issues)

## General Troubleshooting Steps

### 1. Check Container Status

```bash
docker ps -a
```

Look for:
- Containers in "Exited" state
- Containers constantly restarting
- Missing containers

### 2. Check Logs

```bash
# View all logs
docker-compose -f target/fabric-docker/docker-compose.yaml logs

# View specific container logs
docker logs <container-name>

# Follow logs in real-time
docker logs -f <container-name>
```

### 3. Verify Environment Variables

```bash
echo $FPC_PATH
echo $SGX_MODE
echo $CHANNEL_NAME
echo $CC_ID
```

### 4. Clean and Restart

```bash
./scripts/teardown.sh --prune
./scripts/setup.sh
```

## Setup Issues

### Issue: "FPC_PATH is not set"

**Symptom**: Setup script fails with FPC_PATH error

**Solution**:
```bash
export FPC_PATH=/path/to/fabric-private-chaincode
echo "export FPC_PATH=$FPC_PATH" >> ~/.bashrc
source ~/.bashrc
```

### Issue: "Fablo command not found"

**Symptom**: Setup script cannot find Fablo

**Solution**:
```bash
# Download Fablo script (official method)
curl -Lf https://github.com/hyperledger-labs/fablo/releases/download/1.2.0/fablo.sh -o ./fablo
chmod +x ./fablo

# Verify installation
./fablo --version

# Or let the setup script download it automatically
./scripts/setup.sh
```

### Issue: "Docker permission denied"

**Symptom**: Cannot connect to Docker daemon

**Solution**:
```bash
# Add user to docker group
sudo usermod -aG docker $USER

# Log out and log back in, or run:
newgrp docker

# Verify
docker ps
```

### Issue: "Port already in use"

**Symptom**: Setup fails with "address already in use"

**Solution**:
```bash
# Check which process is using the port
sudo lsof -i :7041  # Replace with the conflicting port

# Stop conflicting services or change ports in fablo-config.yaml
```

### Issue: "ERCC not built"

**Symptom**: Setup fails because ERCC binary is missing

**Solution**:
```bash
cd $FPC_PATH/ercc
make clean
make all docker
```

### Issue: "Echo-go chaincode not built"

**Symptom**: Setup fails because echo-go is not built

**Solution**:
```bash
cd $FPC_PATH/samples/chaincode/echo-go
make clean
make build docker
```

## Network Issues

### Issue: "Cannot connect to peer"

**Symptom**: Client or scripts cannot reach peer

**Solution**:
```bash
# Check if peer is running
docker ps | grep peer

# Check peer logs
docker logs peer0.org1.example.com

# Verify network connectivity
docker network inspect fablo_default

# Restart peer
docker restart peer0.org1.example.com
```

### Issue: "TLS handshake failed"

**Symptom**: TLS certificate errors

**Solution**:
```bash
# Regenerate network
./scripts/teardown.sh
fablo generate
fablo up

# Verify certificate paths
ls -la target/fabric-config/crypto-config/
```

### Issue: "Channel not found"

**Symptom**: Operations fail with "channel does not exist"

**Solution**:
```bash
# Check if channel exists
peer channel list

# Recreate channel
./scripts/teardown.sh
./scripts/setup.sh
```

## Chaincode Issues

### Issue: "Chaincode not installed"

**Symptom**: Chaincode operations fail

**Solution**:
```bash
# Check installed chaincodes
peer lifecycle chaincode queryinstalled

# Reinstall if needed
cd $FPC_PATH/samples/deployment/fablo
./scripts/deploy-chaincode.sh
```

### Issue: "Chaincode container not starting"

**Symptom**: Chaincode container exits immediately

**Solution**:
```bash
# Check chaincode logs
docker logs echo-go.peer0.org1.example.com

# Common causes:
# 1. Missing package ID environment variable
source config/chaincode-env.sh

# 2. Image not built
cd $FPC_PATH/samples/chaincode/echo-go
make docker

# 3. Network connectivity
docker network inspect fablo_default
```

### Issue: "Chaincode endorsement failed"

**Symptom**: Invoke operations fail with endorsement error

**Solution**:
```bash
# Check endorsement policy
peer lifecycle chaincode querycommitted -C mychannel -n echo-go

# Verify all required peers are running
docker ps | grep peer

# Check peer logs for errors
docker logs peer0.org1.example.com
docker logs peer0.org2.example.com
```

### Issue: "Chaincode definition not agreed"

**Symptom**: Cannot commit chaincode

**Solution**:
```bash
# Check approval status
peer lifecycle chaincode checkcommitreadiness \
  --channelID mychannel \
  --name echo-go \
  --version 1.0 \
  --sequence 1

# Reapprove for all organizations
./scripts/deploy-chaincode.sh
```

## Enclave Issues

### Issue: "Enclave initialization failed"

**Symptom**: init-enclave.sh fails

**Solution**:
```bash
# Check ERCC is running
docker ps | grep ercc
docker logs ercc.peer0.org1.example.com

# Verify ERCC is committed
peer lifecycle chaincode querycommitted -C mychannel -n ercc

# Retry initialization
./scripts/init-enclave.sh
```

### Issue: "Enclave not registered"

**Symptom**: Chaincode works but enclave not in registry

**Solution**:
```bash
# Check enclave registration
peer chaincode query -C mychannel -n ercc -c '{"Args":["queryListEnclaves"]}'

# Re-register
./scripts/init-enclave.sh

# Verify
peer chaincode query -C mychannel -n ercc -c '{"Args":["queryListEnclaves"]}'
```

### Issue: "Cannot retrieve enclave credentials"

**Symptom**: Client cannot get encryption keys

**Solution**:
```bash
# Check if enclave is initialized
peer chaincode query -C mychannel -n ercc \
  -c '{"Args":["queryChaincodeEncryptionKey","echo-go"]}'

# If empty, reinitialize
./scripts/init-enclave.sh

# For simulation mode, this is expected behavior
echo $SGX_MODE  # Should be SIM or HW
```

## Client Issues

### Issue: "fpcclient: command not found"

**Symptom**: Cannot run fpcclient

**Solution**:
```bash
# Build the client
cd $FPC_PATH/samples/application/simple-cli-go
make

# Verify binary exists
ls -la fpcclient

# Make executable (if needed)
chmod +x fpcclient
```

### Issue: "Client connection timeout"

**Symptom**: Client cannot connect to network

**Solution**:
```bash
# Source environment variables
source $FPC_PATH/samples/deployment/fablo/config/fpcclient-env.sh

# Verify peer is reachable
peer channel list

# Check connection profile
cat $GATEWAY_CONFIG
```

### Issue: "Client initialization fails"

**Symptom**: fpcclient init command fails

**Solution**:
```bash
# Ensure enclave is initialized first
cd $FPC_PATH/samples/deployment/fablo
./scripts/init-enclave.sh

# Verify ERCC is accessible
peer chaincode query -C mychannel -n ercc -c '{"Args":["queryChaincodes"]}'

# Retry client init
cd $FPC_PATH/samples/application/simple-cli-go
./fpcclient init peer0.org1.example.com
```

## Performance Issues

### Issue: "Slow chaincode invocations"

**Symptom**: Operations take too long

**Solution**:
```bash
# Check system resources
docker stats

# Increase Docker resources (Docker Desktop)
# Settings -> Resources -> Increase CPU/Memory

# Check for network latency
ping localhost

# Review peer logs for bottlenecks
docker logs peer0.org1.example.com | grep -i "slow\|timeout"
```

### Issue: "High memory usage"

**Symptom**: System running out of memory

**Solution**:
```bash
# Check memory usage
free -h
docker stats

# Reduce number of containers
# Edit fablo-config.yaml to use fewer peers

# Clean up unused Docker resources
docker system prune -a
```

## SGX-Specific Issues

### Issue: "SGX device not found"

**Symptom**: /dev/sgx_enclave or /dev/sgx_provision missing

**Solution**:
```bash
# Check if SGX is enabled in BIOS
# Reboot and enable Intel SGX in BIOS settings

# Install SGX drivers
# Follow: https://github.com/intel/linux-sgx-driver

# Verify devices
ls -la /dev/sgx*

# Use simulation mode if no SGX hardware
export SGX_MODE=SIM
```

### Issue: "AESM service not running"

**Symptom**: Enclave operations fail in HW mode

**Solution**:
```bash
# Check AESM service status
sudo systemctl status aesmd

# Start AESM service
sudo systemctl start aesmd

# Enable on boot
sudo systemctl enable aesmd

# Verify
sudo systemctl status aesmd
```

### Issue: "SGX attestation failed"

**Symptom**: Enclave attestation errors in HW mode

**Solution**:
```bash
# For development, use simulation mode
export SGX_MODE=SIM

# For production, configure IAS credentials
export SGX_CREDENTIALS_PATH=$FPC_PATH/config/ias
# Add your IAS API key and SPID to the credentials path

# Verify attestation configuration
cat $SGX_CREDENTIALS_PATH/api_key.txt
```

### Issue: "Enclave out of memory"

**Symptom**: Enclave operations fail with OOM

**Solution**:
```bash
# Increase enclave heap size
# Edit chaincode Makefile and increase ENCLAVE_HEAP_SIZE

# Rebuild chaincode
cd $FPC_PATH/samples/chaincode/echo-go
make clean
make build docker

# Redeploy
cd $FPC_PATH/samples/deployment/fablo
./scripts/deploy-chaincode.sh
```

## Diagnostic Commands

### Network Health Check

```bash
# Run validation script
./scripts/validate.sh

# Check all containers
docker ps -a

# Check networks
docker network ls

# Check volumes
docker volume ls
```

### Detailed Logging

```bash
# Enable debug logging
export FABRIC_LOGGING_SPEC=DEBUG

# View detailed peer logs
docker logs peer0.org1.example.com 2>&1 | grep -i error

# View chaincode logs with timestamps
docker logs --timestamps echo-go.peer0.org1.example.com
```

### Resource Monitoring

```bash
# Monitor in real-time
docker stats

# Check disk usage
df -h
docker system df

# Check network connections
netstat -tulpn | grep -E "7041|7061|7030"
```

## Getting Help

If you cannot resolve your issue:

1. **Check logs** thoroughly for error messages
2. **Run validation** script: `./scripts/validate.sh`
3. **Search existing issues**:
   - [FPC Issues](https://github.com/hyperledger/fabric-private-chaincode/issues)
   - [Fablo Issues](https://github.com/hyperledger-labs/fablo/issues)
4. **Ask for help**:
   - [Hyperledger Discord](https://discord.gg/hyperledger)
   - [FPC Mailing List](https://lists.hyperledger.org/g/fabric-private-chaincode)
5. **Create an issue** with:
   - Error messages
   - Logs
   - Steps to reproduce
   - Environment details (OS, Docker version, SGX mode)

## Additional Resources

- [FPC Documentation](https://github.com/hyperledger/fabric-private-chaincode)
- [Fablo Documentation](https://github.com/hyperledger-labs/fablo/wiki)
- [Hyperledger Fabric Troubleshooting](https://hyperledger-fabric.readthedocs.io/en/latest/troubleshooting.html)
- [Intel SGX Documentation](https://www.intel.com/content/www/us/en/developer/tools/software-guard-extensions/overview.html)