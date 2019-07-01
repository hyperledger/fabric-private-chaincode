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

CC_ID=ecc
# TODO: once issue #86 is fixed, change above to ecc_echo_test

#this is the path that will be used for the docker build of the chaincode enclave
ENCLAVE_SO_PATH=examples/echo/_build/lib/

CC_VERS=0
num_rounds=10

echo_test() {
    expect_switcheroo_fail=$1

    # install, init, and register (auction) chaincode
    try ${PEER_CMD} chaincode install -l fpc-c -n ${CC_ID} -v ${CC_VERS} -p github.com/hyperledger-labs/fabric-private-chaincode/ecc/ -e ${ENCLAVE_SO_PATH}
    sleep 3

    # init is special case as it might expectedly fail if switcheroo wasn't done yet ..
    try ${PEER_CMD} chaincode instantiate -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -v ${CC_VERS} -c '{"args":["init"]}' -V ecc-vscc
    sleep 3

    try ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["setup", "ercc"]}' --waitForEvent

    try ${PEER_CMD} chaincode query -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["getEnclavePk"]}'

    say "do echos"
    for (( i=1; i<=$num_rounds; i++ ))
    do
        # echos
        try ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args": ["[\"echo-$i\"]", ""]}' --waitForEvent
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
