#!/bin/bash

# Copyright Intel Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

# for the post-test script callout we want all variables expoerted as env-variables ...
set -a

SCRIPTDIR="$(dirname $(readlink --canonicalize ${BASH_SOURCE}))"
FPC_TOP_DIR="${SCRIPTDIR}/.."
FABRIC_SCRIPTDIR="${FPC_TOP_DIR}/fabric/bin/"

: ${FABRIC_CFG_PATH:="${SCRIPTDIR}/config"}

. ${FABRIC_SCRIPTDIR}/lib/common_utils.sh
. ${FABRIC_SCRIPTDIR}/lib/common_ledger.sh

CC_VERS=0

run_test() {
    # install, init, and register (auction) chaincode
    try ${PEER_CMD} chaincode install -l fpc-c -n auction_test -v ${CC_VERS} -p examples/auction/_build/lib

    try ${PEER_CMD} chaincode install -l golang -n example02 -v ${CC_VERS} -p github.com/hyperledger/fabric/examples/chaincode/go/example02/cmd

    try ${PEER_CMD} chaincode instantiate -o ${ORDERER_ADDR} -C ${CHAN_ID} -n auction_test -v ${CC_VERS} -c '{"args":["My Auction"]}' -V ecc-vscc
    sleep 3

    try ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n auction_test -c '{"Args": ["[\"create\",\"MyAuction\"]", ""]}' --waitForEvent

    try ${PEER_CMD} chaincode instantiate -o ${ORDERER_ADDR} -C ${CHAN_ID} -n example02 -v ${CC_VERS} -c '{"args":["init", "bob", "100", "alice", "200"]}'
    sleep 3

    try ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n example02 -c '{"args": ["invoke", "bob", "alice", "99"]}' --waitForEvent

    try ${PEER_CMD} chaincode install -l fpc-c -n echo_test -v ${CC_VERS} -p examples/echo/_build/lib
    sleep 3

    try ${PEER_CMD} chaincode instantiate -o ${ORDERER_ADDR} -C ${CHAN_ID} -n echo_test -v ${CC_VERS} -c '{"args":[]}' -V ecc-vscc
    sleep 3

    try ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n echo_test -c '{"Args": ["[\"moin\"]", ""]}' --waitForEvent

    try ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n auction_test -c '{"Args":["[\"submit\",\"MyAuction\", \"JohnnyCash0\", \"0\"]", ""]}' --waitForEvent
}

# 1. prepare
para
say "Preparing Test with mixed concurrent chaincodes, FPC and non-FPC ..."
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

# if we pass an argument it is supposed to be a sourceable script:
# we use that to extract blocks for a scenario for the ledger unit-test
if [ "$1" != "" ]; then
	say "Running post-test script '$@'"
	"$@"
fi

say "- shutdown ledger"
ledger_shutdown

para
yell "Test PASSED"

exit 0
