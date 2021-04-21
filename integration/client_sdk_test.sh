#!/bin/bash

# Copyright IBM Corp. All Rights Reserved.
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
CC_PATH=${FPC_PATH}/samples/chaincode/auction/_build/lib/
CC_LANG=fpc-c
CC_VER="$(cat ${CC_PATH}/mrenclave)"
CC_SEQ="1"
CC_EP="OR('SampleOrg.member')" # note that we use .member as NodeOUs is disabled with the crypto material used in the integration tests.

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

say "- run test"
run_test

say "- shutdown ledger"
ledger_shutdown

para
yell "${TEST_NAME} test PASSED"
exit 0
