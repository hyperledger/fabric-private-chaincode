#!/bin/bash

# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

SCRIPTDIR="$(dirname $(readlink --canonicalize ${BASH_SOURCE}))"
FPC_PATH="${SCRIPTDIR}/.."
FABRIC_SCRIPTDIR="${FPC_PATH}/fabric/bin/"

: ${FABRIC_CFG_PATH:="${SCRIPTDIR}/config"}

. ${FABRIC_SCRIPTDIR}/lib/common_utils.sh
. ${FABRIC_SCRIPTDIR}/lib/common_ledger.sh

CC_ID=auction_test
CC_PATH=${FPC_PATH}/samples/chaincode/auction/_build/lib/
CC_LANG=fpc-c
CC_VER="$(cat ${CC_PATH}/mrenclave)"
CC_SEQ="1"
CC_EP="OR('SampleOrg.member')" # note that we use .member as NodeOUs is disabled with the crypto material used in the integration tests.

num_rounds=3
num_clients=10
NUM_FAILURES=0
NUM_TESTS=0

auction_test() {
    PKG=/tmp/${CC_ID}.tar.gz

    try ${PEER_CMD} lifecycle chaincode package --lang ${CC_LANG} --label ${CC_ID} --path ${CC_PATH} ${PKG}
    try ${PEER_CMD} lifecycle chaincode install ${PKG}

    PKG_ID=$(${PEER_CMD} lifecycle chaincode queryinstalled | awk "/Package ID: ${CC_ID}/{print}" | sed -n 's/^Package ID: //; s/, Label:.*$//;p')

    # first call negated as it fails due to specification of validation plugin
    try_fail ${PEER_CMD} lifecycle chaincode approveformyorg -o ${ORDERER_ADDR} -C ${CHAN_ID} --package-id ${PKG_ID} --name ${CC_ID} --version ${CC_VER} --sequence ${CC_SEQ} --signature-policy ${CC_EP} -E mock-escc -V fpc-vscc
    try ${PEER_CMD} lifecycle chaincode approveformyorg -o ${ORDERER_ADDR} -C ${CHAN_ID} --package-id ${PKG_ID} --name ${CC_ID} --version ${CC_VER} --sequence ${CC_SEQ} --signature-policy ${CC_EP}

    try ${PEER_CMD} lifecycle chaincode checkcommitreadiness -C ${CHAN_ID} --name ${CC_ID} --version ${CC_VER} --sequence ${CC_SEQ} --signature-policy ${CC_EP}

    try ${PEER_CMD} lifecycle chaincode commit -o ${ORDERER_ADDR} -C ${CHAN_ID} --name ${CC_ID} --version ${CC_VER} --sequence ${CC_SEQ} --signature-policy ${CC_EP}

    try ${PEER_CMD} lifecycle chaincode initEnclave -o ${ORDERER_ADDR} --peerAddresses "localhost:7051" --name ${CC_ID}

    try ${PEER_CMD} lifecycle chaincode querycommitted -C ${CHAN_ID}

    # Scenario 1
    becho ">>>> Close and evaluate non existing auction. Response should be AUCTION_NOT_EXISTING"
    try_out_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"init", "Args": ["MyAuctionHouse"]}' --waitForEvent
    check_result "OK"
    try_out_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"close", "Args": ["MyAuction"]}' --waitForEvent
    check_result "AUCTION_NOT_EXISTING"
    try_out_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"eval", "Args": ["MyAuction0"]}' # Don't do --waitForEvent, so potentially there is some parallelism here ..
    check_result "AUCTION_NOT_EXISTING"

    # Scenario 2
    becho ">>>> Create an auction. Response should be OK"
    try_out_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"init", "Args": ["MyAuctionHouse"]}' --waitForEvent
    check_result "OK"
    try_out_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"create", "Args": ["MyAuction1"]}' --waitForEvent
    check_result "OK"
    becho ">>>> Create two equivalent bids. Response should be OK"
    try_out_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"submit", "Args": ["MyAuction1", "JohnnyCash0", "2"]}' # Don't do --waitForEvent, so potentially there is some parallelism here ..
    check_result "OK"
    try_out_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"submit", "Args": ["MyAuction1", "JohnnyCash1", "2"]}' # Don't do --waitForEvent, so potentially there is some parallelism here ..
    check_result "OK"
    becho ">>>> Close auction. Response should be OK"
    try_out_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"close", "Args": ["MyAuction1"]}' --waitForEvent
    check_result "OK"
    becho ">>>> Submit a bid on a closed auction. Response should be AUCTION_ALREADY_CLOSED"
    try_out_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"submit", "Args": ["MyAuction1", "JohnnyCash2", "2"]}' # Don't do --waitForEvent, so potentially there is some parallelism here ..
    check_result "AUCTION_ALREADY_CLOSED";
    becho ">>>> Evaluate auction. Response should be DRAW"
    try_out_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"eval", "Args": ["MyAuction1"]}' # Don't do --waitForEvent, so potentially there is some parallelism here ..
    check_result "DRAW"

    # Scenario 3
    becho ">>>> Create an auction. Response should be OK"
    try_out_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"init", "Args": ["MyAuctionHouse"]}' --waitForEvent
    check_result "OK"
    try_out_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"create", "Args": ["MyAuction2"]}' --waitForEvent
    check_result "OK"
    for (( i=0; i<=$num_rounds; i++ ))
    do
        becho ">>>> Submit unique bid. Response should be OK"
        b="$(($i%$num_clients))"
        try_out_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"submit", "Args": ["MyAuction2", "JohnnyCash'$b'", "'$b'"]}' # Don't do --waitForEvent, so potentially there is some parallelism here ..
        check_result "OK"
    done
    becho ">>>> Close auction. Response should be OK"
    try_out_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"close", "Args": ["MyAuction2"]}' --waitForEvent
    check_result "OK"
    becho ">>>> Evaluate auction. Auction Result should be printed out"
    try_out_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"eval", "Args": ["MyAuction2"]}' # Don't do --waitForEvent, so potentially there is some parallelism here ..
    check_result '{"bidder":"JohnnyCash3","value":3}'

    # Scenario 4
    becho ">>>> Create a new auction. Response should be OK"
    try_out_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"init", "Args": ["MyAuctionHouse"]}' --waitForEvent
    check_result "OK"
    try_out_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"create", "Args": ["MyAuction3"]}' --waitForEvent
    check_result "OK"
    becho  ">>>> Create a duplicate auction. Response should be AUCTION_ALREADY_EXISTING"
    try_out_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"create", "Args": ["MyAuction3"]}' --waitForEvent
    check_result "AUCTION_ALREADY_EXISTING"
    becho ">>>> Close auction and evaluate. Response should be OK"
    try_out_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"close", "Args": ["MyAuction3"]}' --waitForEvent
    check_result "OK"
    becho ">>>> Close an already closed auction. Response should be AUCTION_ALREADY_CLOSED"
    try_out_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"close", "Args": ["MyAuction3"]}' --waitForEvent
    check_result "AUCTION_ALREADY_CLOSED"
    becho ">>>> Evaluate auction. Response should be NO_BIDS"
    try_out_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"eval", "Args": ["MyAuction3"]}' # Don't do --waitForEvent, so potentially there is some parallelism here ..
    check_result "NO_BIDS"

    # Code below is used to test bug in issue #42
    becho ">>>> Create a new auction. Response should be OK"
    try_out_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"init", "Args": ["MyAuctionHouse"]}' --waitForEvent
    check_result "OK"
    try_out_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"create", "Args": ["MyAuction4"]}' --waitForEvent
    check_result "OK"
}

# 1. prepare
para
say "Preparing Auction Test ..."
# - clean up relevant docker images
docker_clean ${ERCC_ID}

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
if [[ "$NUM_FAILURES" == 0 ]]; then
    yell "Auction test PASSED"
else
    yell "Auction test had ${NUM_FAILURES} failures out of ${NUM_TESTS} tests"
    exit 1
fi
exit 0
