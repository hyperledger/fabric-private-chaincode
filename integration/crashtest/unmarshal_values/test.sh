#!/usr/bin/env bash

if [[ -z "${FPC_PATH}" ]]; then
  echo "Error: FPC_PATH not set"; exit 1
fi

FABRIC_CFG_PATH="${FPC_PATH}/integration/config"
FABRIC_SCRIPTDIR="${FPC_PATH}/fabric/bin/"

. "${FABRIC_SCRIPTDIR}"/lib/common_utils.sh
. "${FABRIC_SCRIPTDIR}"/lib/common_ledger.sh

# this is the path points to FPC chaincode binary
CC_PATH=${FPC_PATH}/integration/crashtest/unmarshal_values/_build/lib/

CC_ID="${CC_ID-crash}"
CC_VER="$(cat "${CC_PATH}"/mrenclave)"
CC_EP="OR('SampleOrg.member')"
CC_SEQ="1"

run_test() {
    say "- install chaincode"
    PKG="/tmp/${CC_ID}.tar.gz"
    ${PEER_CMD} lifecycle chaincode package --lang fpc-c --label "${CC_ID}" --path "${CC_PATH}" "${PKG}"
    ${PEER_CMD} lifecycle chaincode install "${PKG}"

    PKG_ID=$(${PEER_CMD} lifecycle chaincode queryinstalled | awk "/Package ID: ${CC_ID}/{print}" | sed -n 's/^Package ID: //; s/, Label:.*$//;p')

    ${PEER_CMD} lifecycle chaincode approveformyorg -o "${ORDERER_ADDR}" -C "${CHAN_ID}" --package-id "${PKG_ID}" --name "${CC_ID}" --version "${CC_VER}" --sequence ${CC_SEQ} --signature-policy ${CC_EP}
    ${PEER_CMD} lifecycle chaincode checkcommitreadiness -C "${CHAN_ID}" --name "${CC_ID}" --version "${CC_VER}" --sequence ${CC_SEQ} --signature-policy ${CC_EP}
    ${PEER_CMD} lifecycle chaincode commit -o "${ORDERER_ADDR}" -C "${CHAN_ID}" --name "${CC_ID}" --version "${CC_VER}" --sequence ${CC_SEQ} --signature-policy ${CC_EP}

    ${PEER_CMD} lifecycle chaincode initEnclave -o "${ORDERER_ADDR}" --peerAddresses "localhost:7051" --name "${CC_ID}"

    say "- interact with the FPC chaincode using our client app"

    export CC_ID
    export CHAN_ID
    try go test -v ./test
}

trap ledger_shutdown EXIT

say "Setup ledger ..."
ledger_init

para
say "Run test ..."
run_test

para
say "Shutdown ledger ..."
ledger_shutdown

yell "Test PASSED"

exit 0