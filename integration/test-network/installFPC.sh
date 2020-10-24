#!/bin/bash

# Copyright IBM Corp. All Rights Reserved.
# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

if [[ -z "${FPC_PATH}"  ]]; then
    echo "Error: FPC_PATH not set"
    exit 1
fi

SAMPLES_PATH=$FPC_PATH/integration/test-network/fabric-samples
NETWORK_PATH=$SAMPLES_PATH/test-network
export FABRIC_CFG_PATH=${SAMPLES_PATH}/config
export PATH=${SAMPLES_PATH}/bin:$PATH

ORDERER="localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile ${NETWORK_PATH}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem"
PEER1="localhost:7051 --tlsRootCertFiles ${NETWORK_PATH}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"
PEER2="localhost:9051 --tlsRootCertFiles ${NETWORK_PATH}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt"

ERCC_EP="AND('Org1MSP.member', 'Org2MSP.member')"
ECC_EP="AND('Org1MSP.member', 'Org2MSP.member')"

CHANNEL_ID=mychannel

# Prepare deployment

PKG_PATH=${FPC_PATH}/integration/test-network/_deployment
. package.sh "${PKG_PATH}"

# test network settings

cd ${NETWORK_PATH} || exit
. ./scripts/envVar.sh
# scripts defined in fail if some vars are not defined (due to 'set -e...')
# -> make sure they are at least defined as empty (but allow override by caller)
OVERRIDE_ORG=${OVERRIDE_ORG:-}
VERBOSE=${VERBOSE:-}

# Install ERCC and ECC

setGlobals 1
peer lifecycle chaincode install $PKG_PATH/ercc.tgz
peer lifecycle chaincode install $PKG_PATH/ecc.tgz
peer lifecycle chaincode queryinstalled

setGlobals 2
peer lifecycle chaincode install $PKG_PATH/ercc.tgz
peer lifecycle chaincode install $PKG_PATH/ecc.tgz
peer lifecycle chaincode queryinstalled

# Get chaincode PGK IDs

ALL_INSTALLED=/tmp/installed_chaincodes
peer lifecycle chaincode queryinstalled > ${ALL_INSTALLED}

PKGID_ERCC=$(cat ${ALL_INSTALLED} | grep ercc | awk '{print $3}' | sed 's/.$//')
PKGID_ECC=$(cat ${ALL_INSTALLED}  | grep ecc | awk '{print $3}' | sed 's/.$//')

# Approve

setGlobals 2
peer lifecycle chaincode approveformyorg -o ${ORDERER} --channelID ${CHANNEL_ID} --name ercc --signature-policy "${ERCC_EP}" --version 1.0 --package-id ${PKGID_ERCC} --sequence 1
peer lifecycle chaincode approveformyorg -o ${ORDERER} --channelID ${CHANNEL_ID} --name ecc --signature-policy "${ECC_EP}" --version 1.0 --package-id ${PKGID_ECC} --sequence 1

setGlobals 1
peer lifecycle chaincode approveformyorg -o ${ORDERER} --channelID ${CHANNEL_ID} --name ercc --signature-policy "${ERCC_EP}" --version 1.0 --package-id ${PKGID_ERCC} --sequence 1
peer lifecycle chaincode approveformyorg -o ${ORDERER} --channelID ${CHANNEL_ID} --name ecc --signature-policy "${ECC_EP}" --version 1.0 --package-id ${PKGID_ECC} --sequence 1

# Commit

peer lifecycle chaincode commit -o ${ORDERER} --channelID ${CHANNEL_ID} --name ercc --signature-policy "${ERCC_EP}" --peerAddresses ${PEER1} --peerAddresses ${PEER2} --version 1.0 --sequence 1
peer lifecycle chaincode commit -o ${ORDERER} --channelID ${CHANNEL_ID} --name ecc --signature-policy "${ECC_EP}" --peerAddresses ${PEER1} --peerAddresses ${PEER2} --version 1.0 --sequence 1

# Show chaincodes

echo "${PKGID_ERCC}"
peer lifecycle chaincode querycommitted --output json -o ${ORDERER} --channelID ${CHANNEL_ID} --name ercc
echo "${PKGID_ECC}"
peer lifecycle chaincode querycommitted --output json -o ${ORDERER} --channelID ${CHANNEL_ID} --name ecc
