#!/bin/bash

# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

SCRIPTDIR="$(dirname $(readlink --canonicalize ${BASH_SOURCE}))"
FPC_PATH="${SCRIPTDIR}/.."
FABRIC_SCRIPTDIR="${FPC_PATH}/fabric/bin/"

: ${FABRIC_CFG_PATH:="${FPC_PATH}/integration/config"}

. ${FABRIC_SCRIPTDIR}/lib/common_utils.sh
. ${FABRIC_SCRIPTDIR}/lib/common_ledger.sh
TEST_NAME=TestGoClientSDK
CHAINCODES=(auction_test echo_test)
CC_PATH=${FPC_PATH}/examples/auction/_build/lib/
CC_LANG=fpc-c
CC_VER="$(cat ${CC_PATH}/mrenclave)"
CC_SEQ="1"
CC_EP="OR('SampleOrg.member')" # note that we use .member as NodeOUs is disabled with the crypto material used in the integration tests.

setup_chaincodes() {

    for CC_ID in ${CHAINCODES[@]}; do 
        say "Install ${CC_ID}"
        PKG=/tmp/${CC_ID}.tar.gz
        try ${PEER_CMD} lifecycle chaincode package --lang ${CC_LANG} --label ${CC_ID} --path ${CC_PATH} ${PKG}
        try ${PEER_CMD} lifecycle chaincode install ${PKG}
        PKG_ID=$(${PEER_CMD} lifecycle chaincode queryinstalled | awk "/Package ID: ${CC_ID}/{print}" | sed -n 's/^Package ID: //; s/, Label:.*$//;p')
        try ${PEER_CMD} lifecycle chaincode approveformyorg -o ${ORDERER_ADDR} -C ${CHAN_ID} --package-id ${PKG_ID} --name ${CC_ID} --version ${CC_VER} --sequence ${CC_SEQ} --signature-policy ${CC_EP}
        try ${PEER_CMD} lifecycle chaincode checkcommitreadiness -C ${CHAN_ID} --name ${CC_ID} --version ${CC_VER} --sequence ${CC_SEQ} --signature-policy ${CC_EP}
        try ${PEER_CMD} lifecycle chaincode commit -o ${ORDERER_ADDR} -C ${CHAN_ID} --name ${CC_ID} --version ${CC_VER} --sequence ${CC_SEQ} --signature-policy ${CC_EP}
    done  

    say "Init enclave for ${CHAINCODES[0]}" 
    # only init enclave for the auction test
    try ${PEER_CMD} lifecycle chaincode initEnclave -o ${ORDERER_ADDR} --peerAddresses "localhost:7051" --name ${CHAINCODES[0]}

    try ${PEER_CMD} lifecycle chaincode querycommitted -C ${CHAN_ID}
}

run_test() {
    # call sdk test
    try cd client_sdk/go && go test -v -run ${TEST_NAME}

    # cleanup
    rm -rf keystore wallet
}

# 1. prepare
para
say "Preparing ${TEST_NAME} Test ..."
# - clean up relevant docker images
docker_clean ${ERCC_ID}

trap ledger_shutdown EXIT

para
say "Run ${TEST_NAME} test"

say "- setup ledger"
ledger_init

say "- setup FPC chaincodes"
setup_chaincodes

say "- run test"
run_test

say "- shutdown ledger"
ledger_shutdown

para
yell "${TEST_NAME} test PASSED"
exit 0
