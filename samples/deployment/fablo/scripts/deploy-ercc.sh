#!/bin/bash

# Copyright IBM Corp. All Rights Reserved.
# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

# Script to deploy FPC Enclave Registry Chaincode (ERCC)

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

# Configuration
CHANNEL_NAME="${CHANNEL_NAME:-mychannel}"
ERCC_ID="ercc"
ERCC_VERSION="${ERCC_VERSION:-1.0}"
ERCC_SEQUENCE="${ERCC_SEQUENCE:-1}"

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

# Set peer environment for Org1
set_peer_org1() {
    export CORE_PEER_LOCALMSPID="Org1MSP"
    export CORE_PEER_TLS_ENABLED=true
    export CORE_PEER_TLS_ROOTCERT_FILE="$SAMPLE_DIR/target/fabric-config/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"
    export CORE_PEER_MSPCONFIGPATH="$SAMPLE_DIR/target/fabric-config/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp"
    export CORE_PEER_ADDRESS="localhost:7041"
}

# Set peer environment for Org2
set_peer_org2() {
    export CORE_PEER_LOCALMSPID="Org2MSP"
    export CORE_PEER_TLS_ENABLED=true
    export CORE_PEER_TLS_ROOTCERT_FILE="$SAMPLE_DIR/target/fabric-config/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt"
    export CORE_PEER_MSPCONFIGPATH="$SAMPLE_DIR/target/fabric-config/crypto-config/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp"
    export CORE_PEER_ADDRESS="localhost:7061"
}

# Package ERCC
package_ercc() {
    log_info "Packaging ERCC..."
    
    cd "$FPC_PATH/ercc"
    
    # Create connection.json for chaincode as a service
    cat > connection.json <<EOF
{
  "address": "ercc.peer0.org1.example.com:9999",
  "dial_timeout": "10s",
  "tls_required": false
}
EOF
    
    # Create metadata.json
    cat > metadata.json <<EOF
{
  "type": "external",
  "label": "${ERCC_ID}_${ERCC_VERSION}"
}
EOF
    
    # Package the chaincode
    tar czf code.tar.gz connection.json
    tar czf "${ERCC_ID}.tar.gz" metadata.json code.tar.gz
    
    log_success "ERCC packaged: ${ERCC_ID}.tar.gz"
}

# Install ERCC on Org1 peer
install_ercc_org1() {
    log_info "Installing ERCC on Org1 peer..."
    
    set_peer_org1
    
    cd "$FPC_PATH/ercc"
    
    peer lifecycle chaincode install "${ERCC_ID}.tar.gz"
    
    # Get package ID
    export ERCC_PKG_ID_ORG1=$(peer lifecycle chaincode queryinstalled | grep "${ERCC_ID}_${ERCC_VERSION}" | awk '{print $3}' | sed 's/,$//')
    
    if [ -z "$ERCC_PKG_ID_ORG1" ]; then
        log_error "Failed to get ERCC package ID for Org1"
        exit 1
    fi
    
    log_success "ERCC installed on Org1: $ERCC_PKG_ID_ORG1"
    
    # Save package ID for docker-compose
    echo "export ERCC_PKG_ID_ORG1=$ERCC_PKG_ID_ORG1" >> "$SAMPLE_DIR/config/ercc-env.sh"
}

# Install ERCC on Org2 peer
install_ercc_org2() {
    log_info "Installing ERCC on Org2 peer..."
    
    set_peer_org2
    
    cd "$FPC_PATH/ercc"
    
    peer lifecycle chaincode install "${ERCC_ID}.tar.gz"
    
    # Get package ID
    export ERCC_PKG_ID_ORG2=$(peer lifecycle chaincode queryinstalled | grep "${ERCC_ID}_${ERCC_VERSION}" | awk '{print $3}' | sed 's/,$//')
    
    if [ -z "$ERCC_PKG_ID_ORG2" ]; then
        log_error "Failed to get ERCC package ID for Org2"
        exit 1
    fi
    
    log_success "ERCC installed on Org2: $ERCC_PKG_ID_ORG2"
    
    # Save package ID for docker-compose
    echo "export ERCC_PKG_ID_ORG2=$ERCC_PKG_ID_ORG2" >> "$SAMPLE_DIR/config/ercc-env.sh"
}

# Start ERCC containers
start_ercc_containers() {
    log_info "Starting ERCC containers..."
    
    # Source the package IDs
    source "$SAMPLE_DIR/config/ercc-env.sh"
    
    # Start ERCC containers using docker-compose
    cd "$SAMPLE_DIR"
    docker-compose -f compose/ercc-compose.yaml up -d
    
    # Wait for containers to be ready
    sleep 5
    
    log_success "ERCC containers started"
}

# Approve ERCC for Org1
approve_ercc_org1() {
    log_info "Approving ERCC for Org1..."
    
    set_peer_org1
    source "$SAMPLE_DIR/config/ercc-env.sh"
    
    peer lifecycle chaincode approveformyorg \
        --channelID "$CHANNEL_NAME" \
        --name "$ERCC_ID" \
        --version "$ERCC_VERSION" \
        --package-id "$ERCC_PKG_ID_ORG1" \
        --sequence "$ERCC_SEQUENCE" \
        --tls \
        --cafile "$SAMPLE_DIR/target/fabric-config/crypto-config/ordererOrganizations/orderer.example.com/orderers/orderer0.group1.orderer.example.com/tls/ca.crt" \
        --signature-policy "OR('Org1MSP.member','Org2MSP.member')"
    
    log_success "ERCC approved for Org1"
}

# Approve ERCC for Org2
approve_ercc_org2() {
    log_info "Approving ERCC for Org2..."
    
    set_peer_org2
    source "$SAMPLE_DIR/config/ercc-env.sh"
    
    peer lifecycle chaincode approveformyorg \
        --channelID "$CHANNEL_NAME" \
        --name "$ERCC_ID" \
        --version "$ERCC_VERSION" \
        --package-id "$ERCC_PKG_ID_ORG2" \
        --sequence "$ERCC_SEQUENCE" \
        --tls \
        --cafile "$SAMPLE_DIR/target/fabric-config/crypto-config/ordererOrganizations/orderer.example.com/orderers/orderer0.group1.orderer.example.com/tls/ca.crt" \
        --signature-policy "OR('Org1MSP.member','Org2MSP.member')"
    
    log_success "ERCC approved for Org2"
}

# Commit ERCC
commit_ercc() {
    log_info "Committing ERCC to channel..."
    
    set_peer_org1
    
    peer lifecycle chaincode commit \
        --channelID "$CHANNEL_NAME" \
        --name "$ERCC_ID" \
        --version "$ERCC_VERSION" \
        --sequence "$ERCC_SEQUENCE" \
        --tls \
        --cafile "$SAMPLE_DIR/target/fabric-config/crypto-config/ordererOrganizations/orderer.example.com/orderers/orderer0.group1.orderer.example.com/tls/ca.crt" \
        --peerAddresses localhost:7041 \
        --tlsRootCertFiles "$SAMPLE_DIR/target/fabric-config/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt" \
        --peerAddresses localhost:7061 \
        --tlsRootCertFiles "$SAMPLE_DIR/target/fabric-config/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt" \
        --signature-policy "OR('Org1MSP.member','Org2MSP.member')"
    
    log_success "ERCC committed to channel"
}

# Verify ERCC deployment
verify_ercc() {
    log_info "Verifying ERCC deployment..."
    
    set_peer_org1
    
    # Query committed chaincode
    peer lifecycle chaincode querycommitted --channelID "$CHANNEL_NAME" --name "$ERCC_ID"
    
    log_success "ERCC deployment verified"
}

# Main execution
main() {
    log_info "Starting ERCC deployment..."
    
    # Create config directory if it doesn't exist
    mkdir -p "$SAMPLE_DIR/config"
    
    # Clear previous ERCC environment file
    > "$SAMPLE_DIR/config/ercc-env.sh"
    
    package_ercc
    install_ercc_org1
    install_ercc_org2
    start_ercc_containers
    
    # Wait for containers to be fully ready
    sleep 10
    
    approve_ercc_org1
    approve_ercc_org2
    commit_ercc
    
    # Wait for commit to complete
    sleep 5
    
    verify_ercc
    
    log_success "ERCC deployment completed successfully"
}

# Run main function
main "$@"

# Made with Bob
