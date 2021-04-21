#!/bin/bash

# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

# for the post-test script callout we want all variables expoerted as env-variables ...
set -a

SCRIPTDIR="$(dirname $(readlink --canonicalize ${BASH_SOURCE}))"
FPC_PATH="${SCRIPTDIR}/.."
FABRIC_SCRIPTDIR="${FPC_PATH}/fabric/bin/"

: ${FABRIC_CFG_PATH:="${SCRIPTDIR}/config"}

. ${FABRIC_SCRIPTDIR}/lib/common_utils.sh
. ${FABRIC_SCRIPTDIR}/lib/common_ledger.sh

CC_EP="OR('SampleOrg.member')" # note that we use .member as NodeOUs is disabled with the crypto material used in the integration tests.
NUM_FAILURES=0
NUM_TESTS=0

run_test() {

    # install samples/chaincode/auction
    CC_PATH=${FPC_PATH}/samples/chaincode/auction/_build/lib/
    CC_VER="$(cat ${CC_PATH}/mrenclave)"
    CC_SEQ="1"
    PKG=/tmp/auction_test.tar.gz

    try ${PEER_CMD} lifecycle chaincode package --lang fpc-c --label auction_test --path ${CC_PATH} ${PKG}
    try ${PEER_CMD} lifecycle chaincode install ${PKG}
    PKG_ID=$(${PEER_CMD} lifecycle chaincode queryinstalled | awk "/Package ID: auction_test/{print}" | sed -n 's/^Package ID: //; s/, Label:.*$//;p')

    try ${PEER_CMD} lifecycle chaincode approveformyorg -o ${ORDERER_ADDR} -C ${CHAN_ID} --package-id ${PKG_ID} --name auction_test --version ${CC_VER} --sequence ${CC_SEQ} --signature-policy ${CC_EP}

    try ${PEER_CMD} lifecycle chaincode checkcommitreadiness -C ${CHAN_ID} --name auction_test --version ${CC_VER} --sequence ${CC_SEQ} --signature-policy ${CC_EP}

    try ${PEER_CMD} lifecycle chaincode commit -o ${ORDERER_ADDR} -C ${CHAN_ID} --name auction_test --version ${CC_VER} --sequence ${CC_SEQ} --signature-policy ${CC_EP}

    try ${PEER_CMD} lifecycle chaincode initEnclave -o ${ORDERER_ADDR} --peerAddresses "localhost:7051" --name auction_test

    # install something non-fpc
    CC_PATH="github.com/hyperledger/fabric-samples/chaincode/marbles02/go"
    CC_VER="0"
    CC_SEQ="1"
    PKG=/tmp/marbles02.tar.gz
    try ${PEER_CMD} lifecycle chaincode package --lang golang --label marbles02 --path ${CC_PATH} ${PKG}
    try ${PEER_CMD} lifecycle chaincode install ${PKG}
    PKG_ID=$(${PEER_CMD} lifecycle chaincode queryinstalled | awk "/Package ID: marbles02/{print}" | sed -n 's/^Package ID: //; s/, Label:.*$//;p')
    try ${PEER_CMD} lifecycle chaincode approveformyorg -o ${ORDERER_ADDR} -C ${CHAN_ID} --package-id ${PKG_ID} --name marbles02 --version ${CC_VER} --sequence ${CC_SEQ}
    try ${PEER_CMD} lifecycle chaincode checkcommitreadiness -C ${CHAN_ID} --name marbles02 --version ${CC_VER} --sequence ${CC_SEQ}
    try ${PEER_CMD} lifecycle chaincode commit -o ${ORDERER_ADDR} -C ${CHAN_ID} --name marbles02 --version ${CC_VER} --sequence ${CC_SEQ}

    try_out_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n auction_test -c '{"Args":["init", "MyAuctionHouse"]}' --waitForEvent
    check_result "OK"

    try_out_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n auction_test -c '{"Args":["create", "MyAuction"]}' --waitForEvent
    check_result "OK"

    try_out_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n marbles02 -c '{"Args":["initMarble","marble1","blue","35","tom"]}' --waitForEvent

    # install samples/chaincode/echo
    CC_PATH=${FPC_PATH}/samples/chaincode/echo/_build/lib/
    CC_VER="$(cat ${CC_PATH}/mrenclave)"
    CC_SEQ="1"
    PKG=/tmp/echo_test.tar.gz

    try ${PEER_CMD} lifecycle chaincode package --lang fpc-c --label echo_test --path ${FPC_PATH}/samples/chaincode/echo/_build/lib ${PKG}
    try ${PEER_CMD} lifecycle chaincode install ${PKG}
    PKG_ID=$(${PEER_CMD} lifecycle chaincode queryinstalled | awk "/Package ID: echo_test/{print}" | sed -n 's/^Package ID: //; s/, Label:.*$//;p')

    # first call negated as it fails due to specification of validation plugin
    try_fail ${PEER_CMD} lifecycle chaincode approveformyorg -o ${ORDERER_ADDR} -C ${CHAN_ID} --package-id ${PKG_ID} --name echo_test --version ${CC_VER} --sequence ${CC_SEQ} -E mock-escc -V fpc-vscc
    try ${PEER_CMD} lifecycle chaincode approveformyorg -o ${ORDERER_ADDR} -C ${CHAN_ID} --package-id ${PKG_ID} --name echo_test --version ${CC_VER} --sequence ${CC_SEQ}

    # first call negated as it fails due to specification of validation plugin
    try_fail ${PEER_CMD} lifecycle chaincode checkcommitreadiness -C ${CHAN_ID} --name echo_test --version ${CC_VER} --sequence ${CC_SEQ} -E mock-escc -V fpc-vscc
    try ${PEER_CMD} lifecycle chaincode checkcommitreadiness -C ${CHAN_ID} --name echo_test --version ${CC_VER} --sequence ${CC_SEQ}

    # first call negated as it fails due to specification of validation plugin
    try_fail ${PEER_CMD} lifecycle chaincode commit -o ${ORDERER_ADDR} -C ${CHAN_ID} --name echo_test --version ${CC_VER} --sequence ${CC_SEQ} -E mock-escc -V fpc-vscc
    try ${PEER_CMD} lifecycle chaincode commit -o ${ORDERER_ADDR} -C ${CHAN_ID} --name echo_test --version ${CC_VER} --sequence ${CC_SEQ}

    #first call negated as it fails
    try_fail ${PEER_CMD} lifecycle chaincode initEnclave -o ${ORDERER_ADDR} --peerAddresses "localhost:7051" --name wrong-cc-id
    try ${PEER_CMD} lifecycle chaincode initEnclave -o ${ORDERER_ADDR} --peerAddresses "localhost:7051" --name echo_test

    try_out_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n echo_test -c '{"Args": ["moin"]}' --waitForEvent
    check_result "moin"

    try_out_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n auction_test -c '{"Args":["submit", "MyAuction", "JohnnyCash0", "0"]}' --waitForEvent
    check_result "OK"

    try ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n marbles02 -c '{"Args":["readMarble","marble1"]}' --waitForEvent

    try_out_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n echo_test -c '{"Args": ["bonjour"]}' --waitForEvent
    check_result "bonjour"
}

# 1. prepare
para
say "Preparing Test with mixed concurrent chaincodes, FPC and non-FPC ..."
# - clean up relevant docker images
docker_clean ${ERCC_ID}
docker_clean example02

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
if [[ "$NUM_FAILURES" == 0 ]]; then
    yell "Deployement test PASSED"
else
    yell "Deployement test had ${NUM_FAILURES} failures out of ${NUM_TESTS} tests"
    exit 1
fi
exit 0


