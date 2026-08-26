#!/bin/bash

# Copyright IBM Corp. All Rights Reserved.
# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

# Environment variables for FPC client (simple-cli-go)
# Source this file before using the fpcclient:
#   source config/fpcclient-env.sh

# Get the sample directory
SAMPLE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# FPC Configuration
export FPC_PATH="${FPC_PATH:-}"
if [ -z "$FPC_PATH" ]; then
    echo "Warning: FPC_PATH is not set. Please set it to your FPC repository path."
    echo "Example: export FPC_PATH=/path/to/fabric-private-chaincode"
fi

# Chaincode Configuration
export CC_ID="${CC_ID:-echo-go}"
export CHANNEL_NAME="${CHANNEL_NAME:-mychannel}"

# SGX Configuration
export SGX_MODE="${SGX_MODE:-SIM}"

# Peer Configuration (Org1)
export CORE_PEER_ID="peer0.org1.example.com"
export CORE_PEER_ADDRESS="localhost:7041"
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_TLS_ROOTCERT_FILE="$SAMPLE_DIR/target/fabric-config/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"
export CORE_PEER_MSPCONFIGPATH="$SAMPLE_DIR/target/fabric-config/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp"

# Gateway Configuration
export GATEWAY_CONFIG="$SAMPLE_DIR/config/connection-org1.yaml"

# Orderer Configuration
export ORDERER_ADDR="localhost:7030"
export ORDERER_CA="$SAMPLE_DIR/target/fabric-config/crypto-config/ordererOrganizations/orderer.example.com/orderers/orderer0.group1.orderer.example.com/tls/ca.crt"

# SGX Credentials Path (for hardware mode)
if [ "$SGX_MODE" = "HW" ]; then
    export SGX_CREDENTIALS_PATH="${SGX_CREDENTIALS_PATH:-$FPC_PATH/config/ias}"
fi

# Display configuration
echo "FPC Client Environment Variables:"
echo "  FPC_PATH:         $FPC_PATH"
echo "  CC_ID:            $CC_ID"
echo "  CHANNEL_NAME:     $CHANNEL_NAME"
echo "  SGX_MODE:         $SGX_MODE"
echo "  CORE_PEER_ID:     $CORE_PEER_ID"
echo "  CORE_PEER_ADDRESS: $CORE_PEER_ADDRESS"
echo "  GATEWAY_CONFIG:   $GATEWAY_CONFIG"
echo ""
echo "Ready to use fpcclient!"
echo "Navigate to: cd \$FPC_PATH/samples/application/simple-cli-go"
echo "Then run: ./fpcclient init $CORE_PEER_ID"

# Made with Bob
