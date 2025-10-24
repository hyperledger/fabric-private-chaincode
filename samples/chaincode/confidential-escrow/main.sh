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
    local env_file="${1:-}"

    # Get the directory where the script is located
    SCRIPT_DIR="$FPC_PATH/samples/chaincode/confidential-escrow/"

    # If no argument provided, use default .env
    if [ -z "$env_file" ]; then
        env_file="$SCRIPT_DIR/.env"
    fi

    if [ -f "$env_file" ]; then
        source "$env_file"
        log_info "Environment variables loaded from $(basename $env_file)"
    else
        log_error "Environment file not found: $env_file"
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
    local env_file="${1:-}"

    log_info "=== SETTING UP DOCKER ENVIRONMENT ==="

    # Source the appropriate environment file
    if [ -n "$env_file" ]; then
        source_env "$env_file"
    else
        source_env # Use default .env
    fi

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

WALLET_UUID=""
WALLET2_UUID=""
ESCROW_UUID=""
DIGITAL_ASSET_UUID=""
DIGITAL_ASSET_JSON=""

store_asset_data() {
    local output="$1"
    DIGITAL_ASSET_JSON=$(echo "$output" | grep '^>' | sed 's/^> //')
    DIGITAL_ASSET_UUID=$(echo "$DIGITAL_ASSET_JSON" | grep -o '"@key":"digitalAsset:[^"]*"' | cut -d':' -f3 | tr -d '"')
}

store_wallet_data() {
    local output="$1"
    WALLET_UUID=$(echo "$output" | grep '^>' | sed 's/^> //' | grep -o '"@key":"wallet:[^"]*"' | cut -d':' -f3 | tr -d '"')
}

store_escrow_data() {
    local output="$1"
    ESCROW_UUID=$(echo "$output" | grep '^>' | sed 's/^> //' | grep -o '"@key":"escrow:[^"]*"' | cut -d':' -f3 | tr -d '"')
}

test_create_asset() {
    log_info "Creating digital asset..."
    cd $FPC_PATH/samples/application/simple-cli-go
    local output=$(./fpcclient invoke createDigitalAsset '{
        "name": "CBDCC",
        "symbol": "CBDCC", 
        "decimals": 2,
        "totalSupply": 1000000,
        "owner": "central_bank",
        "issuerHash": "sha256:abc123"
      }' 2>&1)
    echo "$output"
    store_asset_data "$output"
}

test_create_wallet() {
    log_info "Creating wallet..."
    cd $FPC_PATH/samples/application/simple-cli-go
    local output=$(./fpcclient invoke createWallet "{
    \"walletId\": \"wallet-111\",
    \"ownerId\": \"Abhinav\",
    \"ownerCertHash\": \"sha256:def456\", 
    \"balances\": [0],
    \"digitalAssetTypes\": [$DIGITAL_ASSET_JSON]
  }" 2>&1)
    echo "$output"
    echo "DIGITIAL ASSET JSON:> $DIGITAL_ASSET_JSON"
    store_wallet_data "$output"
}

test_create_wallet2() {
    log_info "Creating second wallet..."
    cd $FPC_PATH/samples/application/simple-cli-go
    local output=$(./fpcclient invoke createWallet "{
        \"walletId\": \"wallet-222\",
        \"ownerId\": \"Bob\",
        \"ownerCertHash\": \"sha256:ghi789\", 
        \"balances\": [0],
        \"digitalAssetTypes\": [$DIGITAL_ASSET_JSON]
    }" 2>&1)
    echo "$output"
    echo "DIGITIAL ASSET JSON:> $DIGITAL_ASSET_JSON"
    WALLET2_UUID=$(echo "$output" | grep '^>' | sed 's/^> //' | grep -o '"@key":"wallet:[^"]*"' | cut -d':' -f3 | tr -d '"')
}

test_create_user_dir() {
    log_info "Creating user directory..."
    cd $FPC_PATH/samples/application/simple-cli-go
    ./fpcclient invoke createUserDir "{
        \"publicKeyHash\": \"pubkey-hash-123\",
        \"walletId\": \"$WALLET_UUID\",
        \"certHash\": \"sha256:def456\"
    }"
}

# test_create_escrow() {
#     log_info "Creating escrow..."
#     cd $FPC_PATH/samples/application/simple-cli-go
#     local output=$(./fpcclient invoke createEscrow "{
#     \"escrowId\": \"escrow-111\",
#     \"buyerPubKey\": \"buyer_pub\",
#     \"sellerPubKey\": \"seller_pub\",
#     \"amount\": 20,
#     \"assetType\": $DIGITAL_ASSET_JSON,
#     \"conditionValue\": \"$(echo -n 'secret123' | sha256sum | cut -d' ' -f1)\",
#     \"status\": \"Active\",
#     \"buyerCertHash\": \"sha256:buyer_cert\"
#   }" 2>&1)
#     echo "$output"
#     store_escrow_data "$output"
# }

test_create_and_lock_escrow() {
    log_info "Creating escrow and locking funds..."
    cd $FPC_PATH/samples/application/simple-cli-go
    local output=$(./fpcclient invoke createAndLockEscrow "{
        \"escrowId\": \"escrow-111\",
        \"buyerPubKey\": \"buyer_pub\",
        \"sellerPubKey\": \"seller_pub\",
        \"sellerWalletId\": \"$WALLET2_UUID\",
        \"amount\": 20,
        \"assetType\": $DIGITAL_ASSET_JSON,
        \"conditionValue\": \"$(echo -n 'secret123' | sha256sum | cut -d' ' -f1)\",
        \"payerWalletId\": \"$WALLET_UUID\",
        \"payerCertHash\": \"sha256:def456\"
    }" 2>&1)
    echo "$output"
    store_escrow_data "$output"
}

test_query_asset() {
    log_info "Querying digital asset..."
    cd $FPC_PATH/samples/application/simple-cli-go
    log_info $DIGITAL_ASSET_UUID
    log_info $DIGITAL_ASSET_JSON
    ./fpcclient query readDigitalAsset "{\"uuid\": \"$DIGITAL_ASSET_UUID\"}"
}

test_query_wallet() {
    log_info "Querying wallet..."
    cd $FPC_PATH/samples/application/simple-cli-go
    ./fpcclient query readWallet "{\"uuid\": \"$WALLET_UUID\"}"
}

test_get_balance() {
    log_info "Testing getBalance transaction"

    # Test getting balance for CBDCC in wallet1
    cd $FPC_PATH/samples/application/simple-cli-go
    ./fpcclient query getBalance "{
        \"walletId\": \"$WALLET_UUID\",
        \"assetSymbol\": \"CBDCC\", 
        \"ownerCertHash\": \"sha256:def456\"
    }"
}

test_get_wallet_by_owner() {
    log_info "Testing getWalletByOwner transaction"
    cd $FPC_PATH/samples/application/simple-cli-go
    ./fpcclient query getWalletByOwner "{
        \"walletUuid\": \"$WALLET_UUID\",
        \"ownerCertHash\": \"sha256:def456\"
    }"
}

test_mint_tokens() {
    log_info "Testing mintTokens transaction"
    cd $FPC_PATH/samples/application/simple-cli-go
    ./fpcclient invoke mintTokens "{
        \"assetId\": \"$DIGITAL_ASSET_UUID\",
        \"walletId\": \"$WALLET_UUID\",
        \"amount\": 100,
        \"issuerCertHash\": \"sha256:abc123\"
    }"
}

test_transfer_tokens() {
    log_info "Testing transferTokens transaction"
    cd $FPC_PATH/samples/application/simple-cli-go
    ./fpcclient invoke transferTokens "{
        \"fromWalletId\": \"$WALLET_UUID\",
        \"toWalletId\": \"$WALLET2_UUID\", 
        \"assetId\": \"$DIGITAL_ASSET_UUID\",
        \"amount\": 50,
        \"senderCertHash\": \"sha256:def456\"
    }"
}

test_burn_tokens() {
    log_info "Testing burnTokens transaction"
    cd $FPC_PATH/samples/application/simple-cli-go
    ./fpcclient invoke burnTokens "{
        \"assetId\": \"$DIGITAL_ASSET_UUID\",
        \"walletId\": \"$WALLET_UUID\",
        \"amount\": 25,
        \"issuerCertHash\": \"sha256:abc123\"
    }"
}

# test_lock_funds_in_escrow() {
#     log_info "Testing lockFundsInEscrow transaction"
#     cd $FPC_PATH/samples/application/simple-cli-go
#     ./fpcclient invoke lockFundsInEscrow "{
#         \"escrowId\": \"$ESCROW_UUID\",
#         \"payerWalletId\": \"$WALLET_UUID\",
#         \"amount\": 20,
#         \"assetId\": \"$DIGITAL_ASSET_UUID\",
#         \"payerCertHash\": \"sha256:def456\"
#     }"
# }

test_get_escrow_balance() {
    log_info "Testing getEscrowBalance transaction"
    cd $FPC_PATH/samples/application/simple-cli-go
    ./fpcclient query getEscrowBalance "{
        \"walletId\": \"$WALLET_UUID\",
        \"assetSymbol\": \"CBDCC\",
        \"ownerCertHash\": \"sha256:def456\"
    }"
}

test_verify_escrow_condition() {
    log_info "Testing verifyEscrowCondition transaction"
    cd $FPC_PATH/samples/application/simple-cli-go
    ./fpcclient invoke verifyEscrowCondition "{
        \"escrowId\": \"$ESCROW_UUID\",
        \"secret\": \"secret\",
        \"parcelId\": \"123\"
    }"
}

test_release_escrow() {
    log_info "Testing releaseEscrow transaction"
    cd $FPC_PATH/samples/application/simple-cli-go
    ./fpcclient invoke releaseEscrow "{
        \"escrowId\": \"$ESCROW_UUID\",
        \"payerWalletId\": \"$WALLET_UUID\"
    }"
}

test_refund_escrow() {
    log_info "Testing refundEscrow transaction (on new escrow)"
    cd $FPC_PATH/samples/application/simple-cli-go
    # Create another escrow for refund test
    local output=$(./fpcclient invoke createEscrow "{
        \"escrowId\": \"escrow-refund\",
        \"buyerPubKey\": \"buyer_pub2\",
        \"sellerPubKey\": \"seller_pub2\",
        \"amount\": 5,
        \"assetType\": $DIGITAL_ASSET_JSON,
        \"conditionValue\": \"different_condition\",
        \"status\": \"Created\",
        \"buyerCertHash\": \"sha256:buyer_cert2\"
    }" 2>&1)

    local REFUND_ESCROW_UUID=$(echo "$output" | grep '^>' | sed 's/^> //' | grep -o '"@key":"escrow:[^"]*"' | cut -d':' -f3 | tr -d '"')

    # Lock funds in refund escrow
    ./fpcclient invoke lockFundsInEscrow "{
        \"escrowId\": \"$REFUND_ESCROW_UUID\",
        \"payerWalletId\": \"$WALLET_UUID\",
        \"amount\": 5,
        \"assetId\": \"$DIGITAL_ASSET_UUID\",
        \"payerCertHash\": \"sha256:def456\"
    }"

    # Now refund
    ./fpcclient invoke refundEscrow "{
        \"escrowId\": \"$REFUND_ESCROW_UUID\",
        \"payerWalletId\": \"$WALLET_UUID\"
    }"
}

# test_read_user_dir() {
#     log_info "Testing readUserDir transaction with authentication"
#     cd $FPC_PATH/samples/application/simple-cli-go
#     # This should work - correct cert hash
#     ./fpcclient query readUserDir "{
#         \"uuid\": \"pubkey-hash-123\",
#         \"certHash\": \"sha256:def456\"
#     }"
#
#     log_info "Testing unauthorized access (should fail)"
#     # This should fail - wrong cert hash
#     ./fpcclient query readUserDir "{
#         \"uuid\": \"pubkey-hash-123\",
#         \"certHash\": \"wrong_hash\"
#     }" || log_info "Expected failure - unauthorized access blocked"
# }

test_query_escrow() {
    log_info "Querying escrow..."
    cd $FPC_PATH/samples/application/simple-cli-go
    ./fpcclient query readEscrow "{\"uuid\": \"$ESCROW_UUID\"}"
}

# Batch operations
run_tests() {
    log_info "=== RUNNING TESTS ==="
    source_env
    test_schema
    test_debug
    test_create_asset
    test_create_wallet
    test_create_wallet2
    test_create_user_dir
    test_create_escrow
    test_query_asset
    test_query_wallet
    test_query_escrow
    test_get_wallet_by_owner
    test_mint_tokens
    test_get_balance
    test_transfer_tokens
    test_get_balance
    test_burn_tokens
    test_get_balance
    test_get_wallet_by_owner

    # New escrow tests
    log_info "=== ESCROW SYSTEM TESTS ==="
    test_lock_funds_in_escrow
    test_get_balance
    test_get_escrow_balance
    test_verify_escrow_condition
    test_release_escrow
    test_get_balance
    test_get_escrow_balance
    test_refund_escrow
    test_get_balance
    test_get_escrow_balance
    # test_read_user_dir

    log_success "=== TESTS COMPLETED ==="
}

# Main menu
show_menu() {
    echo
    echo "=== FPC CONTROL SCRIPT ==="
    echo "SETUP OPTIONS:"
    echo "1. Full Setup (ERCC + Network + Install)"
    echo "2. Quick Setup (Skip ERCC build)"
    echo "3. Setup Docker Environment"
    echo
    echo "TEST OPTIONS:"
    echo "4. Run All Tests"
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
    "docker")
        setup_docker
        ;;
    "test-all")
        run_tests
        ;;
    "menu")
        while true; do
            show_menu
            read -p "Choose an option (0-8): " choice
            case $choice in
            1) main "full" ;;
            2) main "quick" ;;
            3) main "docker" ;;
            4) main "test-all" ;;
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

# Only run main if script is executed directly, not sourced
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    main "$@"
fi
