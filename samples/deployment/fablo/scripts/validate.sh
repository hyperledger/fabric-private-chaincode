#!/bin/bash

# Copyright IBM Corp. All Rights Reserved.
# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

# Script to validate FPC with Fablo deployment

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

# Counters
PASSED=0
FAILED=0

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[✓]${NC} $1"
    ((PASSED++))
}

log_error() {
    echo -e "${RED}[✗]${NC} $1"
    ((FAILED++))
}

log_warning() {
    echo -e "${YELLOW}[!]${NC} $1"
}

# Set peer environment for Org1
set_peer_org1() {
    export CORE_PEER_LOCALMSPID="Org1MSP"
    export CORE_PEER_TLS_ENABLED=true
    export CORE_PEER_TLS_ROOTCERT_FILE="$SAMPLE_DIR/target/fabric-config/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"
    export CORE_PEER_MSPCONFIGPATH="$SAMPLE_DIR/target/fabric-config/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp"
    export CORE_PEER_ADDRESS="localhost:7041"
}

# Check Docker containers
check_containers() {
    log_info "Checking Docker containers..."
    
    local REQUIRED_CONTAINERS=(
        "peer0.org1.example.com"
        "peer0.org2.example.com"
        "orderer0.group1.orderer.example.com"
        "ercc.peer0.org1.example.com"
        "ercc.peer0.org2.example.com"
        "echo-go.peer0.org1.example.com"
        "echo-go.peer0.org2.example.com"
    )
    
    for container in "${REQUIRED_CONTAINERS[@]}"; do
        if docker ps --format '{{.Names}}' | grep -q "^${container}$"; then
            log_success "Container running: $container"
        else
            log_error "Container not running: $container"
        fi
    done
    echo ""
}

# Check network connectivity
check_network() {
    log_info "Checking network connectivity..."
    
    set_peer_org1
    
    # Check if peer is reachable
    if peer channel list &> /dev/null; then
        log_success "Peer connectivity OK"
    else
        log_error "Cannot connect to peer"
    fi
    
    # Check if channel exists
    if peer channel list 2>&1 | grep -q "$CHANNEL_NAME"; then
        log_success "Channel exists: $CHANNEL_NAME"
    else
        log_error "Channel not found: $CHANNEL_NAME"
    fi
    echo ""
}

# Check ERCC deployment
check_ercc() {
    log_info "Checking ERCC deployment..."
    
    set_peer_org1
    
    # Check if ERCC is committed
    if peer lifecycle chaincode querycommitted --channelID "$CHANNEL_NAME" --name "$ERCC_ID" &> /dev/null; then
        log_success "ERCC is committed on channel"
        
        # Get ERCC details
        local ERCC_INFO=$(peer lifecycle chaincode querycommitted --channelID "$CHANNEL_NAME" --name "$ERCC_ID" 2>&1)
        echo "$ERCC_INFO" | grep -E "Version|Sequence" | sed 's/^/  /'
    else
        log_error "ERCC is not committed on channel"
    fi
    
    # Test ERCC query
    if peer chaincode query -C "$CHANNEL_NAME" -n "$ERCC_ID" -c '{"Args":["queryChaincodes"]}' &> /dev/null; then
        log_success "ERCC query successful"
    else
        log_error "ERCC query failed"
    fi
    echo ""
}

# Check echo-go chaincode deployment
check_chaincode() {
    log_info "Checking echo-go chaincode deployment..."
    
    set_peer_org1
    
    # Check if chaincode is committed
    if peer lifecycle chaincode querycommitted --channelID "$CHANNEL_NAME" --name "$CC_ID" &> /dev/null; then
        log_success "Echo-go chaincode is committed on channel"
        
        # Get chaincode details
        local CC_INFO=$(peer lifecycle chaincode querycommitted --channelID "$CHANNEL_NAME" --name "$CC_ID" 2>&1)
        echo "$CC_INFO" | grep -E "Version|Sequence" | sed 's/^/  /'
    else
        log_error "Echo-go chaincode is not committed on channel"
    fi
    echo ""
}

# Check enclave registration
check_enclave() {
    log_info "Checking enclave registration..."
    
    set_peer_org1
    
    # Query list of registered enclaves
    local ENCLAVES=$(peer chaincode query -C "$CHANNEL_NAME" -n "$ERCC_ID" -c '{"Args":["queryListEnclaves"]}' 2>&1)
    
    if [ $? -eq 0 ]; then
        if echo "$ENCLAVES" | grep -q "$CC_ID"; then
            log_success "Enclave registered for $CC_ID"
            echo "$ENCLAVES" | sed 's/^/  /'
        else
            log_error "Enclave not registered for $CC_ID"
            log_info "Registered enclaves:"
            echo "$ENCLAVES" | sed 's/^/  /'
        fi
    else
        log_error "Failed to query enclave registry"
    fi
    echo ""
}

# Test chaincode invocation
test_invocation() {
    log_info "Testing chaincode invocation..."
    
    set_peer_org1
    
    local TEST_MSG="Validation test at $(date +%s)"
    local INVOKE_CMD='{"Args":["storeAsset","'$TEST_MSG'"]}'
    
    if peer chaincode invoke \
        -o localhost:7030 \
        --ordererTLSHostnameOverride orderer0.group1.orderer.example.com \
        --tls \
        --cafile "$SAMPLE_DIR/target/fabric-config/crypto-config/ordererOrganizations/orderer.example.com/orderers/orderer0.group1.orderer.example.com/tls/ca.crt" \
        -C "$CHANNEL_NAME" \
        -n "$CC_ID" \
        -c "$INVOKE_CMD" \
        --waitForEvent \
        --peerAddresses localhost:7041 \
        --tlsRootCertFiles "$SAMPLE_DIR/target/fabric-config/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt" \
        &> /dev/null; then
        log_success "Chaincode invocation successful"
    else
        log_error "Chaincode invocation failed"
    fi
    echo ""
}

# Test chaincode query
test_query() {
    log_info "Testing chaincode query..."
    
    set_peer_org1
    
    # Wait a bit for the previous invocation to be processed
    sleep 2
    
    local QUERY_CMD='{"Args":["retrieveAsset"]}'
    local RESULT=$(peer chaincode query -C "$CHANNEL_NAME" -n "$CC_ID" -c "$QUERY_CMD" 2>&1)
    
    if [ $? -eq 0 ]; then
        log_success "Chaincode query successful"
        log_info "Query result: $RESULT"
    else
        log_error "Chaincode query failed"
    fi
    echo ""
}

# Check SGX mode
check_sgx_mode() {
    log_info "Checking SGX configuration..."
    
    if [ -n "${SGX_MODE:-}" ]; then
        log_success "SGX_MODE is set: $SGX_MODE"
    else
        log_warning "SGX_MODE is not set (will default to HW)"
    fi
    
    # Check if running in simulation mode
    if [ "${SGX_MODE:-HW}" = "SIM" ]; then
        log_info "Running in SIMULATION mode (no SGX hardware required)"
    else
        log_info "Running in HARDWARE mode (requires SGX-capable CPU)"
        
        # Check for SGX devices
        if [ -e /dev/sgx_enclave ] && [ -e /dev/sgx_provision ]; then
            log_success "SGX devices found"
        else
            log_warning "SGX devices not found (may cause issues in HW mode)"
        fi
    fi
    echo ""
}

# Display validation summary
display_summary() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}  Validation Summary${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
    echo -e "Tests passed: ${GREEN}$PASSED${NC}"
    echo -e "Tests failed: ${RED}$FAILED${NC}"
    echo ""
    
    if [ $FAILED -eq 0 ]; then
        echo -e "${GREEN}✓ All validation checks passed!${NC}"
        echo ""
        echo "The FPC with Fablo deployment is working correctly."
        echo "You can now use the simple-cli-go client to interact with the chaincode."
        return 0
    else
        echo -e "${RED}✗ Some validation checks failed${NC}"
        echo ""
        echo "Please review the errors above and check:"
        echo "  - All containers are running"
        echo "  - Network is properly configured"
        echo "  - Chaincodes are deployed correctly"
        echo "  - Enclave is initialized and registered"
        return 1
    fi
}

# Main execution
main() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}  FPC with Fablo - Validation${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
    
    check_sgx_mode
    check_containers
    check_network
    check_ercc
    check_chaincode
    check_enclave
    test_invocation
    test_query
    
    display_summary
}

# Run main function
main "$@"

# Made with Bob
