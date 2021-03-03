#!/bin/bash

# Copyright IBM Corp. All Rights Reserved.
# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

# Package a fpc chaincode for CaaS mode
# (for normal external-builder, use '$FPC_PATH/fabric/bin/peer.sh lifecycle chaincode package')

set -euo pipefail

#DEBUG=true # uncomment (or define when calling script) to show debug output

if [[ -z "${FPC_PATH}" ]]; then
  echo "Error: FPC_PATH not set"
  exit 1
fi

. "${FPC_PATH}"/utils/packaging/utils.sh

if [ "$#" -ne 6 ]; then
  echo "ERROR: incorrect number of parameters" >&2
  echo "Use: ./package.sh <deployment-path> <ercc-id> <ercc-version> <cc-id> <cc-version> <peer-id>" >&2

  exit 1
fi


DEPLOYMENT_PATH="$1"
ERCC_ID="$2"
ERCC_VER="$3"
CC_ID="$4"
CC_VER="$5"
PEER="$6"

TYPE="external"
CHAINCODE_SERVER_PORT=9999

endpoint="${ERCC_ID}.${PEER}:${CHAINCODE_SERVER_PORT}"
packageName="${ERCC_ID}.${PEER}.tgz"
packageChaincode "${DEPLOYMENT_PATH}" "${packageName}" "${ERCC_ID}" "${ERCC_VER}" "${TYPE}" "${endpoint}" "${PEER}"

endpoint="${CC_ID}.${PEER}:${CHAINCODE_SERVER_PORT}"
packageName="${CC_ID}.${PEER}.tgz"
packageChaincode "${DEPLOYMENT_PATH}" "${packageName}" "${CC_ID}" "${CC_VER}" "${TYPE}" "${endpoint}" "${PEER}"
