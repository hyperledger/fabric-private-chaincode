#!/bin/bash
# Copyright IBM Corp. All Rights Reserved.
# Copyright Intel Corp. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

SCRIPTDIR="$(dirname $(readlink --canonicalize ${BASH_SOURCE}))"
. ${SCRIPTDIR}/common.sh

rm -rf /tmp/hyperledger/production/*
ORDERER_GENERAL_GENESISPROFILE=SampleDevModeSolo ${FABRIC_BIN_DIR}/orderer
