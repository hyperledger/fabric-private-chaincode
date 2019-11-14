#!/bin/bash
#
# Copyright IBM Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#
export SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
export PATH=${SCRIPT_DIR}/../bin:${PWD}:$PATH
export FABRIC_CFG_PATH=${SCRIPT_DIR}/../network-config
CHANNEL_NAME=mychannel

# remove previous crypto material and config transactions
rm -fr ${FABRIC_CFG_PATH}/config/*
mkdir -p ${FABRIC_CFG_PATH}/config
rm -fr ${FABRIC_CFG_PATH}/crypto-config/*

# generate crypto material
cryptogen generate --config=${FABRIC_CFG_PATH}/crypto-config.yaml --output=${FABRIC_CFG_PATH}/crypto-config/
if [ "$?" -ne 0 ]; then
  echo "Failed to generate crypto material..."
  exit 1
fi

mv ${FABRIC_CFG_PATH}/crypto-config/peerOrganizations/org1.example.com/ca/*_sk ${FABRIC_CFG_PATH}/crypto-config/peerOrganizations/org1.example.com/ca/ca.org1.example.com_sk

#generate channel configuration transaction
configtxgen -profile OneOrgChannel -outputCreateChannelTx ${FABRIC_CFG_PATH}/config/channel.tx -channelID $CHANNEL_NAME
if [ "$?" -ne 0 ]; then
  echo "Failed to generate channel configuration transaction..."
  exit 1
fi

#generate genesis block for orderer
configtxgen -profile OneOrgOrdererGenesis -outputBlock ${FABRIC_CFG_PATH}/config/genesis.block
if [ "$?" -ne 0 ]; then
  echo "Failed to generate orderer genesis block..."
  exit 1
fi
