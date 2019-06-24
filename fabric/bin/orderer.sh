#!/bin/bash
# Copyright Intel Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

#RUN=echo # uncomment to dry-run peer call

SCRIPTDIR="$(dirname $(readlink --canonicalize ${BASH_SOURCE}))"
FPC_TOP_DIR="${SCRIPTDIR}/../../"
CONFIG_HOME="$(pwd)"

. ${SCRIPTDIR}/lib/common_ledger.sh

ARGS_EXEC=( "$@" ) # params to eventually pass to real orderer /default: just pass all original args ..


# Call real orderer
try $RUN exec ${FABRIC_BIN_DIR}/orderer "${ARGS_EXEC[@]}"
