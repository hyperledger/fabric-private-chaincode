#!/bin/bash

# Copyright IBM Corp. All Rights Reserved.
# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

# Script to tear down the FPC with Fablo network

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

# Stop FPC chaincode containers
stop_fpc_containers() {
    log_info "Stopping FPC chaincode containers..."
    
    cd "$SAMPLE_DIR"
    
    # Stop echo-go containers
    if [ -f compose/ecc-compose.yaml ]; then
        docker-compose -f compose/ecc-compose.yaml down -v 2>/dev/null || true
        log_success "Echo-go containers stopped"
    fi
    
    # Stop ERCC containers
    if [ -f compose/ercc-compose.yaml ]; then
        docker-compose -f compose/ercc-compose.yaml down -v 2>/dev/null || true
        log_success "ERCC containers stopped"
    fi
}

# Stop Fablo network
stop_fablo_network() {
    log_info "Stopping Fablo network..."
    
    cd "$SAMPLE_DIR"
    
    # Stop network using Fablo
    if command -v fablo &> /dev/null; then
        fablo down 2>/dev/null || true
        log_success "Fablo network stopped"
    else
        log_warning "Fablo command not found, attempting manual cleanup"
        
        # Manual cleanup if Fablo is not available
        if [ -f target/fabric-docker/docker-compose.yaml ]; then
            docker-compose -f target/fabric-docker/docker-compose.yaml down -v 2>/dev/null || true
        fi
    fi
}

# Remove generated artifacts
clean_artifacts() {
    log_info "Cleaning up generated artifacts..."
    
    cd "$SAMPLE_DIR"
    
    # Remove Fablo generated files
    if [ -d target ]; then
        rm -rf target
        log_success "Removed target directory"
    fi
    
    # Remove config files
    if [ -d config/credentials ]; then
        rm -rf config/credentials
        log_success "Removed credentials directory"
    fi
    
    if [ -f config/ercc-env.sh ]; then
        rm -f config/ercc-env.sh
        log_success "Removed ERCC environment file"
    fi
    
    if [ -f config/chaincode-env.sh ]; then
        rm -f config/chaincode-env.sh
        log_success "Removed chaincode environment file"
    fi
    
    # Remove chaincode packages
    if [ -f "$FPC_PATH/ercc/ercc.tar.gz" ]; then
        rm -f "$FPC_PATH/ercc/ercc.tar.gz"
        rm -f "$FPC_PATH/ercc/code.tar.gz"
        rm -f "$FPC_PATH/ercc/connection.json"
        rm -f "$FPC_PATH/ercc/metadata.json"
        log_success "Removed ERCC package files"
    fi
    
    if [ -f "$FPC_PATH/samples/chaincode/echo-go/echo-go.tar.gz" ]; then
        rm -f "$FPC_PATH/samples/chaincode/echo-go/echo-go.tar.gz"
        rm -f "$FPC_PATH/samples/chaincode/echo-go/code.tar.gz"
        rm -f "$FPC_PATH/samples/chaincode/echo-go/connection.json"
        rm -f "$FPC_PATH/samples/chaincode/echo-go/metadata.json"
        log_success "Removed echo-go package files"
    fi
}

# Remove Docker volumes
clean_volumes() {
    log_info "Cleaning up Docker volumes..."
    
    # Remove Fablo network volumes
    docker volume ls -q | grep -E "fablo|fabric" | xargs -r docker volume rm 2>/dev/null || true
    
    log_success "Docker volumes cleaned"
}

# Remove Docker networks
clean_networks() {
    log_info "Cleaning up Docker networks..."
    
    # Remove Fablo network
    docker network rm fablo_default 2>/dev/null || true
    
    log_success "Docker networks cleaned"
}

# Stop and remove all related containers
clean_containers() {
    log_info "Stopping and removing all related containers..."
    
    # Stop and remove FPC containers
    docker ps -a | grep -E "echo-go|ercc|peer|orderer|ca" | awk '{print $1}' | xargs -r docker rm -f 2>/dev/null || true
    
    log_success "Containers cleaned"
}

# Prune Docker system (optional)
prune_docker() {
    if [ "${PRUNE_DOCKER:-false}" = "true" ]; then
        log_info "Pruning Docker system..."
        docker system prune -f
        log_success "Docker system pruned"
    fi
}

# Display cleanup summary
display_summary() {
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}  Teardown Complete${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo "The following have been cleaned up:"
    echo "  ✓ FPC chaincode containers"
    echo "  ✓ Fablo network"
    echo "  ✓ Generated artifacts"
    echo "  ✓ Docker volumes"
    echo "  ✓ Docker networks"
    echo ""
    echo "To start the network again, run:"
    echo -e "  ${BLUE}./scripts/setup.sh${NC}"
    echo ""
}

# Main execution
main() {
    log_info "Starting teardown of FPC with Fablo network..."
    echo ""
    
    # Check if FPC_PATH is set
    if [ -z "${FPC_PATH:-}" ]; then
        log_warning "FPC_PATH not set, some cleanup operations may be skipped"
    fi
    
    stop_fpc_containers
    echo ""
    
    stop_fablo_network
    echo ""
    
    clean_containers
    echo ""
    
    clean_volumes
    echo ""
    
    clean_networks
    echo ""
    
    clean_artifacts
    echo ""
    
    prune_docker
    echo ""
    
    display_summary
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --prune)
            export PRUNE_DOCKER=true
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --prune    Also prune Docker system (removes unused images, etc.)"
            echo "  -h, --help Show this help message"
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            echo "Use -h or --help for usage information"
            exit 1
            ;;
    esac
done

# Run main function
main "$@"

# Made with Bob
