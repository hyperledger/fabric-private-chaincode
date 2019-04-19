#!/bin/bash
SCRIPTDIR="$(dirname $(readlink --canonicalize ${BASH_SOURCE}))"
. ${SCRIPTDIR}/common.sh

rm -rf /tmp/hyperledger/*
ORDERER_GENERAL_GENESISPROFILE=SampleDevModeSolo ${FABRIC_BIN_DIR}/orderer
