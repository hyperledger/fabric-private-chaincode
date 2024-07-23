#!/usr/bin/env bash

# Copyright IBM Corp. All Rights Reserved.
# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

CHANNEL_NAME=${CHANNEL_NAME:-mychannel}
FABRIC_SAMPLES=${FPC_PATH}/samples/deployment/test-network/fabric-samples
NETWORK_CMD=${FABRIC_SAMPLES}/test-network/network.sh

if [ ! -f "${NETWORK_CMD}" ]; then
  echo "Error: ${NETWORK_CMD} does not exist"
  exit 1
fi

# define endorsement policies
ERCC_EP="OutOf(2,'Org1MSP.peer','Org2MSP.peer')"
ECC_EP="OutOf(2,'Org1MSP.peer','Org2MSP.peer')"

# define chaincode details

ERCC_ID="ercc"
ERCC_VER="1.0"

# this is important
CC_ID=${CC_ID}
CC_VER=${CC_VER}

CC_PATH="." # the actual path is not needed here since we build the container images manually

# install ercc and ecc
${NETWORK_CMD} deployCCAAS -c "$CHANNEL_NAME" -ccn "$ERCC_ID" -ccp "$CC_PATH" -ccv "$ERCC_VER" -ccep "$ERCC_EP" -ccaasdocker false
${NETWORK_CMD} deployCCAAS -c "$CHANNEL_NAME" -ccn "$CC_ID" -ccp "$CC_PATH" -ccv "$CC_VER" -ccep "$ECC_EP" -ccaasdocker false

# export chaincode package ids
ORG1_ERCC_PKG_ID=$(${NETWORK_CMD} cc list -org 1 | grep "Label: ${ERCC_ID}_${ERCC_VER}" | awk '{print $3}' | sed 's/.$//')
ORG2_ERCC_PKG_ID=$(${NETWORK_CMD} cc list -org 2 | grep "Label: ${ERCC_ID}_${ERCC_VER}" | awk '{print $3}' | sed 's/.$//')
ORG1_ECC_PKG_ID=$(${NETWORK_CMD} cc list -org 1 | grep "Label: ${CC_ID}_${CC_VER}" | awk '{print $3}' | sed 's/.$//')
ORG2_ECC_PKG_ID=$(${NETWORK_CMD} cc list -org 2 | grep "Label: ${CC_ID}_${CC_VER}" | awk '{print $3}' | sed 's/.$//')

cat > .env << EOF
ORG1_ECC_PKG_ID="${ORG1_ECC_PKG_ID}"
ORG1_ERCC_PKG_ID="${ORG1_ERCC_PKG_ID}"
ORG2_ECC_PKG_ID="${ORG2_ECC_PKG_ID}"
ORG2_ERCC_PKG_ID="${ORG2_ERCC_PKG_ID}"
EOF
