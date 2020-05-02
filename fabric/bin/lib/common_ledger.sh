# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

# assume
# - FPC_TOP_DIR is defined
# - FABRIC_CFG_PATH is defined
# optional config overrides
# - FABRIC_PATH
# - FABRIC_BIN_DIR

[ -d "${FPC_TOP_DIR}" ] || (echo "FPC_TOP_DIR not properly defined in '${FPC_TOP_DIR}'"; exit 1; )

: ${FABRIC_PATH:="${FPC_TOP_DIR}/../../hyperledger/fabric/"}
: ${FABRIC_BIN_DIR:="${FABRIC_PATH}/build/bin"}

FABRIC_SCRIPTDIR="${FPC_TOP_DIR}/fabric/bin/"

. ${FABRIC_SCRIPTDIR}/lib/common_utils.sh


# Check consistency of variables affecting ledger
#-----------------------------------------------------
# must be called _after_ parse_fabric_config & define_common_vars
ledger_precond_check() {
	[ -d "${FABRIC_BIN_DIR}" ] || die "FABRIC_BIN_DIR not properly defined as '${FABRIC_BIN_DIR}'"
	[ -x "${FABRIC_BIN_DIR}/peer" ] || die "peer command does not exist in '${FABRIC_BIN_DIR}'"
	[ -x "${FABRIC_BIN_DIR}/orderer" ] || die "orderer command does not exist in '${FABRIC_BIN_DIR}'"
	[ -x "${FABRIC_BIN_DIR}/configtxgen" ] || die "configtxgen command does not exist in '${FABRIC_BIN_DIR}'"
	[ ! -z "${FABRIC_STATE_DIR}" ] || die "Undefined fabric ledger state directory '${FABRIC_STATE_DIR}'"
	(cd "${FABRIC_CFG_PATH}" && [ -e "${SPID_FILE}" ]) || die "spid not properly configured in ${FABRIC_CFG_PATH}/core.yaml or file '${SPID_FILE}' does not exist"
	(cd "${FABRIC_CFG_PATH}" && [ -e "${API_KEY_FILE}" ]) || die "apiKey not properly configured in ${FABRIC_CFG_PATH}/core.yaml or apiKey file '${API_KEY_FILE}' does not exist"
}


# Defaults et al.
#----------------------------------
# must be called _after_ parse_fabric_config
define_common_vars() {
    # use our wrapper for commands, so provide convience functions ..
    # note: as we might have defined FABRIC_CFG_PATH in this script, we have the pass it along!
    PEER_CMD="env FABRIC_CFG_PATH=${FABRIC_CFG_PATH} ${FABRIC_SCRIPTDIR}/peer.sh"
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

    SPID_FILE=$(perl -0777 -n -e 'm/spid:\s*file:\s*(\S+)/i && print "$1"' ${CONFIG_DIR}/core.yaml)
    API_KEY_FILE=$(perl -0777 -n -e 'm/apiKey:\s*file:\s*(\S+)/i && print "$1"' ${CONFIG_DIR}/core.yaml)
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
    ORDERER_GENERAL_GENESISPROFILE=SampleDevModeEtcdRaft ${ORDERER_CMD} 1>${ORDERER_LOG_OUT} 2>${ORDERER_LOG_ERR} &
    export ORDERER_PID=$!
    echo "${ORDERER_PID}" > ${ORDERER_PID_FILE}
    sleep 1
    kill -0 ${ORDERER_PID} || die "Orderer quit too quickly: (for log see ${ORDERER_LOG_OUT} & ${ORDERER_LOG_ERR})"

    # 3. start peer
    LD_LIBRARY_PATH=${LD_LIBRARY_PATH:+"$LD_LIBRARY_PATH:"}${FPC_TOP_DIR}/tlcc/enclave/lib \
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
	unset PEER_PID
	rm ${PEER_PID_FILE}
    fi

    if [ -z "${ORDERER_PID}" ]; then
	# maybe we have it in pidfile ..
	ORDERER_PID=$(cat ${ORDERER_PID_FILE} 2> /dev/null)
    fi
    if [ ! -z "${ORDERER_PID}" ]; then
	kill ${ORDERER_PID}
	unset ORDERER_PID
	rm ${ORDERER_PID_FILE}
    fi
}

# Check the chaincode's response (ResponseData) of
#   peer chaincode invoke/query
# (executed via 'try_r' macro) against expected result.
# In case of failures, tries to increment integer variable FAILURES
check_result() {
    # Parse out the Response Data from the payload
    CI_RESPONSE=${RESPONSE}
    CI_RESPONSE=${CI_RESPONSE##*ResponseData\\\":\\\"}
    CI_RESPONSE=${CI_RESPONSE%%\\*}
    # Convert and de-encrypt it
    CI_RESPONSE=$(echo ${CI_RESPONSE} | base64 -d)
    say ${CI_RESPONSE}
    # Test response to expected result
    if [[ ${CI_RESPONSE} == "$1" ]]; then
	gecho "PASSED"
    else
	if [[ ${CI_RESPONSE} == $RESPONSE ]]; then
	    CONTEXT=""
	else
	    CONTEXT=" context: '${RESPONSE}'"
	fi
	recho "FAILED (expected '${1}', got '${CI_RESPONSE}' ${CONTEXT})"
        export FAILURES=$(($FAILURES+1))
    fi
}


# "Main"
parse_fabric_config "${FABRIC_CFG_PATH}"
ledger_precond_check
define_common_vars

