#!/bin/bash

# Copyright IBM Corp. All Rights Reserved.
# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

# Script to initialize FPC enclave for echo-go chaincode

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
ERCC_ID="ercc"

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

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Set peer environment for Org1
set_peer_org1() {
    export CORE_PEER_LOCALMSPID="Org1MSP"
    export CORE_PEER_TLS_ENABLED=true
    export CORE_PEER_TLS_ROOTCERT_FILE="$SAMPLE_DIR/target/fabric-config/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"
    export CORE_PEER_MSPCONFIGPATH="$SAMPLE_DIR/target/fabric-config/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp"
    export CORE_PEER_ADDRESS="localhost:7041"
}

# Initialize enclave
init_enclave() {
    log_info "Initializing FPC enclave for $CC_ID..."
    
    set_peer_org1
    
    # Call initEnclave function via ERCC
    local INIT_CMD='{"Args":["initEnclave"]}'
    
    peer chaincode invoke \
        -o localhost:7030 \
        --ordererTLSHostnameOverride orderer0.group1.orderer.example.com \
        --tls \
        --cafile "$SAMPLE_DIR/target/fabric-config/crypto-config/ordererOrganizations/orderer.example.com/orderers/orderer0.group1.orderer.example.com/tls/ca.crt" \
        -C "$CHANNEL_NAME" \
        -n "$ERCC_ID" \
        -c "$INIT_CMD" \
        --waitForEvent \
        --peerAddresses localhost:7041 \
        --tlsRootCertFiles "$SAMPLE_DIR/target/fabric-config/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"
    
    log_success "Enclave initialization invoked"
    
    # Wait for enclave to initialize
    log_info "Waiting for enclave to initialize..."
    sleep 10
}

# Get enclave credentials
get_enclave_credentials() {
    log_info "Retrieving enclave credentials..."
    
    set_peer_org1
    
    # Query for enclave credentials
    local QUERY_CMD='{"Args":["queryChaincodeEncryptionKey","'$CC_ID'"]}'
    
    local CREDENTIALS=$(peer chaincode query \
        -C "$CHANNEL_NAME" \
        -n "$ERCC_ID" \
        -c "$QUERY_CMD" 2>&1)
    
    if [ $? -eq 0 ] && [ -n "$CREDENTIALS" ]; then
        log_success "Enclave credentials retrieved"
        
        # Save credentials to file
        mkdir -p "$SAMPLE_DIR/config/credentials"
        echo "$CREDENTIALS" > "$SAMPLE_DIR/config/credentials/${CC_ID}_credentials.json"
        
        log_info "Credentials saved to: $SAMPLE_DIR/config/credentials/${CC_ID}_credentials.json"
    else
        log_warning "Could not retrieve enclave credentials yet. This is normal if using simulation mode."
    fi
}

# Register enclave with ERCC
register_enclave() {
    log_info "Registering enclave with ERCC..."
    
    set_peer_org1
    
    # Call registerEnclave function via ERCC
    local REGISTER_CMD='{"Args":["registerEnclave"]}'
    
    peer chaincode invoke \
        -o localhost:7030 \
        --ordererTLSHostnameOverride orderer0.group1.orderer.example.com \
        --tls \
        --cafile "$SAMPLE_DIR/target/fabric-config/crypto-config/ordererOrganizations/orderer.example.com/orderers/orderer0.group1.orderer.example.com/tls/ca.crt" \
        -C "$CHANNEL_NAME" \
        -n "$ERCC_ID" \
        -c "$REGISTER_CMD" \
        --waitForEvent \
        --peerAddresses localhost:7041 \
        --tlsRootCertFiles "$SAMPLE_DIR/target/fabric-config/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"
    
    log_success "Enclave registration invoked"
    
    # Wait for registration to complete
    sleep 5
}

# Verify enclave registration
verify_enclave_registration() {
    log_info "Verifying enclave registration..."
    
    set_peer_org1
    
    # Query list of registered enclaves
    local QUERY_CMD='{"Args":["queryListEnclaves"]}'
    
    local ENCLAVES=$(peer chaincode query \
        -C "$CHANNEL_NAME" \
        -n "$ERCC_ID" \
        -c "$QUERY_CMD" 2>&1)
    
    if echo "$ENCLAVES" | grep -q "$CC_ID"; then
        log_success "Enclave successfully registered for $CC_ID"
        echo "$ENCLAVES"
    else
        log_error "Enclave not found in registry"
        log_info "Registered enclaves:"
        echo "$ENCLAVES"
        exit 1
    fi
}

# Test basic chaincode invocation
test_chaincode() {
    log_info "Testing chaincode invocation..."
    
    set_peer_org1
    
    # Try a simple echo invocation
    local TEST_MSG="Test message from init-enclave script"
    local INVOKE_CMD='{"Args":["storeAsset","'$TEST_MSG'"]}'
    
    peer chaincode invoke \
        -o localhost:7030 \
        --ordererTLSHostnameOverride orderer0.group1.orderer.example.com \
        --tls \
        --cafile "$SAMPLE_DIR/target/fabric-config/crypto-config/ordererOrganizations/orderer.example.com/orderers/orderer0.group1.orderer.example.com/tls/ca.crt" \
        -C "$CHANNEL_NAME" \
        -n "$CC_ID" \
        -c "$INVOKE_CMD" \
        --waitForEvent \
        --peerAddresses localhost:7041 \
        --tlsRootCertFiles "$SAMPLE_DIR/target/fabric-config/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"
    
    if [ $? -eq 0 ]; then
        log_success "Chaincode invocation successful"
        
        # Query the result
        sleep 2
        local QUERY_CMD='{"Args":["retrieveAsset"]}'
        local RESULT=$(peer chaincode query \
            -C "$CHANNEL_NAME" \
            -n "$CC_ID" \
            -c "$QUERY_CMD" 2>&1)
        
        log_info "Query result: $RESULT"
    else
        log_warning "Chaincode invocation failed. This might be expected if additional setup is needed."
    fi
}

# Main execution
main() {
    log_info "Starting FPC enclave initialization for $CC_ID..."
    echo ""
    
    init_enclave
    echo ""
    
    get_enclave_credentials
    echo ""
    
    register_enclave
    echo ""
    
    verify_enclave_registration
    echo ""
    
    test_chaincode
    echo ""
    
    log_success "FPC enclave initialization completed successfully"
    echo ""
    log_info "You can now use the simple-cli-go client to interact with the chaincode"
}

# Run main function
main "$@"

# Made with Bob
