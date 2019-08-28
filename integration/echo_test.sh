#!/bin/bash

# Copyright Intel Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

SCRIPTDIR="$(dirname $(readlink --canonicalize ${BASH_SOURCE}))"
FPC_TOP_DIR="${SCRIPTDIR}/.."
FABRIC_SCRIPTDIR="${FPC_TOP_DIR}/fabric/bin/"

: ${FABRIC_CFG_PATH:="${SCRIPTDIR}/config"}

. ${FABRIC_SCRIPTDIR}/lib/common_utils.sh
. ${FABRIC_SCRIPTDIR}/lib/common_ledger.sh

CC_ID=echo_test

#this is the path that will be used for the docker build of the chaincode enclave
ENCLAVE_SO_PATH=examples/echo/_build/lib/

CC_VERS=0
num_rounds=10

echo_test() {
    # install, init, and register (auction) chaincode
    try ${PEER_CMD} chaincode install -l fpc-c -n ${CC_ID} -v ${CC_VERS} -p ${ENCLAVE_SO_PATH}
    sleep 3

    try ${PEER_CMD} chaincode instantiate -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -v ${CC_VERS} -c '{"args":[]}' -V ecc-vscc
    sleep 3

    say "do echos"
    for (( i=1; i<=$num_rounds; i++ ))
    do
        # echos
        try ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args": ["[\"echo-'$i'\"]", ""]}' --waitForEvent
    done
}

# 1. prepare
para
say "Preparing Echo Test ..."
# - clean up relevant docker images
docker_clean ${ERCC_ID}
docker_clean ${CC_ID}

trap ledger_shutdown EXIT


para
say "Run echo test"

say "- setup ledger"
ledger_init

say "- echo test"
echo_test

say "- shutdown ledger"
ledger_shutdown

para
yell "Echo test PASSED"

exit 0
