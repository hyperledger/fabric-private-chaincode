#!/bin/bash
# Copyright Intel Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
SCRIPTDIR="$(dirname $(readlink --canonicalize ${BASH_SOURCE}))"
FPC_TOP_DIR="${SCRIPTDIR}/.."
CONFIG_HOME="${SCRIPTDIR}/config"

. ${SCRIPTDIR}/common_utils.sh
. ${SCRIPTDIR}/common_ledger.sh

CC_ID=ecc
# TODO: once issue #86 is fixed, change above to ecc_auction_test
CC_VERS=0
num_rounds=3
num_clients=10

auction_test() {
    expect_switcheroo_fail=$1

    # install, init, and register (auction) chaincode
    # install some dummy chaincode (we manually need to create the image)
    try ${PEER_CMD} chaincode install -n ${CC_ID} -v ${CC_VERS} -p github.com/hyperledger/fabric/examples/chaincode/go/example02/cmd
    sleep 3

    # init is special case as it might expectedly fail if switcheroo wasn't done yet ..
    stdinerr=$(${PEER_CMD} chaincode instantiate -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -v ${CC_VERS} -c '{"args":["init"]}' -V ecc-vscc 2>&1)
    rc=$?
    echo "${stdinerr}"
    if [ ${rc} != 0 ]; then
	if ( [ "${expect_switcheroo_fail}" -eq 1 ] &&
	     $(echo ${stdinerr} | grep -q 'Incorrect number of arguments. Expecting 4')); then
	    say "INFO: switcheroo-related failure happened but was expected";
	    return
	else
	    die "ABORT: Unexpected failure (rc=${rc})"
	fi
    fi

    sleep 3

    try ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["setup", "ercc"]}'
    sleep 3

    try ${PEER_CMD} chaincode query -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["getEnclavePk"]}'

    # create auction
    try ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args": ["[\"create\",\"MyAuction\"]", ""]}'
    sleep 3

    say "invoke submit"
    try ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["[\"submit\",\"MyAuction\", \"JohnnyCash0\", \"0\"]", ""]}'
    sleep 3

    for (( i=1; i<=$num_rounds; i++ ))
    do
        b="$(($i%$num_clients))"
        try ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["[\"submit\",\"MyAuction\", \"JohnnyCash'$b'\", \"'$b'\"]", ""]}'
    done
    sleep 3

    try ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["[\"close\",\"MyAuction\"]",""]}'
    sleep 3

    say "invoke eval"
    for (( i=1; i<=1; i++ ))
    do
        try ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["[\"eval\",\"MyAuction\"]", ""]}'
    done
}

# 1. prepare
para
say "Preparing Auction Test ..."
# - clean up relevant docker images
docker_clean ${ERCC_ID}
docker_clean ${CC_ID}

trap ledger_shutdown EXIT


# 2. First run, this should fail due to docker-switcheroo
para
say "FIRST RUN: this should fail due to docker-switcheroo ..."

say "- setup ledger"
ledger_init

say "- auction test"
auction_test 1
#   should fail with
#    Error: could not assemble transaction, err proposal response was not successful, error code 500, msg transaction returned with failure: Incorrect number of arguments. Expecting 4

say "- shutdown ledger"
ledger_shutdown

say "- do switcheroo"
(cd ${FPC_TOP_DIR}/ecc; CC_NAME=${CC_ID} make docker) || die "ERROR: cannot perform switcheroo"


# 3. Second run, this should work ..
para
say "SECOND RUN: this should (hopefully :-) succeed ..."

say "- setup ledger"
ledger_init

say "- run auction test"
auction_test 0

say "- shutdown ledger"
ledger_shutdown

para
yell "Auction test PASSED"

exit 0
