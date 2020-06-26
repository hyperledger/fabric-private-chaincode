#!/bin/bash
#
# Copyright IBM Corp All Rights Reserved
# Copyright 2019 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0
#

export PATH=${SCRIPT_DIR}/../bin:${PWD}:$PATH
export FABRIC_CFG_PATH=${SCRIPT_DIR}/../network-config
export FPC_PATH="${FPC_PATH:-${SCRIPT_DIR}/../../..}"
if [ -z ${DOCKERD_FPC_PATH+x} ]; then export DOCKERD_FPC_PATH=${FPC_PATH}; fi

# load general script utils, .e.g, die, say, yell and alike ...
. ${FPC_PATH}/fabric/bin/lib/common_utils.sh

# SGX mode: make sure it is set so we consistently use the same value also when we call make
# Note: make might define in config*.mk its own value without necessarily it being an env variable
export SGX_MODE=${SGX_MODE:=SIM}

# Variables which we allow the caller override ..
export FABRIC_VERSION=${FABRIC_VERSION:=2.1.1}
export FPC_VERSION=${FPC_VERSION:=latest}
export NET_ID=${NET_ID:=dev}
# above should correspond to 'peer->networkId' in ../network-config/core.yaml. when refactoring, we might
# consider auto-extracting as done in ../../fabric/bin/lib/common_ledger.sh ...
export CHANNEL_NAME=${CHANNEL_NAME:=mychannel}
export NODE_WALLETS=${NODE_WALLETS:=${SCRIPT_DIR}/../node-sdk/wallet}
export DOCKER_COMPOSE_OPTS=${DOCKER_COMPOSE_OPTS:=}

# don't rewrite paths for Windows Git Bash users
export MSYS_NO_PATHCONV=1

# defaults for services used
export USE_FPC=${USE_FPC:=true} 
export USE_SGX_HW=$(if [ "${SGX_MODE}" = "HW" ]; then echo true; else echo false; fi)
export USE_COUCHDB=${USE_COUCHDB:=false} 
export USE_EXPLORER=${USE_EXPLORER:=false} 

if [[ $USE_FPC = false ]]; then
    export FPC_CONFIG=""
    export PEER_CMD="peer"
else
    export FPC_CONFIG="-fpc"
    export PEER_CMD=/project/src/github.com/hyperledger-labs/fabric-private-chaincode/fabric/bin/peer.sh
    # FABRIC_BIN_DIR needs to be set for FPC Peer CMD
    export FABRIC_BIN_DIR=/project/src/github.com/hyperledger/fabric/build/bin
fi

export NETWORK_CONFIG=${FPC_PATH}/utils/docker-compose/network-config
export DOCKERD_NETWORK_CONFIG=${DOCKERD_FPC_PATH}/utils/docker-compose/network-config
export COMPOSE_PROJECT_NAME="fabric$(echo ${FPC_CONFIG} | sed 's/[^a-zA-Z0-9]//g')"
# Note: COMPOSE_PROJECT_NAME should have only chars in [a-zA-Z0-9], see https://github.com/docker/compose/issues/4002

export DOCKER_COMPOSE_CMD="docker-compose"
export DOCKER_COMPOSE_OPTS="${DOCKER_COMPOSE_OPTS:+${DOCKER_COMPOSE_OPTS} }-f ${NETWORK_CONFIG}/docker-compose.yml"
if ${USE_COUCHDB}; then
	export DOCKER_COMPOSE_OPTS="${DOCKER_COMPOSE_OPTS} -f ${NETWORK_CONFIG}/docker-compose-couchdb.yml"
fi
if ${USE_EXPLORER}; then
	export DOCKER_COMPOSE_OPTS="${DOCKER_COMPOSE_OPTS} -f ${NETWORK_CONFIG}/docker-compose-explorer.yml"
fi
if ${USE_SGX_HW}; then
	SGX_DEVICE_PATH=$(if [ -e "/dev/isgx" ]; then echo "/dev/isgx"; elif [ -e "/dev/sgx" ]; then echo "/dev/sgx"; else echo "none"; fi)
	if [ "${SGX_DEVICE_PATH}" = "none" ]; then 
		echo "ERROR: SGX_MODE is HW but no sgx device found";
		exit 1;
	fi	
	# Note: the echo is hack to get an error with error message back. The _ make
	# sure the whole message surfaces to the console
	export DOCKER_COMPOSE_CMD="env SGX_DEVICE_PATH=${SGX_DEVICE_PATH} SGX_CONFIG_ROOT="${DOCKERD_FPC_PATH}/config/ias" ${DOCKER_COMPOSE_CMD}"
	export DOCKER_COMPOSE_OPTS="${DOCKER_COMPOSE_OPTS} -f ${NETWORK_CONFIG}/docker-compose-sgx-hw.yml"
fi
export DOCKER_COMPOSE="${DOCKER_COMPOSE_CMD} ${DOCKER_COMPOSE_OPTS}"
