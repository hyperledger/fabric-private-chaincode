# Running Procedure

## Prerequisites

- FPC is properly set up and built
- `main.sh` script is placed in the chaincode directory
- `.env` file is created with environment variables

## Setup Files

**1. Set FPC_PATH:**

```bash
export FPC_PATH=/project/src/github.com/hyperledger/fabric-private-chaincode
```

**2. Create .env file:**

```bash
touch .env
# Add all environment variables as provided in the .env artifact
cp .env.example .env
```

## Running Procedure

### 1. In 1st terminal window - Setup and Deploy

```bash
# Get inside dev env
make -C $FPC_PATH/utils/docker run-dev
cd samples/chaincode/confidential-escrow

## For first time setup (includes ERCC build/Fabric network)
./main.sh full

## For subsequent runs (skip ERCC build)
./main.sh quick

## For code changes only
./main.sh chaincode
```

### 2. In 2nd terminal window - Docker Environment

```bash
# Enter docker container
docker exec -it fpc-development-main /bin/bash
cd samples/chaincode/confidential-escrow

# Setup client environment and initialize enclave
./main.sh docker
```

### 3. Run Transactions

```bash
# Run all basic tests
./main.sh test-all

# Run specific test sets
./main.sh test-basic    # Schema, debug, create assets
./main.sh test-query    # Query operations

# Interactive menu for individual tests
./main.sh
```

## Available Commands

| Command                | Description                         |
| ---------------------- | ----------------------------------- |
| `./main.sh full`       | Complete setup including ERCC build |
| `./main.sh quick`      | Quick setup (skip ERCC build)       |
| `./main.sh chaincode`  | Build chaincode only                |
| `./main.sh docker`     | Setup docker environment            |
| `./main.sh test-basic` | Run basic creation tests            |
| `./main.sh test-query` | Run query tests                     |
| `./main.sh test-all`   | Run all tests                       |
| `./main.sh clean`      | Clean and stop network              |
| `./main.sh`            | Interactive menu                    |

## Typical Workflow

1. **First time:** `./main.sh full`
2. **Enter docker:** `docker exec -it fpc-development-main /bin/bash`
3. **Setup client:** `./main.sh docker`
4. **Run tests:** `./main.sh test-all`
5. **Code changes:** Exit docker → `./main.sh chaincode` → Re-enter docker → Test again
