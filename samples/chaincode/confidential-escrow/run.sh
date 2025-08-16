#!/bin/bash

# FPC and CC-Tools Debug Setup Script
# Description: Automated setup script for fast debugging with FPC and cc-tools

set -e # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
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

# Setup dev container
setup_dev_container() {
    log_info "=== SETTING UP DEV CONTAINER ==="
    run_cmd "make -C $FPC_PATH/utils/docker run-dev" "Getting inside Dev container"
}

# Build ERCC (one time requirement)
build_ercc() {
    log_info "=== BUILDING ERCC (One-time requirement) ==="
    run_cmd "GOOS=linux make -C $FPC_PATH/ercc build docker" "Building ERCC"
}

# Build chaincode (run every time code changes)
build_chaincode() {
    log_info "=== BUILDING CHAINCODE ==="
    run_cmd "GOOS=linux make -C $FPC_PATH/samples/chaincode/confidential-escrow with_go docker" "Building confidential-escrow chaincode"
}

# Setup test network
setup_test_network() {
    log_info "=== SETTING UP TEST NETWORK ==="

    cd $FPC_PATH/samples/deployment/test-network/fabric-samples/test-network

    run_cmd "./network.sh down" "Bringing down existing network"
    run_cmd "./network.sh up -ca" "Starting network with CA"
    run_cmd "./network.sh createChannel -c mychannel" "Creating mychannel"
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
start_ercc_ecc() {
    log_info "=== STARTING ERCC-ECC ==="

    export EXTRA_COMPOSE_FILE="$FPC_PATH/samples/chaincode/confidential-escrow/confidential-escrow-compose.yaml"
    cd $FPC_PATH/samples/deployment/test-network
    run_cmd "make ercc-ecc-start" "Starting ERCC-ECC"
}

# Setup client environment (to be run inside docker)
setup_client_env() {
    log_info "=== SETTING UP CLIENT ENVIRONMENT ==="

    cd $FPC_PATH/samples/deployment/test-network
    run_cmd "./update-connection.sh" "Preparing connections profile"
    run_cmd "./update-external-connection.sh" "Updating external connection profile"
}

# Export client settings
export_client_settings() {
    log_info "=== EXPORTING CLIENT SETTINGS ==="

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

    log_success "Client environment variables exported"
}

# Initialize enclave
init_enclave() {
    log_info "=== INITIALIZING ENCLAVE ==="
    cd $FPC_PATH/samples/application/simple-cli-go
    run_cmd "./fpcclient init $CORE_PEER_ID" "Initializing enclave"
}

# Run test commands
run_tests() {
    log_info "=== RUNNING TESTS ==="
    cd $FPC_PATH/samples/application/simple-cli-go

    log_info "Getting schema..."
    ./fpcclient invoke getSchema

    log_info "Running debug test..."
    ./fpcclient invoke debugTest '{}'

    log_info "Creating digital asset..."
    ./fpcclient invoke createDigitalAsset '{
      "name": "CBDC",
      "symbol": "CBDC", 
      "decimals": 2,
      "totalSupply": 1000000,
      "owner": "central_bank",
      "issuerHash": "sha256:abc123"
    }'

    log_info "Creating wallet..."
    ./fpcclient invoke createWallet '{
      "walletId": "wallet-123",
      "ownerId": "Abhinav",
      "ownerCertHash": "sha256:def456", 
      "balance": 0,
      "digitalAssetType": "CBDC"
    }'

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

    log_success "All tests completed successfully!"
}

# Docker helper function
run_in_docker() {
    log_info "=== RUNNING COMMANDS IN DOCKER CONTAINER ==="

    # Create a script to run inside docker
    cat >/tmp/docker_commands.sh <<'EOF'
#!/bin/bash
set -e

# Setup client environment
cd $FPC_PATH/samples/deployment/test-network
./update-connection.sh
./update-external-connection.sh

# Export client settings
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

# Initialize enclave
cd $FPC_PATH/samples/application/simple-cli-go
./fpcclient init $CORE_PEER_ID

echo "Docker setup completed. You can now run tests."
EOF

    chmod +x /tmp/docker_commands.sh

    log_info "Copying script to docker container..."
    docker cp /tmp/docker_commands.sh fpc-development-main:/tmp/

    log_info "Running setup inside docker container..."
    docker exec -it fpc-development-main /bin/bash -c "/tmp/docker_commands.sh"

    rm /tmp/docker_commands.sh
}

# Main menu function
show_menu() {
    echo
    echo "=== FPC DEBUG SETUP SCRIPT ==="
    echo "1. Full Setup (One-time setup)"
    echo "2. Quick Setup (Skip ERCC build)"
    echo "3. Build Chaincode Only"
    echo "4. Setup Docker Environment"
    echo "5. Run Tests Only"
    echo "6. Clean and Restart Network"
    echo "7. Exit"
    echo
}

# Main execution logic
main() {
    check_fpc_path

    case "${1:-menu}" in
    "full")
        log_info "Starting full setup..."
        build_ercc
        build_chaincode
        setup_test_network
        install_fpc
        start_ercc_ecc
        log_success "Full setup completed! Now run: docker exec -it fpc-development-main /bin/bash"
        log_info "Then run this script with 'docker' option inside the container"
        ;;
    "quick")
        log_info "Starting quick setup (skipping ERCC build)..."
        build_chaincode
        setup_test_network
        install_fpc
        start_ercc_ecc
        log_success "Quick setup completed!"
        ;;
    "chaincode")
        log_info "Building chaincode only..."
        build_chaincode
        log_success "Chaincode build completed!"
        ;;
    "docker")
        setup_client_env
        export_client_settings
        init_enclave
        log_success "Docker environment setup completed! You can now run tests."
        ;;
    "test")
        export_client_settings
        run_tests
        ;;
    "clean")
        log_info "Cleaning and restarting network..."
        cd $FPC_PATH/samples/deployment/test-network/fabric-samples/test-network
        ./network.sh down
        log_success "Network cleaned. Run full or quick setup to restart."
        ;;
    "menu")
        while true; do
            show_menu
            read -p "Choose an option (1-7): " choice
            case $choice in
            1) main "full" ;;
            2) main "quick" ;;
            3) main "chaincode" ;;
            4) main "docker" ;;
            5) main "test" ;;
            6) main "clean" ;;
            7) exit 0 ;;
            *) log_error "Invalid option. Please choose 1-7." ;;
            esac
            echo
            read -p "Press Enter to continue..."
        done
        ;;
    *)
        echo "Usage: $0 [full|quick|chaincode|docker|test|clean|menu]"
        echo "  full      - Complete setup including ERCC build"
        echo "  quick     - Quick setup (skip ERCC build)"
        echo "  chaincode - Build chaincode only"
        echo "  docker    - Setup docker environment (run inside container)"
        echo "  test      - Run tests only"
        echo "  clean     - Clean and stop network"
        echo "  menu      - Interactive menu (default)"
        exit 1
        ;;
    esac
}

# Run main function with all arguments
main "$@"
