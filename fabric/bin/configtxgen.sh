#!/bin/bash
# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

#RUN=echo # uncomment to dry-run peer call

SCRIPTDIR="$(dirname $(readlink --canonicalize ${BASH_SOURCE}))"
FPC_PATH="${SCRIPTDIR}/../../"
FABRIC_SCRIPTDIR="${FPC_PATH}/fabric/bin/"

: ${FABRIC_CFG_PATH:=$(pwd)}

. ${FABRIC_SCRIPTDIR}/lib/common_utils.sh
. ${FABRIC_SCRIPTDIR}/lib/common_ledger.sh

ARGS_EXEC=( "$@" ) # params to eventually pass to real orderer /default: just pass all original args ..


# Call real orderer
try $RUN exec ${FABRIC_BIN_DIR}/configtxgen "${ARGS_EXEC[@]}"
