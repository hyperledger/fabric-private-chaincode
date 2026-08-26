#!/bin/bash

# Copyright IBM Corp. All Rights Reserved.
# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

# Script to deploy FPC echo-go chaincode

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
CC_ID="${CC_ID:-echo-go}"
CC_VERSION="${CC_VERSION:-1.0}"
CC_SEQUENCE="${CC_SEQUENCE:-1}"
CC_VER="${CC_VER:-latest}"

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

# Package echo-go chaincode
package_chaincode() {
    log_info "Packaging echo-go chaincode..."
    
    cd "$FPC_PATH/samples/chaincode/echo-go"
    
    # Create connection.json for chaincode as a service
    cat > connection.json <<EOF
{
  "address": "echo-go.peer0.org1.example.com:9999",
  "dial_timeout": "10s",
  "tls_required": false
}
EOF
    
    # Create metadata.json
    cat > metadata.json <<EOF
{
  "type": "external",
  "label": "${CC_ID}_${CC_VERSION}"
}
EOF
    
    # Package the chaincode
    tar czf code.tar.gz connection.json
    tar czf "${CC_ID}.tar.gz" metadata.json code.tar.gz
    
    log_success "Echo-go chaincode packaged: ${CC_ID}.tar.gz"
}

# Install chaincode on Org1 peer
install_chaincode_org1() {
    log_info "Installing echo-go chaincode on Org1 peer..."
    
    set_peer_org1
    
    cd "$FPC_PATH/samples/chaincode/echo-go"
    
    peer lifecycle chaincode install "${CC_ID}.tar.gz"
    
    # Get package ID
    export CC_PKG_ID_ORG1=$(peer lifecycle chaincode queryinstalled | grep "${CC_ID}_${CC_VERSION}" | awk '{print $3}' | sed 's/,$//')
    
    if [ -z "$CC_PKG_ID_ORG1" ]; then
        log_error "Failed to get chaincode package ID for Org1"
        exit 1
    fi
    
    log_success "Echo-go chaincode installed on Org1: $CC_PKG_ID_ORG1"
    
    # Save package ID for docker-compose
    echo "export CC_PKG_ID_ORG1=$CC_PKG_ID_ORG1" >> "$SAMPLE_DIR/config/chaincode-env.sh"
}

# Install chaincode on Org2 peer
install_chaincode_org2() {
    log_info "Installing echo-go chaincode on Org2 peer..."
    
    set_peer_org2
    
    cd "$FPC_PATH/samples/chaincode/echo-go"
    
    peer lifecycle chaincode install "${CC_ID}.tar.gz"
    
    # Get package ID
    export CC_PKG_ID_ORG2=$(peer lifecycle chaincode queryinstalled | grep "${CC_ID}_${CC_VERSION}" | awk '{print $3}' | sed 's/,$//')
    
    if [ -z "$CC_PKG_ID_ORG2" ]; then
        log_error "Failed to get chaincode package ID for Org2"
        exit 1
    fi
    
    log_success "Echo-go chaincode installed on Org2: $CC_PKG_ID_ORG2"
    
    # Save package ID for docker-compose
    echo "export CC_PKG_ID_ORG2=$CC_PKG_ID_ORG2" >> "$SAMPLE_DIR/config/chaincode-env.sh"
}

# Start chaincode containers
start_chaincode_containers() {
    log_info "Starting echo-go chaincode containers..."
    
    # Source the package IDs
    source "$SAMPLE_DIR/config/chaincode-env.sh"
    
    # Export CC_VER for docker-compose
    export CC_VER
    
    # Start chaincode containers using docker-compose
    cd "$SAMPLE_DIR"
    docker-compose -f compose/ecc-compose.yaml up -d
    
    # Wait for containers to be ready
    sleep 5
    
    log_success "Echo-go chaincode containers started"
}

# Approve chaincode for Org1
approve_chaincode_org1() {
    log_info "Approving echo-go chaincode for Org1..."
    
    set_peer_org1
    source "$SAMPLE_DIR/config/chaincode-env.sh"
    
    peer lifecycle chaincode approveformyorg \
        --channelID "$CHANNEL_NAME" \
        --name "$CC_ID" \
        --version "$CC_VERSION" \
        --package-id "$CC_PKG_ID_ORG1" \
        --sequence "$CC_SEQUENCE" \
        --tls \
        --cafile "$SAMPLE_DIR/target/fabric-config/crypto-config/ordererOrganizations/orderer.example.com/orderers/orderer0.group1.orderer.example.com/tls/ca.crt" \
        --signature-policy "OR('Org1MSP.member','Org2MSP.member')"
    
    log_success "Echo-go chaincode approved for Org1"
}

# Approve chaincode for Org2
approve_chaincode_org2() {
    log_info "Approving echo-go chaincode for Org2..."
    
    set_peer_org2
    source "$SAMPLE_DIR/config/chaincode-env.sh"
    
    peer lifecycle chaincode approveformyorg \
        --channelID "$CHANNEL_NAME" \
        --name "$CC_ID" \
        --version "$CC_VERSION" \
        --package-id "$CC_PKG_ID_ORG2" \
        --sequence "$CC_SEQUENCE" \
        --tls \
        --cafile "$SAMPLE_DIR/target/fabric-config/crypto-config/ordererOrganizations/orderer.example.com/orderers/orderer0.group1.orderer.example.com/tls/ca.crt" \
        --signature-policy "OR('Org1MSP.member','Org2MSP.member')"
    
    log_success "Echo-go chaincode approved for Org2"
}

# Commit chaincode
commit_chaincode() {
    log_info "Committing echo-go chaincode to channel..."
    
    set_peer_org1
    
    peer lifecycle chaincode commit \
        --channelID "$CHANNEL_NAME" \
        --name "$CC_ID" \
        --version "$CC_VERSION" \
        --sequence "$CC_SEQUENCE" \
        --tls \
        --cafile "$SAMPLE_DIR/target/fabric-config/crypto-config/ordererOrganizations/orderer.example.com/orderers/orderer0.group1.orderer.example.com/tls/ca.crt" \
        --peerAddresses localhost:7041 \
        --tlsRootCertFiles "$SAMPLE_DIR/target/fabric-config/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt" \
        --peerAddresses localhost:7061 \
        --tlsRootCertFiles "$SAMPLE_DIR/target/fabric-config/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt" \
        --signature-policy "OR('Org1MSP.member','Org2MSP.member')"
    
    log_success "Echo-go chaincode committed to channel"
}

# Verify chaincode deployment
verify_chaincode() {
    log_info "Verifying echo-go chaincode deployment..."
    
    set_peer_org1
    
    # Query committed chaincode
    peer lifecycle chaincode querycommitted --channelID "$CHANNEL_NAME" --name "$CC_ID"
    
    log_success "Echo-go chaincode deployment verified"
}

# Main execution
main() {
    log_info "Starting echo-go chaincode deployment..."
    
    # Create config directory if it doesn't exist
    mkdir -p "$SAMPLE_DIR/config"
    
    # Clear previous chaincode environment file
    > "$SAMPLE_DIR/config/chaincode-env.sh"
    
    package_chaincode
    install_chaincode_org1
    install_chaincode_org2
    start_chaincode_containers
    
    # Wait for containers to be fully ready
    sleep 10
    
    approve_chaincode_org1
    approve_chaincode_org2
    commit_chaincode
    
    # Wait for commit to complete
    sleep 5
    
    verify_chaincode
    
    log_success "Echo-go chaincode deployment completed successfully"
}

# Run main function
main "$@"

# Made with Bob
