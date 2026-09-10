#!/bin/bash

# Copyright IBM Corp. All Rights Reserved.
# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

# Main setup script for FPC with Fablo deployment
# This script orchestrates the complete setup process

set -euo pipefail

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SAMPLE_DIR="$(dirname "$SCRIPT_DIR")"

# Source common functions if available
if [ -f "$SCRIPT_DIR/common.sh" ]; then
    source "$SCRIPT_DIR/common.sh"
fi

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

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    log_success "Docker found: $(docker --version)"
    
    # Check Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    log_success "Docker Compose found: $(docker-compose --version)"
    
    # Check FPC_PATH
    if [ -z "${FPC_PATH:-}" ]; then
        log_error "FPC_PATH environment variable is not set."
        log_error "Please set it to your FPC repository path: export FPC_PATH=/path/to/fabric-private-chaincode"
        exit 1
    fi
    
    if [ ! -d "$FPC_PATH" ]; then
        log_error "FPC_PATH directory does not exist: $FPC_PATH"
        exit 1
    fi
    log_success "FPC_PATH is set: $FPC_PATH"
    
    # Check SGX_MODE
    if [ -z "${SGX_MODE:-}" ]; then
        log_warning "SGX_MODE not set, defaulting to SIM mode"
        export SGX_MODE=SIM
    fi
    log_info "SGX_MODE: $SGX_MODE"
    
    # Check if FPC is built
    if [ ! -f "$FPC_PATH/ercc/ercc" ]; then
        log_error "FPC ERCC not built. Please run: make -C $FPC_PATH/ercc all"
        exit 1
    fi
    log_success "FPC ERCC is built"
    
    if [ ! -f "$FPC_PATH/samples/chaincode/echo-go/_build/lib/enclave.signed.so" ]; then
        log_error "Echo-go chaincode not built. Please run: make -C $FPC_PATH/samples/chaincode/echo-go"
        exit 1
    fi
    log_success "Echo-go chaincode is built"
}

# Install Fablo if not present
install_fablo() {
    log_info "Checking Fablo installation..."
    
    # Check if fablo exists in current directory or PATH
    if [ -f "./fablo" ]; then
        log_success "Fablo found in current directory"
        return 0
    fi
    
    if command -v fablo &> /dev/null; then
        log_success "Fablo is already installed: $(fablo --version)"
        return 0
    fi
    
    log_info "Installing Fablo..."
    
    # Download Fablo script from official repository
    # Using the recommended installation method from https://github.com/hyperledger-labs/fablo
    FABLO_VERSION="${FABLO_VERSION:-1.2.0}"
    
    curl -Lf "https://github.com/hyperledger-labs/fablo/releases/download/${FABLO_VERSION}/fablo.sh" -o ./fablo
    
    if [ $? -ne 0 ]; then
        log_error "Failed to download Fablo"
        log_info "Please check your internet connection or download manually from:"
        log_info "https://github.com/hyperledger-labs/fablo/releases"
        exit 1
    fi
    
    chmod +x ./fablo
    log_success "Fablo installed successfully (version ${FABLO_VERSION})"
}

# Generate and start Fablo network
start_fablo_network() {
    log_info "Generating Fablo network configuration..."
    
    cd "$SAMPLE_DIR"
    
    # Generate network
    fablo generate
    
    log_info "Starting Fablo network..."
    fablo up
    
    # Apply FPC-specific overrides
    log_info "Applying FPC-specific Docker Compose overrides..."
    docker-compose -f "$SAMPLE_DIR/target/fabric-docker/docker-compose.yaml" \
                   -f "$SAMPLE_DIR/docker-compose-fpc-override.yaml" \
                   up -d
    
    log_success "Fablo network started successfully"
    
    # Wait for network to be ready
    log_info "Waiting for network to be ready..."
    sleep 10
}

# Deploy ERCC
deploy_ercc() {
    log_info "Deploying ERCC (Enclave Registry Chaincode)..."
    
    if [ -f "$SCRIPT_DIR/deploy-ercc.sh" ]; then
        bash "$SCRIPT_DIR/deploy-ercc.sh"
    else
        log_error "deploy-ercc.sh script not found"
        exit 1
    fi
    
    log_success "ERCC deployed successfully"
}

# Deploy echo-go chaincode
deploy_chaincode() {
    log_info "Deploying echo-go chaincode..."
    
    if [ -f "$SCRIPT_DIR/deploy-chaincode.sh" ]; then
        bash "$SCRIPT_DIR/deploy-chaincode.sh"
    else
        log_error "deploy-chaincode.sh script not found"
        exit 1
    fi
    
    log_success "Echo-go chaincode deployed successfully"
}

# Initialize FPC enclave
init_enclave() {
    log_info "Initializing FPC enclave..."
    
    if [ -f "$SCRIPT_DIR/init-enclave.sh" ]; then
        bash "$SCRIPT_DIR/init-enclave.sh"
    else
        log_error "init-enclave.sh script not found"
        exit 1
    fi
    
    log_success "FPC enclave initialized successfully"
}

# Validate deployment
validate_deployment() {
    log_info "Validating deployment..."
    
    if [ -f "$SCRIPT_DIR/validate.sh" ]; then
        bash "$SCRIPT_DIR/validate.sh"
    else
        log_warning "validate.sh script not found, skipping validation"
        return 0
    fi
    
    log_success "Deployment validated successfully"
}

# Print next steps
print_next_steps() {
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}  FPC with Fablo Setup Complete!${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo "Next steps:"
    echo ""
    echo "1. Source the environment variables:"
    echo -e "   ${BLUE}source $SAMPLE_DIR/config/fpcclient-env.sh${NC}"
    echo ""
    echo "2. Navigate to the simple-cli-go client:"
    echo -e "   ${BLUE}cd \$FPC_PATH/samples/application/simple-cli-go${NC}"
    echo ""
    echo "3. Initialize the client (first time only):"
    echo -e "   ${BLUE}./fpcclient init peer0.org1.example.com${NC}"
    echo ""
    echo "4. Invoke the chaincode:"
    echo -e "   ${BLUE}./fpcclient invoke \"Hello FPC with Fablo!\"${NC}"
    echo ""
    echo "5. Query the chaincode:"
    echo -e "   ${BLUE}./fpcclient query${NC}"
    echo ""
    echo "To tear down the network:"
    echo -e "   ${BLUE}cd $SAMPLE_DIR && ./scripts/teardown.sh${NC}"
    echo ""
}

# Main execution
main() {
    log_info "Starting FPC with Fablo setup..."
    echo ""
    
    check_prerequisites
    echo ""
    
    install_fablo
    echo ""
    
    start_fablo_network
    echo ""
    
    deploy_ercc
    echo ""
    
    deploy_chaincode
    echo ""
    
    init_enclave
    echo ""
    
    validate_deployment
    echo ""
    
    print_next_steps
}

# Run main function
main "$@"

# Made with Bob
