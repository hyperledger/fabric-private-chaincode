#!/bin/bash

# FPC Test Commands Script
# Description: Easy testing script for FPC chaincode functions

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[TEST]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Export environment variables
export_env_vars() {
    export CC_ID=confidential-escrow
    export CHANNEL_NAME=mychannel
    export CORE_PEER_ADDRESS=localhost:7051
    export CORE_PEER_ID=peer0.org1.example.com
    export CORE_PEER_ORG_NAME=org1
    export CORE_PEER_LOCALMSPID=Org1MSP
    export CORE_PEER_MSPCONFIGPATH=$FPC_PATH/samples/deployment/test-network/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
    export CORE_PEER_TLS_CERT_FILE=$FPC_PATH/samples/deployment/test-network/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/server.crt
    export CORE_PEER_TLS_ENABLED="true"
    export CORE_PEER_TLS_KEY_FILE=$FPC_PATH/samples/deployment/test-network/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/server.key
    export CORE_PEER_TLS_ROOTCERT_FILE=$FPC_PATH/samples/deployment/test-network/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
    export ORDERER_CA=$FPC_PATH/samples/deployment/test-network/fabric-samples/test-network/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
    export GATEWAY_CONFIG=$FPC_PATH/samples/deployment/test-network/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/connection-org1.yaml
    export FPC_ENABLED=true
    export RUN_CCAAS=true
}

# Auto-detect and navigate to the right directory
setup_paths() {
    # Check if FPC_PATH is set
    if [ -z "$FPC_PATH" ]; then
        log_error "FPC_PATH is not set. Please export FPC_PATH=/path/to/your/fpc"
        exit 1
    fi

    # Define the fpcclient directory
    FPCCLIENT_DIR="$FPC_PATH/samples/application/simple-cli-go"

    # Check if fpcclient exists in the expected location
    if [ ! -f "$FPCCLIENT_DIR/fpcclient" ]; then
        log_error "fpcclient not found at $FPCCLIENT_DIR/fpcclient"
        log_error "Make sure FPC is properly built and fpcclient exists"
        exit 1
    fi

    # Store current directory to return later if needed
    ORIGINAL_DIR=$(pwd)

    log_info "Using fpcclient at: $FPCCLIENT_DIR"

    # Change to fpcclient directory
    cd "$FPCCLIENT_DIR"
}

# Test functions
test_get_schema() {
    log_info "Getting schema..."
    ./fpcclient invoke getSchema
    log_success "Schema retrieved"
}

test_debug() {
    log_info "Running debug test..."
    ./fpcclient invoke debugTest '{}'
    log_success "Debug test completed"
}

test_create_digital_asset() {
    log_info "Creating digital asset (CBDC)..."
    ./fpcclient invoke createDigitalAsset '{
      "name": "CBDC",
      "symbol": "CBDC", 
      "decimals": 2,
      "totalSupply": 1000000,
      "owner": "central_bank",
      "issuerHash": "sha256:abc123"
    }'
    log_success "Digital asset created"
}

test_create_wallet() {
    log_info "Creating wallet for Abhinav..."
    ./fpcclient invoke createWallet '{
      "walletId": "wallet-123",
      "ownerId": "Abhinav",
      "ownerCertHash": "sha256:def456", 
      "balance": 0,
      "digitalAssetType": "CBDC"
    }'
    log_success "Wallet created"
}

test_create_escrow() {
    log_info "Creating escrow..."
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
    log_success "Escrow created"
}

# Query functions
query_wallet() {
    local wallet_id=${1:-"wallet-123"}
    log_info "Querying wallet: $wallet_id"
    cd "$FPCCLIENT_DIR"
    ./fpcclient invoke readWallet '{"uuid": "f7e3afb2-c6b9-5ae4-80de-102d57e86333"}'
}

query_escrow() {
    local escrow_id=${1:-"escrow-456"}
    log_info "Querying escrow: $escrow_id"
    cd "$FPCCLIENT_DIR"
    ./fpcclient invoke readEscrow '{"uuid": "403496d6-f2d6-5adc-ab55-a32d87764261"}'
}

query_digital_asset() {
    local asset_name=${1:-"CBDC"}
    log_info "Querying digital asset: $asset_name"
    cd "$FPCCLIENT_DIR"
    ./fpcclient invoke readDigitalAsset '{"uuid": "59a4e99c-c705-513e-8db2-f04fa81bceb8"}'
}

# Batch testing functions
run_basic_tests() {
    log_info "=== RUNNING BASIC TESTS ==="
    test_get_schema
    test_debug
    test_create_digital_asset
    test_create_wallet
    test_create_escrow
    log_success "=== BASIC TESTS COMPLETED ==="
}

run_query_tests() {
    log_info "=== RUNNING QUERY TESTS ==="
    query_digital_asset
    query_wallet
    query_escrow
    log_success "=== QUERY TESTS COMPLETED ==="
}

# Interactive testing menu
show_test_menu() {
    echo
    echo "=== FPC TEST MENU ==="
    echo "1. Run Basic Tests (schema, debug, create assets)"
    echo "2. Run Query Tests"
    echo "3. Run All Tests"
    echo "4. Individual Tests Submenu"
    echo "5. Custom Query"
    echo "0. Exit"
    echo
}

show_individual_menu() {
    echo
    echo "=== INDIVIDUAL TESTS ==="
    echo "1. Get Schema"
    echo "2. Debug Test"
    echo "3. Create Digital Asset"
    echo "4. Create Wallet"
    echo "5. Create Escrow"
    echo "6. Query Wallet"
    echo "7. Query Escrow"
    echo "8. Query Digital Asset"
    echo "0. Back to Main Menu"
    echo
}

run_custom_query() {
    echo "Available functions: getWallet, getEscrow, getDigitalAsset, etc."
    read -p "Enter function name: " func_name
    read -p "Enter JSON parameters (or {} for empty): " params

    log_info "Running custom query: $func_name with params: $params"
    cd "$FPCCLIENT_DIR"
    ./fpcclient query $func_name "$params"
}

# Main execution
main() {
    setup_paths
    export_env_vars

    case "${1:-menu}" in
    "basic")
        run_basic_tests
        ;;
    "query")
        run_query_tests
        ;;
    "all")
        run_basic_tests
        echo
        run_query_tests
        ;;
    "menu")
        while true; do
            show_test_menu
            read -p "Choose an option (1-7): " choice
            case $choice in
            1) run_basic_tests ;;
            2) run_query_tests ;;
            3)
                run_basic_tests
                echo
                run_query_tests
                echo
                run_advanced_tests
                ;;
            4)
                while true; do
                    show_individual_menu
                    read -p "Choose an option (1-12): " ind_choice
                    case $ind_choice in
                    1) test_get_schema ;;
                    2) test_debug ;;
                    3) test_create_digital_asset ;;
                    4) test_create_wallet ;;
                    5) test_create_escrow ;;
                    6)
                        read -p "Enter wallet ID (default: wallet-123): " wid
                        query_wallet "${wid:-wallet-123}"
                        ;;
                    7)
                        read -p "Enter escrow ID (default: escrow-456): " eid
                        query_escrow "${eid:-escrow-456}"
                        ;;
                    8)
                        read -p "Enter asset name (default: CBDC): " aname
                        query_digital_asset "${aname:-CBDC}"
                        ;;
                    9)
                        read -p "Enter wallet ID (default: wallet-123): " wid
                        read -p "Enter amount (default: 5000): " amt
                        test_fund_wallet "${wid:-wallet-123}" "${amt:-5000}"
                        ;;
                    10)
                        read -p "Enter from wallet ID (default: wallet-123): " from_w
                        read -p "Enter to wallet ID (default: wallet-456): " to_w
                        read -p "Enter amount (default: 1000): " amt
                        test_transfer_funds "${from_w:-wallet-123}" "${to_w:-wallet-456}" "${amt:-1000}"
                        ;;
                    11)
                        read -p "Enter escrow ID (default: escrow-456): " eid
                        read -p "Enter reveal value (default: secret123): " rval
                        test_complete_escrow "${eid:-escrow-456}" "${rval:-secret123}"
                        ;;
                    12) break ;;
                    *) log_error "Invalid option. Please choose 1-12." ;;
                    esac
                    echo
                    read -p "Press Enter to continue..."
                done
                ;;
            5) run_custom_query ;;
            0) exit 0 ;;
            *) log_error "Invalid option. Please choose 1-7." ;;
            esac
            echo
            read -p "Press Enter to continue..."
        done
        ;;
    *)
        echo "Usage: $0 [basic|query|advanced|all|menu]"
        echo "  basic    - Run basic creation tests"
        echo "  query    - Run query tests"
        echo "  advanced - Run advanced scenario tests"
        echo "  all      - Run all tests"
        echo "  menu     - Interactive menu (default)"
        exit 1
        ;;
    esac
}

# Run main function
main "$@"
