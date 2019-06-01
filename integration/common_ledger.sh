# Copyright Intel Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

# assume
# - FPC_TOP_DIR is defined
# - CONFIG_HOME is defined
# - common_utils.sh is loaded
# optional config overrides
# - FABRIC_BIN_DIR
# - FABRIC_STATE_DIR

: ${FABRIC_BIN_DIR:="${FPC_TOP_DIR}/../../hyperledger/fabric/.build/bin"}
: ${FABRIC_STATE_DIR:="/tmp/hyperledger/test/"}

PEER_CMD=${FABRIC_BIN_DIR}/peer
ORDERER_CMD=${FABRIC_BIN_DIR}/orderer
CONFIGTXGEN_CMD=${FABRIC_BIN_DIR}/configtxgen

ORDERER_ADDR=localhost:7050
CHAN_ID=mychannel
ERCC_ID=ercc
ERCC_VERSION=0

ORDERER_LOG_OUT="${FABRIC_STATE_DIR}/orderer.out"
ORDERER_LOG_ERR="${FABRIC_STATE_DIR}/orderer.err"
PEER_LOG_OUT="${FABRIC_STATE_DIR}/peer.out"
PEER_LOG_ERR="${FABRIC_STATE_DIR}/peer.err"

CHANNEL_TX="${FABRIC_STATE_DIR}/${CHAN_ID}.tx"
CHANNEL_BLOCK="${FABRIC_STATE_DIR}/${CHAN_ID}.block"

docker_clean() {
    cc_name=$1
    docker_image=$(docker images | grep -- -${cc_name}- | awk '{print $1;}')
    if [ ! -z "${docker_image}" ]; then
	docker rmi -f "${docker_image}";
    fi
}

ledger_precond_check() {
	[ -d "${FPC_TOP_DIR}" ] || die "FPC_TOP_DIR not properly defined as '${FPC_TOP_DIR}'"
	[ -d "${FABRIC_BIN_DIR}" ] || die "FABRIC_BIN_DIR not properly defined as '${FABRIC_BIN_DIR}'"
	[ -x "${PEER_CMD}" ] || die "peer command does not exist in '${FABRIC_BIN_DIR}'"
	[ -x "${ORDERER_CMD}" ] || die "orderer command does not exist in '${FABRIC_BIN_DIR}'"
	[ -x "${CONFIGTXGEN_CMD}" ] || die "configtxgen command does not exist in '${FABRIC_BIN_DIR}'"

	[ -d "${CONFIG_HOME}" ] || die "CONFIG_HOME not properly defined as '${CONFIG_HOME}'"
	[ -e "${CONFIG_HOME}/core.yaml" ] || die "no core.yaml in CONFIG_HOME '${CONFIG_HOME}'"
	spid_file=$(perl -0777 -n -e 'm/spid:\s*file:\s*(\S+)/i && print "$1"' ${CONFIG_HOME}/core.yaml)
	(cd ${CONFIG_HOME} && [ -e ${spid_file} ]) || die "spid not properly configured in ${CONFIG_HOME}/core.yaml or file '${spid_file}' does not exist"
	api_key_file=$(perl -0777 -n -e 'm/apiKey:\s*file:\s*(\S+)/i && print "$1"' ${CONFIG_HOME}/core.yaml)
	(cd ${CONFIG_HOME} && [ -e ${api_key_file} ]) || die "apiKey not properly configured in ${CONFIG_HOME}/core.yaml or apiKey file '${api_key_file}' does not exist"

        [ ! -z "${FABRIC_STATE_DIR}" ] || die "FABRIC_STATE_DIR not defined"
}

# Right now unconditionally check
ledger_precond_check

# TODO (eventually: split below into ledger_init and ledger_start
#   so we can shutdown and restart without reseting state.
#   Given the way ercc currently runs, though, it's not immediately
#   clear how to do so and with switcheroo there are anyway not
#   really good scenarios where we currently care ...
ledger_init() {

    cd ${CONFIG_HOME}

    # 1. clean up any prior state
    [ ! -z "${FABRIC_STATE_DIR}" ] || die "FABRIC_STATE_DIR not defined" # just in case ..
    try mkdir -p ${FABRIC_STATE_DIR}
    try rm -rf ${FABRIC_STATE_DIR}/*

    # 2. start orderer
    ORDERER_GENERAL_GENESISPROFILE=SampleDevModeSolo ${ORDERER_CMD} 1>${ORDERER_LOG_OUT} 2>${ORDERER_LOG_ERR} &
    export ORDERER_PID=$!
    sleep 1
    kill -0 ${ORDERER_PID} || die "Orderer quit too quickly: (for log see ${ORDERER_LOG_OUT} & ${ORDERER_LOG_ERR})"

    # 3. start peer
    LD_LIBRARY_PATH=${LD_LIBRARY_PATH:+"$LD_LIBRARY_PATH:"}${FPC_TOP_DIR}/tlcc/enclave/lib \
		   ${FABRIC_BIN_DIR}/peer node start 1>${PEER_LOG_OUT} 2>${PEER_LOG_ERR} &
    export PEER_PID=$!
    sleep 1
    kill -0 ${PEER_PID} || die "Peer quit too quickly: (for log see ${PEER_LOG_OUT} & ${PEER_LOG_ERR})"

    # 4. Setup channel
    # - create channel tx
    try ${CONFIGTXGEN_CMD} -channelID ${CHAN_ID} -profile SampleSingleMSPChannel -outputCreateChannelTx ${CHANNEL_TX}
    # - create genesis block, only by one peer
    try ${PEER_CMD} channel create -o ${ORDERER_ADDR} -c ${CHAN_ID} -f ${CHANNEL_TX} --outputBlock ${CHANNEL_BLOCK}
    # - every peer will have to join (after having received mychannel.block out-of-band)
    try ${PEER_CMD} channel join -b ${CHANNEL_BLOCK}
    # - every peer's tlcc will have to join as well
    #   IMPORTANT: right now a join is _not_ persistant, so on restart of peer,
    #   it will re-join old channels but tlcc will not!
    try ${PEER_CMD} chaincode query -n tlcc -c '{"Args": ["JOIN_CHANNEL"]}' -C ${CHAN_ID}
    sleep 3

    # 5. ercc
    # - install, once per peer
    try ${PEER_CMD} chaincode install -n ${ERCC_ID} -v ${ERCC_VERSION} -p github.com/hyperledger-labs/fabric-private-chaincode/ercc
    sleep 1
    # - instantiate, once per channel, by single peer/admin
    try ${PEER_CMD} chaincode instantiate -n ${ERCC_ID} -v ${ERCC_VERSION} -c '{"args":["init"]}' -C ${CHAN_ID} -V ercc-vscc
    sleep 3
    # - get SPID (mostly as debug output)
    try ${PEER_CMD} chaincode query -n ${ERCC_ID} -c '{"args":["getSPID"]}' -C ${CHAN_ID}
    sleep 3
}

ledger_shutdown() {
    if [ ! -z "${PEER_PID}" ]; then
	kill  ${PEER_PID}
	unset PEER_PID
    fi
    if [ ! -z "${ORDERER_PID}" ]; then
	kill ${ORDERER_PID}
	unset ORDERER_PID
    fi
}
