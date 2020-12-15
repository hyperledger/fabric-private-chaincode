#!/bin/bash

# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

SCRIPTDIR="$(dirname $(readlink --canonicalize ${BASH_SOURCE}))"
FPC_PATH="${SCRIPTDIR}/../.."
FABRIC_SCRIPTDIR="${FPC_PATH}/fabric/bin/"

: ${FABRIC_CFG_PATH:=$(pwd)}

. ${FABRIC_SCRIPTDIR}/lib/common_utils.sh
. ${FABRIC_SCRIPTDIR}/lib/common_ledger.sh

say "- initialize ledger"
ledger_init

say "- ledger with channel '${CHAN_ID}' started based on config in '${FABRIC_CFG_PATH}'"
say "  (peer-id: '${PEER_ID}' / net-id: '${NET_ID}' / api-key-file: '${API_KEY_FILE}' / spid-file: '${SPID_FILE}' / state: '${FABRIC_STATE_DIR}')"
say "  For convenience, a number of useful env-variables when interacting with this ledger:"
cat <<EOT
export \\
   FABRIC_CFG_PATH="${FABRIC_CFG_PATH}" \\
   FABRIC_STATE_DIR="${FABRIC_STATE_DIR}" \\
   PEER_CMD="${PEER_CMD}" \\
   ORDERER_CMD="${ORDERER_CMD}" \\
   CONFIGTXGEN_CMD="${CONFIGTXGEN_CMD}" \\
   ORDERER_ADDR="${ORDERER_ADDR}" \\
   CHAN_ID="${CHAN_ID}" \\
   ERCC_ID="${ERCC_ID}" \\
   ERCC_VERSION="${ERCC_VERSION}" \\
   ORDERER_PID_FILE="${ORDERER_PID_FILE}" \\
   ORDERER_LOG_OUT="${ORDERER_LOG_OUT}" \\
   ORDERER_LOG_ERR="${ORDERER_LOG_ERR}" \\
   PEER_PID_FILE="${PEER_PID_FILE}" \\
   PEER_LOG_OUT="${PEER_LOG_OUT}" \\
   PEER_LOG_ERR="${PEER_LOG_ERR}" \\
EOT

