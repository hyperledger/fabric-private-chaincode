#!/bin/bash

# Copyright 2020 Intel Corporation
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

CC_EP="OR('SampleOrg.member')" # note that we use .member as NodeOUs is disabled with the crypto material used in the integration tests.
FAILURES=0

run_test() {

    # install examples/auction
    CC_PATH=${FPC_TOP_DIR}/examples/auction/_build/lib/
    CC_VER="$(cat ${CC_PATH}/mrenclave)"
    PKG=/tmp/auction_test.tar.gz

    try ${PEER_CMD} lifecycle chaincode package --lang fpc-c --label auction_test --path ${CC_PATH} ${PKG}
    try ${PEER_CMD} lifecycle chaincode install ${PKG}
    PKG_ID=$(${PEER_CMD} lifecycle chaincode queryinstalled | awk "/Package ID: auction_test/{print}" | sed -n 's/^Package ID: //; s/, Label:.*$//;p')

    # first call negated as it fails due to specification of validation plugin
    try_fail ${PEER_CMD} lifecycle chaincode approveformyorg -C ${CHAN_ID} --package-id ${PKG_ID} --name auction_test --version ${CC_VER} --signature-policy ${CC_EP} -E mock-escc -V fpc-vscc
    try ${PEER_CMD} lifecycle chaincode approveformyorg -C ${CHAN_ID} --package-id ${PKG_ID} --name auction_test --version ${CC_VER} --signature-policy ${CC_EP}

    # first call negated as it fails due to specification of validation plugin
    try_fail ${PEER_CMD} lifecycle chaincode checkcommitreadiness -C ${CHAN_ID} --name auction_test --version ${CC_VER} --signature-policy ${CC_EP} -E mock-escc -V fpc-vscc
    try ${PEER_CMD} lifecycle chaincode checkcommitreadiness -C ${CHAN_ID} --name auction_test --version ${CC_VER} --signature-policy ${CC_EP}

    # first call negated as it fails due to specification of validation plugin
    try_fail ${PEER_CMD} lifecycle chaincode commit -C ${CHAN_ID} --name auction_test --version ${CC_VER} --signature-policy ${CC_EP} -E mock-escc -V fpc-vscc
    try ${PEER_CMD} lifecycle chaincode commit -C ${CHAN_ID} --name auction_test --version ${CC_VER} --signature-policy ${CC_EP}

    #first call negated as it fails
    try_fail ${PEER_CMD} lifecycle chaincode createenclave --name wrong-cc-id
    try ${PEER_CMD} lifecycle chaincode createenclave --name auction_test

    # install something non-fpc
    PKG=/tmp/marbles02.tar.gz
    try ${PEER_CMD} lifecycle chaincode package --lang golang --label marbles02 --path github.com/hyperledger/fabric-samples/chaincode/marbles02/go ${PKG}
    try ${PEER_CMD} lifecycle chaincode install ${PKG}
    PKG_ID=$(${PEER_CMD} lifecycle chaincode queryinstalled | awk "/Package ID: marbles02/{print}" | sed -n 's/^Package ID: //; s/, Label:.*$//;p')
    try ${PEER_CMD} lifecycle chaincode approveformyorg -C ${CHAN_ID} --package-id ${PKG_ID} --name marbles02 --version 0
    try ${PEER_CMD} lifecycle chaincode checkcommitreadiness -C ${CHAN_ID} --name marbles02 --version 0
    try ${PEER_CMD} lifecycle chaincode commit -C ${CHAN_ID} --name marbles02 --version 0

    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n auction_test -c '{"Args":["init", "MyAuctionHouse"]}' --waitForEvent
    check_result "OK"

    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n auction_test -c '{"Args":["create", "MyAuction"]}' --waitForEvent
    check_result "OK"

    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n marbles02 -c '{"Args":["initMarble","marble1","blue","35","tom"]}' --waitForEvent

    # install examples/echo
    CC_PATH=${FPC_TOP_DIR}/examples/echo/_build/lib/
    CC_VER="$(cat ${CC_PATH}/mrenclave)"
    PKG=/tmp/echo_test.tar.gz

    try ${PEER_CMD} lifecycle chaincode package --lang fpc-c --label echo_test --path ${FPC_TOP_DIR}/examples/echo/_build/lib ${PKG}
    try ${PEER_CMD} lifecycle chaincode install ${PKG}
    PKG_ID=$(${PEER_CMD} lifecycle chaincode queryinstalled | awk "/Package ID: echo_test/{print}" | sed -n 's/^Package ID: //; s/, Label:.*$//;p')

    # first call negated as it fails due to specification of validation plugin
    try_fail ${PEER_CMD} lifecycle chaincode approveformyorg -C ${CHAN_ID} --package-id ${PKG_ID} --name echo_test --version ${CC_VER} -E mock-escc -V fpc-vscc
    try ${PEER_CMD} lifecycle chaincode approveformyorg -C ${CHAN_ID} --package-id ${PKG_ID} --name echo_test --version ${CC_VER}

    # first call negated as it fails due to specification of validation plugin
    try_fail ${PEER_CMD} lifecycle chaincode checkcommitreadiness -C ${CHAN_ID} --name echo_test --version ${CC_VER} -E mock-escc -V fpc-vscc
    try ${PEER_CMD} lifecycle chaincode checkcommitreadiness -C ${CHAN_ID} --name echo_test --version ${CC_VER}

    # first call negated as it fails due to specification of validation plugin
    try_fail ${PEER_CMD} lifecycle chaincode commit -C ${CHAN_ID} --name echo_test --version ${CC_VER} -E mock-escc -V fpc-vscc
    try ${PEER_CMD} lifecycle chaincode commit -C ${CHAN_ID} --name echo_test --version ${CC_VER}

    #first call negated as it fails
    try_fail ${PEER_CMD} lifecycle chaincode createenclave --name wrong-cc-id
    try ${PEER_CMD} lifecycle chaincode createenclave --name echo_test

    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n echo_test -c '{"Args": ["moin"]}' --waitForEvent
    check_result "moin"

    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n auction_test -c '{"Args":["submit", "MyAuction", "JohnnyCash0", "0"]}' --waitForEvent
    check_result "OK"

    try ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n marbles02 -c '{"Args":["readMarble","marble1"]}' --waitForEvent

    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n echo_test -c '{"Args": ["bonjour"]}' --waitForEvent
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
if [[ "$FAILURES" == 0 ]]; then
    yell "Deployement test PASSED"
else
    yell "Deployement test had ${FAILURES} failures"
    exit 1
fi
exit 0

