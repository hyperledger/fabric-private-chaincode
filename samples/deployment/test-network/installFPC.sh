#!/bin/bash

# Copyright IBM Corp. All Rights Reserved.
# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

#DEBUG=true # uncomment (or define when calling script) to show debug output
# Note: to see only log and not peer command output (which is all on stderr),
#       run the script with '... 2>/dev/null' ..

if [[ -z "${FPC_PATH}"  ]]; then
    echo "Error: FPC_PATH not set"
    exit 1
fi
if [[ -z "${CC_ID}"  ]]; then
    echo "Error: CC_ID not set"
    exit 1
fi
if [[ -z "${CC_PATH}"  ]]; then
    echo "Error: CC_PATH not set"
    exit 1
fi
CHANNEL_ID=mychannel

PEERS=("peer0.org1.example.com" "peer0.org2.example.com")

ERCC_EP="OutOf(2, 'Org1MSP.peer', 'Org2MSP.peer')"
ECC_EP="OutOf(2, 'Org1MSP.peer', 'Org2MSP.peer')"

CC_VER="${CC_VER:-$(cat "${CC_PATH}/_build/lib/mrenclave")}"

ERCC_ID="ercc"
ERCC_VER="1.0"

: ${SEQ_NUM:=1}



# Prepare
#------------

TEST_NET_SCRIPT_PATH=${FPC_PATH}/samples/deployment/test-network
SAMPLES_PATH=${TEST_NET_SCRIPT_PATH}/fabric-samples
NETWORK_PATH=${SAMPLES_PATH}/test-network
export FABRIC_CFG_PATH=${SAMPLES_PATH}/config
export PATH=${SAMPLES_PATH}/bin:$PATH

# Orderer address and certs, including hostname override as we access via localhost
ORDERER_ARGS="-o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile ${NETWORK_PATH}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem"

# Commit (but not approveformyorg) requires TLS Root CAs for all peers
PEER1_ARGS="--peerAddresses localhost:7051 --tlsRootCertFiles ${NETWORK_PATH}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"
PEER2_ARGS="--peerAddresses localhost:9051 --tlsRootCertFiles ${NETWORK_PATH}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt"
PEER_ARGS="${PEER1_ARGS} ${PEER2_ARGS}"

# read test-network settings

cd ${NETWORK_PATH} || exit
. ./scripts/envVar.sh
# scripts defined in fail if some vars are not defined (due to 'set -e...')
# -> make sure they are at least defined as empty (but allow override by caller)
OVERRIDE_ORG=${OVERRIDE_ORG:-}
VERBOSE=${VERBOSE:-}



# Install ERCC and ECC
#---------------------
echo ""
echo "Packaging chaincodes: ERCC (id=${ERCC_ID}/version=${ERCC_VER}) and ECC (id=${CC_ID}/version=${CC_VER}) for peers ${PEERS[@]}"

PKG_PATH=${TEST_NET_SCRIPT_PATH}/packages

# as Org 1
${TEST_NET_SCRIPT_PATH}/package.sh "${PKG_PATH}" "${ERCC_ID}" "${ERCC_VER}" "${CC_ID}" "${CC_VER}" "${PEERS[0]}"

# as Org 2
${TEST_NET_SCRIPT_PATH}/package.sh "${PKG_PATH}" "${ERCC_ID}" "${ERCC_VER}" "${CC_ID}" "${CC_VER}" "${PEERS[1]}"


# Install ERCC and ECC
#---------------------
echo ""
echo "Installing chaincodes"

# as Org 1
setGlobals 1
peer lifecycle chaincode install $PKG_PATH/${ERCC_ID}.${PEERS[0]}.tgz
peer lifecycle chaincode install $PKG_PATH/${CC_ID}.${PEERS[0]}.tgz
peer lifecycle chaincode queryinstalled
ORG1_ALL_INSTALLED=/tmp/installed_chaincodes.org1
peer lifecycle chaincode queryinstalled > ${ORG1_ALL_INSTALLED}
ORG1_ERCC_PKG_ID=$(cat ${ORG1_ALL_INSTALLED} | grep ${ERCC_ID} | awk '{print $3}' | sed 's/.$//')
ORG1_ECC_PKG_ID=$(cat ${ORG1_ALL_INSTALLED}  | grep ${CC_ID} | awk '{print $3}' | sed 's/.$//')

# as Org 2
setGlobals 2
peer lifecycle chaincode install $PKG_PATH/${ERCC_ID}.${PEERS[1]}.tgz
peer lifecycle chaincode install $PKG_PATH/${CC_ID}.${PEERS[1]}.tgz
peer lifecycle chaincode queryinstalled
ORG2_ALL_INSTALLED=/tmp/installed_chaincodes.org2
peer lifecycle chaincode queryinstalled > ${ORG2_ALL_INSTALLED}
ORG2_ERCC_PKG_ID=$(cat ${ORG2_ALL_INSTALLED} | grep ${ERCC_ID} | awk '{print $3}' | sed 's/.$//')
ORG2_ECC_PKG_ID=$(cat ${ORG2_ALL_INSTALLED}  | grep ${CC_ID} | awk '{print $3}' | sed 's/.$//')


# Approve
#------------
echo ""
echo "Approving both chaincodes"

# as Org 1
setGlobals 1
peer lifecycle chaincode approveformyorg ${ORDERER_ARGS} --channelID ${CHANNEL_ID} --name ${ERCC_ID} --signature-policy "${ERCC_EP}" --version ${ERCC_VER} --package-id ${ORG1_ERCC_PKG_ID} --sequence ${SEQ_NUM}
peer lifecycle chaincode approveformyorg ${ORDERER_ARGS} --channelID ${CHANNEL_ID} --name ${CC_ID} --signature-policy "${ECC_EP}" --version ${CC_VER} --package-id ${ORG1_ECC_PKG_ID} --sequence ${SEQ_NUM}

# as Org 2
setGlobals 2
peer lifecycle chaincode approveformyorg ${ORDERER_ARGS} --channelID ${CHANNEL_ID} --name ${ERCC_ID} --signature-policy "${ERCC_EP}" --version ${ERCC_VER} --package-id ${ORG2_ERCC_PKG_ID} --sequence ${SEQ_NUM}
peer lifecycle chaincode approveformyorg ${ORDERER_ARGS} --channelID ${CHANNEL_ID} --name ${CC_ID} --signature-policy "${ECC_EP}" --version ${CC_VER} --package-id ${ORG2_ECC_PKG_ID} --sequence ${SEQ_NUM}


# Commit
#------------
echo ""
echo "Committing both chaincodes"

# as Org 1 (as only org)
setGlobals 1
# DEBUG / TODO clean-me up
peer lifecycle chaincode commit ${ORDERER_ARGS} ${PEER_ARGS} --channelID ${CHANNEL_ID} --name ${ERCC_ID} --signature-policy "${ERCC_EP}" --version ${ERCC_VER} --sequence ${SEQ_NUM}
peer lifecycle chaincode commit ${ORDERER_ARGS} ${PEER_ARGS} --channelID ${CHANNEL_ID} --name ${CC_ID} --signature-policy "${ECC_EP}" --version ${CC_VER} --sequence ${SEQ_NUM}


# Show committed chaincodes
# on Org 1
setGlobals 1
echo ""
echo "Check committed chaincodes on peer from Org1"
peer lifecycle chaincode querycommitted --output json ${ORDERER_ARGS} --channelID ${CHANNEL_ID} --name ${ERCC_ID}
peer lifecycle chaincode querycommitted --output json ${ORDERER_ARGS} --channelID ${CHANNEL_ID} --name ${CC_ID}

# on Org 2
setGlobals 2
echo ""
echo "Check committed chaincodes on peer from Org2"
peer lifecycle chaincode querycommitted --output json ${ORDERER_ARGS} --channelID ${CHANNEL_ID} --name ${ERCC_ID}
peer lifecycle chaincode querycommitted --output json ${ORDERER_ARGS} --channelID ${CHANNEL_ID} --name ${CC_ID}


# Note: above does _not_ yet complete the FPC chaincode initialization.
# The missing 'peer lifecycle chaincode initEnclave' command for CaaS
# must be executed after docker-compose is started. For the test-scenario
# this is done '$FPC_PATH/client_sdk/go/test/main.go'

cat <<EOF
# define (i.e., copy/paste into shell) following environment-variables
# for docker-compose and then start it by calling 'make ercc-ecc-start'
export \\
  ORG1_ECC_PKG_ID=${ORG1_ECC_PKG_ID}\\
  ORG1_ERCC_PKG_ID=${ORG1_ERCC_PKG_ID}\\
  ORG2_ECC_PKG_ID=${ORG2_ECC_PKG_ID}\\
  ORG2_ERCC_PKG_ID=${ORG2_ERCC_PKG_ID}
EOF
