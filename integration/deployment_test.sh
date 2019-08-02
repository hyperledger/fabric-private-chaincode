#!/bin/bash

# Copyright Intel Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

SCRIPTDIR="$(dirname $(readlink --canonicalize ${BASH_SOURCE}))"
FPC_TOP_DIR="${SCRIPTDIR}/.."
CONFIG_HOME="${SCRIPTDIR}/config"
FABRIC_SCRIPTDIR="${FPC_TOP_DIR}/fabric/bin/"

. ${FABRIC_SCRIPTDIR}/lib/common_utils.sh
. ${FABRIC_SCRIPTDIR}/lib/common_ledger.sh

CC_VERS=0
num_rounds=3
num_clients=10

run_test() {
    # install, init, and register (auction) chaincode
    try ${PEER_CMD} chaincode install -l fpc-c -n auction_test -v ${CC_VERS} -p examples/auction/_build/lib
    sleep 3

    try ${PEER_CMD} chaincode install -l fpc-c -n echo_test -v ${CC_VERS} -p examples/echo/_build/lib
    sleep 3

    try ${PEER_CMD} chaincode install -l golang -n example02 -v ${CC_VERS} -p github.com/hyperledger/fabric/examples/chaincode/go/example02/cmd
    sleep 3

    try ${PEER_CMD} chaincode instantiate -o ${ORDERER_ADDR} -C ${CHAN_ID} -n auction_test -v ${CC_VERS} -c '{"args":["init"]}' -V ecc-vscc
    sleep 3

    try ${PEER_CMD} chaincode instantiate -o ${ORDERER_ADDR} -C ${CHAN_ID} -n echo_test -v ${CC_VERS} -c '{"args":["init"]}' -V ecc-vscc
    sleep 3

    try ${PEER_CMD} chaincode instantiate -o ${ORDERER_ADDR} -C ${CHAN_ID} -n example02 -v ${CC_VERS} -c '{"args":["init", "bob", "100", "alice", "200"]}'
    sleep 3

    try ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n example02 -c '{"args": ["invoke", "bob", "alice", "99"]}' --waitForEvent

    try ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n auction_test -c '{"Args":["setup", "ercc"]}' --waitForEvent
    try ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n echo_test -c '{"Args":["setup", "ercc"]}' --waitForEvent

    # create auction
    try ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n auction_test -c '{"Args": ["[\"create\",\"MyAuction\"]", ""]}' --waitForEvent
    try ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n echo_test -c '{"Args": ["[\"moin\"]", ""]}' --waitForEvent

}

# 1. prepare
para
say "Preparing two Test ..."
# - clean up relevant docker images
docker_clean ${ERCC_ID}
docker_clean auction_test
docker_clean echo_test

trap ledger_shutdown EXIT


para
say "Run test"

say "- setup ledger"
ledger_init

say "- this test"
run_test

say "- shutdown ledger"
ledger_shutdown

para
yell "Test PASSED"

exit 0
