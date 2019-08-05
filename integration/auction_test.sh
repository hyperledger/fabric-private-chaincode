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

CC_ID=auction_test

#this is the path that will be used for the docker build of the chaincode enclave
ENCLAVE_SO_PATH=examples/auction/_build/lib/

CC_VERS=0
num_rounds=3
num_clients=10

auction_test() {
    # install, init, and register (auction) chaincode
    try ${PEER_CMD} chaincode install -l fpc-c -n ${CC_ID} -v ${CC_VERS} -p ${ENCLAVE_SO_PATH}
    sleep 3

    try ${PEER_CMD} chaincode instantiate -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -v ${CC_VERS} -c '{"args":["init"]}' -V ecc-vscc
    sleep 3

    try ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["setup", "ercc"]}' --waitForEvent

    try ${PEER_CMD} chaincode query -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["getEnclavePk"]}'

    # create auction
    try ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args": ["[\"create\",\"MyAuction\"]", ""]}' --waitForEvent

    say "invoke submit"
    try ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["[\"submit\",\"MyAuction\", \"JohnnyCash0\", \"0\"]", ""]}' --waitForEvent

    for (( i=1; i<=$num_rounds; i++ ))
    do
        b="$(($i%$num_clients))"
        try ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["[\"submit\",\"MyAuction\", \"JohnnyCash'$b'\", \"'$b'\"]", ""]}' # Don't do --waitForEvent, so potentially there is some parallelism here ..
    done

    try ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["[\"close\",\"MyAuction\"]",""]}' --waitForEvent

    say "invoke eval"
    for (( i=1; i<=1; i++ ))
    do
        try ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["[\"eval\",\"MyAuction\"]", ""]}'  # Don't do --waitForEvent, so potentially there is some parallelism here ..
    done
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
yell "Auction test PASSED"

exit 0
