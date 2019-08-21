#!/bin/bash

# Copyright Intel Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

#set -x
SCRIPTDIR="$(dirname $(readlink --canonicalize ${BASH_SOURCE}))"
FPC_TOP_DIR="${SCRIPTDIR}/.."
FABRIC_SCRIPTDIR="${FPC_TOP_DIR}/fabric/bin/"

: ${FABRIC_CFG_PATH:="${SCRIPTDIR}/config"}

. ${FABRIC_SCRIPTDIR}/lib/common_utils.sh
. ${FABRIC_SCRIPTDIR}/lib/common_ledger.sh

CC_ID=auction_test
RESULT=PASSED
failures=0

#this is the path that will be used for the docker build of the chaincode enclave
ENCLAVE_SO_PATH=examples/auction/_build/lib/

CC_VERS=0
num_rounds=3
num_clients=10

# Try function which returns the response string
try_r() {
    echo "$@" 
    export RESPONSE=""
    "$@" 2>&1 | tee -a /tmp/response.txt || die "test failed: $*"
    export RESPONSE=$(cat /tmp/response.txt | awk -F "\"" '{print $5}' | awk -F "\\" '{print $1}' | base64 -d)
    say $RESPONSE
    rm /tmp/response.txt
}

# Check the Response returned to validate expected Response
check_result() {
    if [[ "$1" == "$2" ]]; then
        export RESULT=PASSED
	gecho $RESULT
    else
        export RESULT=FAILED
	recho $RESULT
        export failures=$((failures+1))
    fi
}

auction_test() {

    # install, init, and register (auction) chaincode
    try ${PEER_CMD} chaincode install -l fpc-c -n ${CC_ID} -v ${CC_VERS} -p ${ENCLAVE_SO_PATH}
    sleep 3

    try ${PEER_CMD} chaincode instantiate -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -v ${CC_VERS} -c '{"args":["My Auction"]}' -V ecc-vscc
    sleep 3

    # create auction
    try ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args": ["[\"create\",\"MyAuction\"]", ""]}' --waitForEvent

    say "invoke submit"
    try ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["[\"submit\",\"MyAuction\", \"JohnnyCash0\", \"0\"]", ""]}' --waitForEvent

    # Scenario 1
    becho ">>>> Close and evaluate non existing auction. Response should be AUCTION_NOT_EXISTING"
    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["[\"close\",\"MyAuction\"]",""]}' --waitForEvent
    check_result $RESPONSE "AUCTION_NOT_EXISTING"
    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["[\"eval\",\"MyAuction\"]", ""]}'  # Don't do --waitForEvent, so potentially there is some parallelism here ..
    check_result $RESPONSE "AUCTION_NOT_EXISTING"

    # Scenario 2
    becho ">>>> Create an auction. Response should be OK"
    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args": ["[\"create\",\"MyAuction1\"]", ""]}' --waitForEvent
    check_result $RESPONSE "OK" 
    becho ">>>> Create two equivalent bids. Response should be OK"
    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["[\"submit\",\"MyAuction1\", \"JohnnyCash0\", \"2\"]", ""]}' # Don't do --waitForEvent, so potentially there is some parallelism here ..
    check_result $RESPONSE "OK" 
    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["[\"submit\",\"MyAuction1\", \"JohnnyCash1\", \"2\"]", ""]}' # Don't do --waitForEvent, so potentially there is some parallelism here ..
    check_result $RESPONSE "OK" 
    becho ">>>> Close auction. Response should be OK"
    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["[\"close\",\"MyAuction1\"]",""]}' --waitForEvent
    check_result $RESPONSE "OK" 
    becho ">>>> Submit a bid on a closed auction. Response should be AUCTION_ALREADY_CLOSED"
    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["[\"close\",\"MyAuction1\"]",""]}' --waitForEvent
    check_result $RESPONSE "AUCTION_ALREADY_CLOSED"; 
    becho ">>>> Evaluate auction. Response should be DRAW"
    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["[\"eval\",\"MyAuction1\"]", ""]}'  # Don't do --waitForEvent, so potentially there is some parallelism here ..
    check_result $RESPONSE "DRAW" 

    # Scenario 3
    becho ">>>> Create an auction. Response should be OK"
    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args": ["[\"create\",\"MyAuction2\"]", ""]}' --waitForEvent
    check_result $RESPONSE "OK" 
    for (( i=0; i<=$num_rounds; i++ ))
    do
        becho ">>>> Submit unique bid. Response should be OK"
        b="$(($i%$num_clients))"
        try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["[\"submit\",\"MyAuction2\", \"JohnnyCash'$b'\", \"'$b'\"]", ""]}' # Don't do --waitForEvent, so potentially there is some parallelism here ..
        check_result $RESPONSE "OK" 
    done
    becho ">>>> Close auction. Response should be OK"
    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["[\"close\",\"MyAuction2\"]",""]}' --waitForEvent
    check_result $RESPONSE "OK"
    becho ">>>> Evaluate auction. Auction Result should be printed out"
    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["[\"eval\",\"MyAuction2\"]", ""]}'  # Don't do --waitForEvent, so potentially there is some parallelism here ..

    # Scenario 4
    becho ">>>> Create a new auction. Response should be OK"
    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args": ["[\"create\",\"MyAuction3\"]", ""]}' --waitForEvent
    check_result $RESPONSE "OK"
    becho  ">>>> Create a duplicate auction. Response should be AUCTION_ALREADY_EXISTING"
    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args": ["[\"create\",\"MyAuction3\"]", ""]}' --waitForEvent
    check_result $RESPONSE "AUCTION_ALREADY_EXISTING"
    becho ">>>> Close auction and evaluate. Response should be OK"
    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["[\"close\",\"MyAuction3\"]",""]}' --waitForEvent
    check_result $RESPONSE "OK"
    becho ">>>> Close an already closed auction. Response should be AUCTION_ALREADY_CLOSED"
    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["[\"close\",\"MyAuction3\"]",""]}' --waitForEvent
    check_result $RESPONSE "AUCTION_ALREADY_CLOSED"
    becho ">>>> Evaluate auction. Response should be NO_BIDS"
    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["[\"eval\",\"MyAuction3\"]", ""]}'  # Don't do --waitForEvent, so potentially there is some parallelism here ..
    check_result $RESPONSE "NO_BIDS"
}

# 1. prepare
para
say "Preparing Auction Test ..."
# - clean up relevant docker images
docker_clean ${ERCC_ID}
docker_clean ${CC_ID}

trap ledger_shutdown EXIT

para
say "Run auction test"

say "- setup ledger"
ledger_init

say "- auction test"
auction_test

say "- shutdown ledger"
ledger_shutdown

para
if [[ "$failures" == 0 ]]; then
    yell "Auction test PASSED"
else
    yell "Auction test had ${failures} failures"
fi
exit 0

