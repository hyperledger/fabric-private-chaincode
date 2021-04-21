# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

# assume
# - FPC_PATH is defined
# - FABRIC_CFG_PATH is defined
# optional config overrides
# - FABRIC_PATH
# - FABRIC_BIN_DIR

[ -d "${FPC_PATH}" ] || (echo "FPC_PATH not properly defined in '${FPC_PATH}'"; exit 1; )

: ${FABRIC_PATH:="${FPC_PATH}/../../hyperledger/fabric/"}
: ${FABRIC_BIN_DIR:="${FPC_PATH}/fabric/_internal/bin"}
: ${FABRIC_UTIL_BIN_DIR:="${FPC_PATH}/utils/fabric"}

FABRIC_SCRIPTDIR="${FPC_PATH}/fabric/bin/"

. ${FABRIC_SCRIPTDIR}/lib/common_utils.sh


# Check consistency of variables affecting ledger
#-----------------------------------------------------
# must be called _after_ parse_fabric_config & define_common_vars
ledger_precond_check() {
	[ -d "${FABRIC_BIN_DIR}" ] || die "FABRIC_BIN_DIR not properly defined as '${FABRIC_BIN_DIR}'"
	[ -x "${FABRIC_BIN_DIR}/peer" ] || die "peer command does not exist in '${FABRIC_BIN_DIR}'"
	[ -x "${FABRIC_BIN_DIR}/orderer" ] || die "orderer command does not exist in '${FABRIC_BIN_DIR}'"
	[ -x "${FABRIC_BIN_DIR}/configtxgen" ] || die "configtxgen command does not exist in '${FABRIC_BIN_DIR}'"
	[ -x "${FABRIC_UTIL_BIN_DIR}/peer-cli-assist" ] || die "peer-cli-assist command does not exist in '${FABRIC_UTIL_BIN_DIR}'"
	[ ! -z "${FABRIC_STATE_DIR}" ] || die "Undefined fabric ledger state directory '${FABRIC_STATE_DIR}'"
}


# Defaults et al.
#----------------------------------
# must be called _after_ parse_fabric_config
define_common_vars() {
    # use our wrapper for commands, so provide convience functions ..
    # note: as we might have defined FABRIC_CFG_PATH in this script, we have the pass it along!
    PEER_CMD="env FABRIC_CFG_PATH=${FABRIC_CFG_PATH} ${FABRIC_SCRIPTDIR}/peer.sh"
    PEER_ASSIST_CMD="${FABRIC_UTIL_BIN_DIR}/peer-cli-assist"
    ORDERER_CMD="env FABRIC_CFG_PATH=${FABRIC_CFG_PATH} ${FABRIC_SCRIPTDIR}/orderer.sh"
    CONFIGTXGEN_CMD="env FABRIC_CFG_PATH=${FABRIC_CFG_PATH} ${FABRIC_SCRIPTDIR}/configtxgen.sh"

    # NOTE: following variables can be overriden by defining them _before_ sourcing common_ledger.sh ..
    : ${ORDERER_ADDR:="localhost:7050"}
    : ${CHAN_ID:="mychannel"}
    : ${ERCC_ID:="ercc"}
    : ${ERCC_VERSION:="0"}
    : ${ERCC_SEQUENCE:="1"}

    ORDERER_PID_FILE="${FABRIC_STATE_DIR}/orderer.pid"
    ORDERER_LOG_OUT="${FABRIC_STATE_DIR}/orderer.out"
    ORDERER_LOG_ERR="${FABRIC_STATE_DIR}/orderer.err"
    PEER_PID_FILE="${FABRIC_STATE_DIR}/peer.pid"
    PEER_LOG_OUT="${FABRIC_STATE_DIR}/peer.out"
    PEER_LOG_ERR="${FABRIC_STATE_DIR}/peer.err"

    CHANNEL_TX="${FABRIC_STATE_DIR}/${CHAN_ID}.tx"
    CHANNEL_BLOCK="${FABRIC_STATE_DIR}/${CHAN_ID}.block"
}


# Fabric config parsing
#----------------------------
# input parameter is FABRIC_CFG_PATH directory
# when succesfull, will have defined following variables
# - SPID_FILE
# - API_KEY_FILE
# - NET_ID
# - PEER_ID
# - FABRIC_STATE_DIR
parse_fabric_config() {
    CONFIG_DIR=$1

    [ -d "${CONFIG_DIR}" ] || die "provided fabric config dir '${CONFIG_DIR}' does not exist"
    [ -e "${CONFIG_DIR}/core.yaml" ] || die "no 'core.yaml' in provided fabric config dir '${CONFIG_DIR}'"

    FABRIC_STATE_DIR=$(perl -0777 -n -e 'm/fileSystemPath:\s*(\S+)/i && print "$1"' ${CONFIG_DIR}/core.yaml)

    PEER_ID=$(perl -0777 -n -e 'm/id:\s*(\S+)/i && print "$1"' ${CONFIG_DIR}/core.yaml)
    NET_ID=$(perl -0777 -n -e 'm/networkId:\s*(\S+)/i && print "$1"' ${CONFIG_DIR}/core.yaml)
}

# Clean-up docker images
#----------------------------
# input parameter is name of chain-code for which docker image(s) should be cleaned up
docker_clean() {
    cc_name=$1
    docker_image=$(docker images | grep -- ${NET_ID}-${PEER_ID}-${cc_name}- | awk '{print $1;}')
    if [ ! -z "${docker_image}" ]; then
	docker rmi -f "${docker_image}";
    fi
}

# Initialize ledger
#--------------------------
# TODO (eventually: split below into ledger_init and ledger_start
#   so we can shutdown and restart without reseting state.
#   Given the way ercc currently runs, though, it's not immediately
#   clear how to do so and with switcheroo there are anyway not
#   really good scenarios where we currently care ...
ledger_init() {

    pushd ${FABRIC_CFG_PATH}

    # 1. clean up any prior state
    #     - ledger state and log files from peer & orderer
    [ ! -z "${FABRIC_STATE_DIR}" ] || die "FABRIC_STATE_DIR not defined" # just in case ..
    try mkdir -p ${FABRIC_STATE_DIR}
    try rm -rf ${FABRIC_STATE_DIR}/*
    #     - external builder artifacts
    #       Note: FABRIC_STATE_DIR which will be ledger specific and allows for
    #         multiple concurrent ledgers (assuming their core.yaml have 
    #         different fileSsystemPaths) these artifacts are also shared 
    #         _across_ ledger definitions (core.yaml).
    #         Contrary to that, external builder are ledger instance agnostic, hence we also
    #         provide an opt-out for the wiping to allow running current ledgers by
    #         defining an environment variable FPC_SKIP_EXTBUILDER_CLEANUP ..
    if [ -z ${FPC_SKIP_EXTBUILDER_CLEANUP+x} ]; then
    	try rm -rf /tmp/fpc-extbuilder.*
    fi

    try ${FABRIC_BIN_DIR}/cryptogen generate --config=./crypto-config.yaml --output="${FABRIC_STATE_DIR}/organizations"
    # 2. start orderer
    # - Creating a genesisblock for orderer-system-channel......."
    try ${CONFIGTXGEN_CMD} -outputBlock ${FABRIC_STATE_DIR}/genesisblock -profile SampleDevModeEtcdRaft -channelID orderer-system-channel
    sleep 1
    ${ORDERER_CMD} 1>${ORDERER_LOG_OUT} 2>${ORDERER_LOG_ERR} &
    export ORDERER_PID=$!
    echo "${ORDERER_PID}" > ${ORDERER_PID_FILE}
    sleep 1
    kill -0 ${ORDERER_PID} || die "Orderer quit too quickly: (for log see ${ORDERER_LOG_OUT} & ${ORDERER_LOG_ERR})"

    # 3. start peer
    LD_LIBRARY_PATH=${LD_LIBRARY_PATH:+"$LD_LIBRARY_PATH:"}${FPC_PATH}/tlcc/enclave/lib \
		   ${PEER_CMD} node start 1>${PEER_LOG_OUT} 2>${PEER_LOG_ERR} &
    export PEER_PID=$!
    echo "${PEER_PID}" > ${PEER_PID_FILE}
    sleep 1
    kill -0 ${PEER_PID} || die "Peer quit too quickly: (for log see ${PEER_LOG_OUT} & ${PEER_LOG_ERR})"

    # 4. Setup channel
    # - create channel tx
    try ${CONFIGTXGEN_CMD} -channelID ${CHAN_ID} -profile SampleSingleMSPChannel -outputCreateChannelTx ${CHANNEL_TX}
    # - create genesis block, only by one peer
    try ${PEER_CMD} channel create -o ${ORDERER_ADDR} -c ${CHAN_ID} -f ${CHANNEL_TX} --outputBlock ${CHANNEL_BLOCK}
    # - every peer will have to join (after having received mychannel.block out-of-band)
    try ${PEER_CMD} channel join -b ${CHANNEL_BLOCK}
    sleep 3

    popd # ${FABRIC_CFG_PATH}
}

# Shutdown ledger (i.e., orderer & peer)
#-----------------------------------------
ledger_shutdown() {
    if [ -z "${PEER_PID}" ]; then
	# maybe we have it in pidfile ..
	PEER_PID=$(cat ${PEER_PID_FILE} 2> /dev/null)
    fi
    if [ ! -z "${PEER_PID}" ]; then
	kill  ${PEER_PID}
	wait  ${PEER_PID} 2> /dev/null
	unset PEER_PID
	rm ${PEER_PID_FILE}
    fi

    if [ -z "${ORDERER_PID}" ]; then
	# maybe we have it in pidfile ..
	ORDERER_PID=$(cat ${ORDERER_PID_FILE} 2> /dev/null)
    fi
    if [ ! -z "${ORDERER_PID}" ]; then
	kill ${ORDERER_PID}
	wait ${ORDERER_PID} 2> /dev/null
	unset ORDERER_PID
	rm ${ORDERER_PID_FILE}
    fi
}

# Check the chaincode's response (ResponseData) of
#   peer chaincode invoke/query
# (executed via 'try_r/try_out_r' macro) against expected result.
# In case of failures, tries to increment integer variable NUM_FAILURES. Increment in all cases NUM_TESTS
check_result() {
    if [ "${RESPONSE_TYPE}" == "out+err" ]; then
	CI_RESPONSE=$(parse_invoke_result_from_log "${RESPONSE}")
	CONTEXT=" context: '${RESPONSE}'"
    else
	CI_RESPONSE="${RESPONSE}"
	CONTEXT=""
    fi

    if [[ ${CI_RESPONSE} == "$1" ]]; then
	gecho "PASSED"
    else
	recho "FAILED (expected '${1}', got '${CI_RESPONSE}'${CONTEXT})"
        export NUM_FAILURES=$(($NUM_FAILURES+1))
    fi
    export NUM_TESTS=$(($NUM_TESTS+1))
}

parse_invoke_result_from_log() {
    RESPONSE="$1"
    RESPONSE=${RESPONSE##*payload:\"}
    RESPONSE=${RESPONSE%%\"*}
    echo "${RESPONSE}"
}


# "Main"
parse_fabric_config "${FABRIC_CFG_PATH}"
ledger_precond_check
define_common_vars

