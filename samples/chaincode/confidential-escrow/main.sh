#!/bin/bash

# Combined FPC Setup and Test Script
# Description: One script to rule them all

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if FPC_PATH is set
check_fpc_path() {
    if [ -z "$FPC_PATH" ]; then
        log_error "FPC_PATH is not set. Please export FPC_PATH=/path/to/your/fpc"
        exit 1
    fi
    log_info "Using FPC_PATH: $FPC_PATH"
}

# Source environment variables
source_env() {
    # Get the directory where the script is located
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

    if [ -f "$SCRIPT_DIR/.env" ]; then
        source "$SCRIPT_DIR/.env"
        log_info "Environment variables loaded from .env"
    else
        log_error ".env file not found. Make sure it exists in script directory: $SCRIPT_DIR"
        exit 1
    fi
}

# Function to run commands with error handling
run_cmd() {
    local cmd="$1"
    local desc="$2"

    log_info "$desc"
    echo "Running: $cmd"

    if eval "$cmd"; then
        log_success "$desc - COMPLETED"
    else
        log_error "$desc - FAILED"
        exit 1
    fi
}

# Build ERCC (one time requirement)
build_ercc() {
    log_info "=== BUILDING ERCC ==="
    run_cmd "GOOS=linux make -C $FPC_PATH/ercc build docker" "Building ERCC"
}

# Build chaincode
build_chaincode() {
    log_info "=== BUILDING CHAINCODE ==="
    run_cmd "GOOS=linux make -C $FPC_PATH/samples/chaincode/confidential-escrow with_go docker" "Building chaincode"
}

# One-time setup (run only once)
initial_setup() {
    log_info "=== INITIAL SETUP (One-time only) ==="
    cd $FPC_PATH/samples/deployment/test-network
    run_cmd "./setup.sh" "Running initial setup"
}

# Setup test network
setup_network() {
    log_info "=== SETTING UP NETWORK ==="
    cd $FPC_PATH/samples/deployment/test-network/fabric-samples/test-network

    run_cmd "./network.sh down" "Bringing down network"
    run_cmd "./network.sh up -ca" "Starting network"
    run_cmd "./network.sh createChannel -c mychannel" "Creating channel"
}

# Install FPC
install_fpc() {
    log_info "=== INSTALLING FPC ==="
    export CC_ID=confidential-escrow
    export CC_PATH="$FPC_PATH/samples/chaincode/confidential-escrow/"
    export CC_VER=$(cat "$FPC_PATH/samples/chaincode/confidential-escrow/mrenclave")

    cd $FPC_PATH/samples/deployment/test-network
    run_cmd "./installFPC.sh" "Installing FPC"
}

# Start ERCC-ECC
start_ercc() {
    log_info "=== STARTING ERCC-ECC ==="
    export EXTRA_COMPOSE_FILE="$FPC_PATH/samples/chaincode/confidential-escrow/confidential-escrow-compose.yaml"
    cd $FPC_PATH/samples/deployment/test-network
    run_cmd "make ercc-ecc-start" "Starting ERCC-ECC"
}

# Setup docker environment
setup_docker() {
    log_info "=== SETTING UP DOCKER ENVIRONMENT ==="
    source_env

    cd $FPC_PATH/samples/deployment/test-network
    run_cmd "./update-connection.sh" "Updating connections"
    run_cmd "./update-external-connection.sh" "Updating external connections"

    cd $FPC_PATH/samples/application/simple-cli-go
    run_cmd "./fpcclient init $CORE_PEER_ID" "Initializing enclave"
    log_success "Docker environment ready!"
}

# Test functions
test_schema() {
    log_info "Getting schema..."
    cd $FPC_PATH/samples/application/simple-cli-go
    ./fpcclient invoke getSchema
}

test_debug() {
    log_info "Running debug test..."
    cd $FPC_PATH/samples/application/simple-cli-go
    ./fpcclient invoke debugTest '{}'
}

test_create_asset() {
    log_info "Creating digital asset..."
    cd $FPC_PATH/samples/application/simple-cli-go
    ./fpcclient invoke createDigitalAsset '{
      "name": "CBDC",
      "symbol": "CBDC", 
      "decimals": 2,
      "totalSupply": 1000000,
      "owner": "central_bank",
      "issuerHash": "sha256:abc123"
    }'
}

test_create_wallet() {
    log_info "Creating wallet..."
    cd $FPC_PATH/samples/application/simple-cli-go
    ./fpcclient invoke createWallet '{
      "walletId": "wallet-123",
      "ownerId": "Abhinav",
      "ownerCertHash": "sha256:def456", 
      "balance": 0,
      "digitalAssetType": "CBDC"
    }'
}

test_create_escrow() {
    log_info "Creating escrow..."
    cd $FPC_PATH/samples/application/simple-cli-go
    ./fpcclient invoke createEscrow '{
      "escrowId": "escrow-456",
      "buyerPubKey": "buyer_pub",
      "sellerPubKey": "seller_pub",
      "amount": 1000,
      "assetType": "CBDC",
      "conditionValue": "sha256:secret123",
      "status": "Active",
      "buyerCertHash": "sha256:buyer_cert"
    }'
}

test_query_asset() {
    log_info "Querying digital asset..."
    cd $FPC_PATH/samples/application/simple-cli-go
    ./fpcclient query readDigitalAsset '{"uuid": "59a4e99c-c705-513e-8db2-f04fa81bceb8"}'
}

test_query_wallet() {
    log_info "Querying wallet..."
    cd $FPC_PATH/samples/application/simple-cli-go
    ./fpcclient query readWallet '{"uuid": "f7e3afb2-c6b9-5ae4-80de-102d57e86333"}'
}

test_query_escrow() {
    log_info "Querying escrow..."
    cd $FPC_PATH/samples/application/simple-cli-go
    ./fpcclient query readEscrow '{"uuid": "403496d6-f2d6-5adc-ab55-a32d87764261"}'
}

# Batch operations
run_basic_tests() {
    log_info "=== RUNNING BASIC TESTS ==="
    source_env
    test_schema
    test_debug
    test_create_asset
    test_create_wallet
    test_create_escrow
    log_success "=== BASIC TESTS COMPLETED ==="
}

run_query_tests() {
    log_info "=== RUNNING QUERY TESTS ==="
    source_env
    test_query_asset
    test_query_wallet
    test_query_escrow
    log_success "=== QUERY TESTS COMPLETED ==="
}

# Clean network
clean_network() {
    log_info "=== CLEANING NETWORK ==="
    cd $FPC_PATH/samples/deployment/test-network/fabric-samples/test-network
    ./network.sh down
    log_success "Network cleaned"
}

# Main menu
show_menu() {
    echo
    echo "=== FPC CONTROL SCRIPT ==="
    echo "SETUP OPTIONS:"
    echo "1. Full Setup (ERCC + Network + Install)"
    echo "2. Quick Setup (Skip ERCC build)"
    echo "3. Build Chaincode Only"
    echo "4. Setup Docker Environment"
    echo "5. Clean Network"
    echo
    echo "TEST OPTIONS:"
    echo "6. Run Basic Tests"
    echo "7. Run Query Tests"
    echo "8. Run All Tests"
    echo
    echo "0. Exit"
    echo
}

# Main execution
main() {
    check_fpc_path

    case "${1:-menu}" in
    "full")
        build_ercc
        build_chaincode
        initial_setup
        setup_network
        install_fpc
        start_ercc
        ;;
    "quick")
        build_chaincode
        setup_network
        install_fpc
        start_ercc
        ;;
    "chaincode")
        build_chaincode
        ;;
    "docker")
        setup_docker
        ;;
    "clean")
        clean_network
        ;;
    "test-basic")
        run_basic_tests
        ;;
    "test-query")
        run_query_tests
        ;;
    "test-all")
        run_basic_tests
        echo
        run_query_tests
        ;;
    "menu")
        while true; do
            show_menu
            read -p "Choose an option (0-8): " choice
            case $choice in
            1) main "full" ;;
            2) main "quick" ;;
            3) main "chaincode" ;;
            4) main "docker" ;;
            5) main "clean" ;;
            6) main "test-basic" ;;
            7) main "test-query" ;;
            8) main "test-all" ;;
            0) exit 0 ;;
            *) log_error "Invalid option" ;;
            esac
            echo
            read -p "Press Enter to continue..."
        done
        ;;
    *)
        echo "Usage: $0 [full|quick|chaincode|docker|clean|test-basic|test-query|test-all|menu]"
        exit 1
        ;;
    esac
}

main "$@"
