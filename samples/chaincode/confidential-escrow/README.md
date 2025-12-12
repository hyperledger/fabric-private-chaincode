# Confidential Escrow Chaincode

A privacy-preserving escrow system built on Hyperledger Fabric Private Chaincode (FPC) that enables secure digital asset management with programmable conditional payments.

## Overview

This chaincode implements a confidential escrow mechanism for digital assets, combining:

- **Privacy-Preserving Transactions**: All transaction data is encrypted within Intel SGX enclaves
- **Programmable Escrow Contracts**: Automated conditional fund releases based on cryptographic verification
- **Multi-Asset Support**: Manage multiple token types within individual wallets
- **Certificate-Based Authorization**: Fine-grained access control using X.509 certificate hashes

## Architecture

### Core Components

**Assets**

- `DigitalAsset`: Fungible tokens with controlled supply (CBDC, stablecoins, etc.)
- `Wallet`: User accounts supporting multiple asset types with separate available and escrowed balances
- `Escrow`: Smart contracts holding funds pending condition fulfillment
- `UserDirectory`: Privacy-preserving public key to wallet UUID mapping

**Transaction Operations**

- Asset lifecycle: Create, mint, transfer, burn
- Wallet management: Create wallets, query balances
- Escrow workflow: Lock funds, verify conditions, release or refund

## Project Structure

```
confidential-escrow/
├── chaincode/
│   ├── assets/           # Asset type definitions
│   ├── transactions/     # Transaction handlers
│   ├── header/           # Chaincode metadata
│   ├── escrow.go         # Main chaincode implementation
│   ├── server.go         # CCaaS server setup
│   └── setup.go          # Component registration
├── main.go               # Entry point
├── main.sh               # Deployment and test automation
└── README.md             # This file
```

### Security Model

1. **Access Control**: All operations require valid certificate hash verification
2. **Atomic Escrow**: Funds move from available to escrowed balance during lock, preventing double-spending
3. **Condition Verification**: SHA-256 hash of `(secret + parcelId)` ensures only authorized parties can release funds
4. **Confidential Execution**: FPC ensures transaction details remain private within SGX enclaves

## Running Procedure

### Prerequisites

- FPC is properly set up and built
- `multi_user_dashboard.sh ` script is placed in the chaincode directory
- `.env.alice` and `.env.bob` file is present

### Setup Files

**1. Set FPC_PATH:**

```bash
export FPC_PATH=/project/src/github.com/hyperledger/fabric-private-chaincode
```

### Running Procedure

#### 1. In 1st terminal window - Setup and Deploy

```bash
# Get inside dev env
make -C $FPC_PATH/utils/docker run-dev
cd samples/chaincode/confidential-escrow

# Interactive menu
./multi_user_dashboard.sh

# Choose Option 1. or 2. as per your setup condn
```

#### 2. In 2nd terminal window - Docker Environment (`Alice`)

```bash
# Enter docker container
docker exec -it fpc-development-main /bin/bash
cd samples/chaincode/confidential-escrow

# Interactive menu
./multi_user_dashboard.sh

# Setup Alice using Option 3.
```

#### 3. In 3rd terminal window - Docker Environment (`Bob`)

```bash
# Enter docker container
docker exec -it fpc-development-main /bin/bash
cd samples/chaincode/confidential-escrow

# Interactive menu
./multi_user_dashboard.sh

# Setup Bob using Option 4.
```

#### 4. In 3rd terminal window - Docker Environment (`Monitor`)

```bash
# Enter docker container
docker exec -it fpc-development-main /bin/bash
cd samples/chaincode/confidential-escrow

# Interactive menu
./multi_user_dashboard.sh

# Setup Bob using Option 5.
```

#### 5. Run Tests

```bash
# Run all basic tests
./multi_user_dashboard.sh

# Chosing Option 7.
```

## Contributing

When adding new features:

1. Define asset types in `chaincode/assets/`
2. Implement transaction logic in `chaincode/transactions/`
3. Register new components in `chaincode/setup.go`
4. Add test cases to `main.sh`
5. Update this README with usage examples
