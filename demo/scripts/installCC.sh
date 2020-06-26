#!/bin/bash

#
# Copyright IBM Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#

set -ev

export WAIT_TIME=15
export version=1.0
export FPC_PATH=/project/src/github.com/hyperledger-labs/fabric-private-chaincode
export FABRIC_BIN_DIR=${FPC_PATH}/../../hyperledger/fabric/build/bin
export PEER_CMD=${FPC_PATH}/fabric/bin/peer.sh

# Command to execute peer cli inside peer container 
# Note: by default the peer runs with peer credentials, so we have to override CORE_PEER_MSPCONFIGPATH env-var that the peer in cli-mode doesn't use peer but admin credentials
REMOTE_PEER_CMD="docker exec -e CORE_PEER_LOCALMSPID=Org1MSP -e CORE_PEER_MSPCONFIGPATH=/etc/hyperledger/msp/users/Admin@org1.example.com/msp peer0.org1.example.com env TERM=${TERM} ${PEER_CMD}"
REMOTE_CAT="docker exec peer0.org1.example.com cat"

# Commands to execute cli _outside_ peer container
# TODO: fix this WIP ....
# REMOTE_PEER_CMD="${PEER_CMD}"
# REMOTE_CAT="cat"
# export FABRIC_CFG_PATH=${FPC_PATH}/utils/docker-compose/network-config-outside-docker
# export CORE_PEER_LOCALMSPID=Org1MSP
# export CORE_PEER_MSPCONFIGPATH=${FPC_PATH}/utils/docker-compose/network-config/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp

CHAN_ID=mychannel

CC_ID=auctioncc
CC_PATH=${FPC_PATH}/demo/chaincode/fpc/_build/lib
CC_LANG=fpc-c
CC_VER="$(${REMOTE_CAT} ${CC_PATH}/mrenclave)"
PKG=/tmp/${CC_ID}.tar.gz

CC_EP="OR('Org1MSP.peer')"

${REMOTE_PEER_CMD} lifecycle chaincode package --lang ${CC_LANG} --label ${CC_ID} --path ${CC_PATH} ${PKG}
${REMOTE_PEER_CMD} lifecycle chaincode install ${PKG}
PKG_ID=$(${REMOTE_PEER_CMD} lifecycle chaincode queryinstalled | awk "/Package ID: ${CC_ID}/{print}" | sed -n 's/^Package ID: //; s/, Label:.*$//;p')
${REMOTE_PEER_CMD} lifecycle chaincode approveformyorg -C ${CHAN_ID} --package-id ${PKG_ID} --name ${CC_ID} --version ${CC_VER} --signature-policy ${CC_EP}
${REMOTE_PEER_CMD} lifecycle chaincode checkcommitreadiness -C ${CHAN_ID} --name ${CC_ID} --version ${CC_VER} --signature-policy ${CC_EP}
${REMOTE_PEER_CMD} lifecycle chaincode commit -C ${CHAN_ID} --name ${CC_ID} --version ${CC_VER}  --signature-policy ${CC_EP}
${REMOTE_PEER_CMD} lifecycle chaincode querycommitted -C ${CHAN_ID}
${REMOTE_PEER_CMD} lifecycle chaincode createenclave --name ${CC_ID}
