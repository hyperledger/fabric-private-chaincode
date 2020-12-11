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

CC_ID="echo"
CC_VER="$(cat ${FPC_PATH}/examples/${CC_ID}/_build/lib/mrenclave)"

ERCC_ID="ercc"
ERCC_VER="1.0"

: ${SEQ_NUM:=1}


CHANNEL_ID=mychannel

# Prepare deployment
PKG_PATH=${FPC_PATH}/integration/test-network/_deployment

. package.sh "${PKG_PATH}" "${ERCC_ID}" "${ERCC_VER}" "${CC_ID}" "${CC_VER}"

# test network settings

cd ${NETWORK_PATH} || exit
. ./scripts/envVar.sh
# scripts defined in fail if some vars are not defined (due to 'set -e...')
# -> make sure they are at least defined as empty (but allow override by caller)
OVERRIDE_ORG=${OVERRIDE_ORG:-}
VERBOSE=${VERBOSE:-}

# Install ERCC and ECC

# Org 1
setGlobals 1
peer lifecycle chaincode install $PKG_PATH/${ERCC_ID}.tgz
peer lifecycle chaincode install $PKG_PATH/${CC_ID}.tgz
peer lifecycle chaincode queryinstalled
ORG1_ALL_INSTALLED=/tmp/installed_chaincodes.org1
peer lifecycle chaincode queryinstalled > ${ORG1_ALL_INSTALLED}
ORG1_ERCC_PKG_ID=$(cat ${ORG1_ALL_INSTALLED} | grep ${ERCC_ID} | awk '{print $3}' | sed 's/.$//')
ORG1_ECC_PKG_ID=$(cat ${ORG1_ALL_INSTALLED}  | grep ${CC_ID} | awk '{print $3}' | sed 's/.$//')

# Org 2
setGlobals 2
peer lifecycle chaincode install $PKG_PATH/${ERCC_ID}.tgz
peer lifecycle chaincode install $PKG_PATH/${CC_ID}.tgz
peer lifecycle chaincode queryinstalled
ORG2_ALL_INSTALLED=/tmp/installed_chaincodes.org2
peer lifecycle chaincode queryinstalled > ${ORG2_ALL_INSTALLED}
ORG2_ERCC_PKG_ID=$(cat ${ORG2_ALL_INSTALLED} | grep ${ERCC_ID} | awk '{print $3}' | sed 's/.$//')
ORG2_ECC_PKG_ID=$(cat ${ORG2_ALL_INSTALLED}  | grep ${CC_ID} | awk '{print $3}' | sed 's/.$//')


# Approve

setGlobals 2
peer lifecycle chaincode approveformyorg -o ${ORDERER} --channelID ${CHANNEL_ID} --name ${ERCC_ID} --signature-policy "${ERCC_EP}" --version ${ERCC_VER} --package-id ${ORG2_ERCC_PKG_ID} --sequence ${SEQ_NUM}
peer lifecycle chaincode approveformyorg -o ${ORDERER} --channelID ${CHANNEL_ID} --name ${CC_ID} --signature-policy "${ECC_EP}" --version ${CC_VER} --package-id ${ORG2_ECC_PKG_ID} --sequence ${SEQ_NUM}

setGlobals 1
peer lifecycle chaincode approveformyorg -o ${ORDERER} --channelID ${CHANNEL_ID} --name ${ERCC_ID} --signature-policy "${ERCC_EP}" --version ${ERCC_VER} --package-id ${ORG1_ERCC_PKG_ID} --sequence ${SEQ_NUM}
peer lifecycle chaincode approveformyorg -o ${ORDERER} --channelID ${CHANNEL_ID} --name ${CC_ID} --signature-policy "${ECC_EP}" --version ${CC_VER} --package-id ${ORG1_ECC_PKG_ID} --sequence ${SEQ_NUM}

# Commit

peer lifecycle chaincode commit -o ${ORDERER} --channelID ${CHANNEL_ID} --name ${ERCC_ID} --signature-policy "${ERCC_EP}" --peerAddresses ${PEER1} --peerAddresses ${PEER2} --version ${ERCC_VER} --sequence ${SEQ_NUM}
peer lifecycle chaincode commit -o ${ORDERER} --channelID ${CHANNEL_ID} --name ${CC_ID} --signature-policy "${ECC_EP}" --peerAddresses ${PEER1} --peerAddresses ${PEER2} --version ${CC_VER} --sequence ${SEQ_NUM}

# Show chaincodes

peer lifecycle chaincode querycommitted --output json -o ${ORDERER} --channelID ${CHANNEL_ID} --name ${ERCC_ID}
peer lifecycle chaincode querycommitted --output json -o ${ORDERER} --channelID ${CHANNEL_ID} --name ${CC_ID}

cat <<EOF
# define following environment-variables for docker-compose:
export \\
  ORG1_ECC_PKG_ID=${ORG1_ECC_PKG_ID}\\
  ORG1_ERCC_PKG_ID=${ORG1_ERCC_PKG_ID}\\
  ORG2_ECC_PKG_ID=${ORG2_ECC_PKG_ID}\\
  ORG2_ERCC_PKG_ID=${ORG2_ERCC_PKG_ID}
EOF
