#!/bin/bash
SCRIPTDIR="$(dirname $(readlink --canonicalize ${BASH_SOURCE}))"
. ${SCRIPTDIR}/common.sh

# core.yaml does not understand environment variables, hence paths are relative to fabric/sgxconfig,
# so make sure we always start peer from that location, regardless where script is invoked
cd ${GOPATH}/src/github.com/hyperledger-labs/fabric-secure-chaincode/fabric/sgxconfig

LD_LIBRARY_PATH=${GOPATH}/src/github.com/hyperledger-labs/fabric-secure-chaincode/tlcc/enclave/lib ${FABRIC_BIN_DIR}/peer node start
