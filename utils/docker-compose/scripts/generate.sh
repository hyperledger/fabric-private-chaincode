#!/bin/bash
#
# Copyright IBM Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#
export SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. ${SCRIPT_DIR}/lib/common.sh

# test pre-conditions and try to remediate
#
# - existances of docker binaries
#   as docker(-compose) will download images, we care only about binaries
#   but download only if we do not have already executables installed and/or
#   in path. We also assume that if you have cryptogen in path, you probably
#   have the others needed, e.g., configtxgen, as well ...
FABRIC_BINARY_PATH="${FPC_PATH}/utils/docker-compose/bin"
if [ ! -d "${FABRIC_BINARY_PATH}" ] && [ -z "$(which cryptogen)" ]; then
    echo "Fabirc binaries not found in '${FABRIC_BINARY_PATH}' or in \$PATH, try to download them ..."
    "${FPC_PATH}/utils/docker-compose/scripts/bootstrap.sh" -d
fi

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
mv ${FABRIC_CFG_PATH}/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/keystore/*_sk ${FABRIC_CFG_PATH}/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/keystore/admin_sk

#generate channel configuration transaction
configtxgen -profile OneOrgChannel -outputCreateChannelTx ${FABRIC_CFG_PATH}/config/channel.tx -channelID ${CHANNEL_NAME}
if [ "$?" -ne 0 ]; then
  echo "Failed to generate channel configuration transaction..."
  exit 1
fi

#generate genesis block for orderer
configtxgen -profile SampleSingleNodeEtcdRaft -outputBlock ${FABRIC_CFG_PATH}/config/genesis.block -channelID orderer-system-channel
if [ "$?" -ne 0 ]; then
  echo "Failed to generate orderer genesis block..."
  exit 1
fi
