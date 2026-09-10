# FPC with Fablo - Usage Guide

This guide demonstrates how to use the FPC with Fablo deployment sample, including chaincode interactions and common operations.

## Table of Contents

- [Quick Start](#quick-start)
- [Using the FPC Client](#using-the-fpc-client)
- [Chaincode Operations](#chaincode-operations)
- [Network Management](#network-management)
- [Advanced Usage](#advanced-usage)
- [Examples](#examples)

## Quick Start

### 1. Start the Network

```bash
cd $FPC_PATH/samples/deployment/fablo
./scripts/setup.sh
```

### 2. Configure Client Environment

```bash
source config/fpcclient-env.sh
```

### 3. Use the Client

```bash
cd $FPC_PATH/samples/application/simple-cli-go

# Initialize (first time only)
./fpcclient init peer0.org1.example.com

# Invoke chaincode
./fpcclient invoke "Hello FPC with Fablo!"

# Query chaincode
./fpcclient query
```

## Using the FPC Client

### Client Overview

The `simple-cli-go` client (`fpcclient`) provides a command-line interface for interacting with FPC chaincodes.

### Available Commands

```bash
./fpcclient --help
```

Commands:
- `init` - Initialize the enclave
- `invoke` - Invoke chaincode (write operation)
- `query` - Query chaincode (read operation)

### Initialize Enclave

The `init` command must be run once before using the chaincode:

```bash
./fpcclient init <peer-id>
```

Example:
```bash
./fpcclient init peer0.org1.example.com
```

This command:
1. Connects to the specified peer
2. Retrieves enclave credentials
3. Establishes a secure channel with the enclave

### Invoke Chaincode

The `invoke` command performs write operations:

```bash
./fpcclient invoke <message>
```

Example:
```bash
./fpcclient invoke "This is a private message"
```

The message is:
1. Encrypted by the client
2. Sent to the enclave
3. Processed inside the secure enclave
4. Stored on the ledger (encrypted)

### Query Chaincode

The `query` command performs read operations:

```bash
./fpcclient query
```

This retrieves the last stored message from the enclave.

## Chaincode Operations

### Echo-Go Chaincode Functions

The echo-go chaincode provides the following functions:

1. **storeAsset** - Store a message
   ```bash
   ./fpcclient invoke "Your message here"
   ```

2. **retrieveAsset** - Retrieve the last message
   ```bash
   ./fpcclient query
   ```

### Direct Peer Commands

You can also interact with the chaincode using peer commands:

#### Setup Environment

```bash
source config/fpcclient-env.sh

# Set peer environment for Org1
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_ADDRESS="localhost:7041"
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_TLS_ROOTCERT_FILE="$FPC_PATH/samples/deployment/fablo/target/fabric-config/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"
export CORE_PEER_MSPCONFIGPATH="$FPC_PATH/samples/deployment/fablo/target/fabric-config/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp"
```

#### Invoke Chaincode

```bash
peer chaincode invoke \
  -o localhost:7030 \
  --ordererTLSHostnameOverride orderer0.group1.orderer.example.com \
  --tls \
  --cafile "$FPC_PATH/samples/deployment/fablo/target/fabric-config/crypto-config/ordererOrganizations/orderer.example.com/orderers/orderer0.group1.orderer.example.com/tls/ca.crt" \
  -C mychannel \
  -n echo-go \
  -c '{"Args":["storeAsset","Test message"]}' \
  --waitForEvent \
  --peerAddresses localhost:7041 \
  --tlsRootCertFiles "$FPC_PATH/samples/deployment/fablo/target/fabric-config/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"
```

#### Query Chaincode

```bash
peer chaincode query \
  -C mychannel \
  -n echo-go \
  -c '{"Args":["retrieveAsset"]}'
```

### Query ERCC

Check enclave registration:

```bash
peer chaincode query \
  -C mychannel \
  -n ercc \
  -c '{"Args":["queryListEnclaves"]}'
```

Query chaincode encryption key:

```bash
peer chaincode query \
  -C mychannel \
  -n ercc \
  -c '{"Args":["queryChaincodeEncryptionKey","echo-go"]}'
```

## Network Management

### Start Network

```bash
cd $FPC_PATH/samples/deployment/fablo
./scripts/setup.sh
```

### Stop Network

```bash
./scripts/teardown.sh
```

### Restart Network

```bash
./scripts/teardown.sh
./scripts/setup.sh
```

### Validate Deployment

```bash
./scripts/validate.sh
```

### View Logs

```bash
# All containers
docker-compose -f target/fabric-docker/docker-compose.yaml logs -f

# Specific container
docker logs -f peer0.org1.example.com
docker logs -f echo-go.peer0.org1.example.com
docker logs -f ercc.peer0.org1.example.com
```

### Check Container Status

```bash
docker ps
```

### Access Container Shell

```bash
# Peer container
docker exec -it peer0.org1.example.com bash

# Chaincode container
docker exec -it echo-go.peer0.org1.example.com bash
```

## Advanced Usage

### Using Different Organizations

Switch to Org2:

```bash
export CORE_PEER_LOCALMSPID="Org2MSP"
export CORE_PEER_ADDRESS="localhost:7061"
export CORE_PEER_TLS_ROOTCERT_FILE="$FPC_PATH/samples/deployment/fablo/target/fabric-config/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt"
export CORE_PEER_MSPCONFIGPATH="$FPC_PATH/samples/deployment/fablo/target/fabric-config/crypto-config/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp"
```

### Upgrading Chaincode

1. Build new version:
   ```bash
   cd $FPC_PATH/samples/chaincode/echo-go
   make build docker
   ```

2. Update version in scripts:
   ```bash
   export CC_VERSION="2.0"
   export CC_SEQUENCE="2"
   ```

3. Redeploy:
   ```bash
   cd $FPC_PATH/samples/deployment/fablo
   ./scripts/deploy-chaincode.sh
   ./scripts/init-enclave.sh
   ```

### Monitoring Performance

Use Docker stats:
```bash
docker stats
```

Monitor specific containers:
```bash
docker stats peer0.org1.example.com echo-go.peer0.org1.example.com
```

## Examples

### Example 1: Basic Echo Operation

```bash
# Start network
cd $FPC_PATH/samples/deployment/fablo
./scripts/setup.sh

# Configure environment
source config/fpcclient-env.sh

# Use client
cd $FPC_PATH/samples/application/simple-cli-go
./fpcclient init peer0.org1.example.com
./fpcclient invoke "Hello from FPC!"
./fpcclient query
```

Expected output:
```
Hello from FPC!
```

### Example 2: Multiple Invocations

```bash
./fpcclient invoke "Message 1"
./fpcclient invoke "Message 2"
./fpcclient invoke "Message 3"
./fpcclient query  # Returns "Message 3"
```

### Example 3: Verify Encryption

The data on the ledger is encrypted. To verify:

```bash
# Query the ledger directly (shows encrypted data)
peer chaincode query -C mychannel -n echo-go -c '{"Args":["retrieveAsset"]}'

# Query through FPC client (shows decrypted data)
./fpcclient query
```

### Example 4: Multi-Org Endorsement

Invoke with endorsements from both organizations:

```bash
peer chaincode invoke \
  -o localhost:7030 \
  --ordererTLSHostnameOverride orderer0.group1.orderer.example.com \
  --tls \
  --cafile "$FPC_PATH/samples/deployment/fablo/target/fabric-config/crypto-config/ordererOrganizations/orderer.example.com/orderers/orderer0.group1.orderer.example.com/tls/ca.crt" \
  -C mychannel \
  -n echo-go \
  -c '{"Args":["storeAsset","Multi-org message"]}' \
  --waitForEvent \
  --peerAddresses localhost:7041 \
  --tlsRootCertFiles "$FPC_PATH/samples/deployment/fablo/target/fabric-config/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt" \
  --peerAddresses localhost:7061 \
  --tlsRootCertFiles "$FPC_PATH/samples/deployment/fablo/target/fabric-config/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt"
```

## Best Practices

1. **Always initialize the enclave** before first use
2. **Source environment variables** before running commands
3. **Check logs** if operations fail
4. **Run validation** after setup to ensure everything works
5. **Clean up** with teardown script when done
6. **Use simulation mode** for development and testing
7. **Monitor resources** to ensure sufficient capacity

## Troubleshooting

For common issues and solutions, see [TROUBLESHOOTING.md](TROUBLESHOOTING.md).

## Additional Resources

- [FPC Documentation](https://github.com/hyperledger/fabric-private-chaincode)
- [Simple CLI Go Client](../../../application/simple-cli-go/README.md)
- [Echo-Go Chaincode](../../../chaincode/echo-go/README.md)
- [Fablo Documentation](https://github.com/hyperledger-labs/fablo/wiki)