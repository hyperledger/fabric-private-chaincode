#!/bin/bash
#
# Copyright IBM Corp All Rights Reserved
# Copyright Intel Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#

export PATH=${SCRIPT_DIR}/../bin:${PWD}:$PATH
export FABRIC_CFG_PATH=${SCRIPT_DIR}/../network-config
export FPC_PATH="${FPC_PATH:-${SCRIPT_DIR}/../../..}"

# Variables which we allow the caller override ..
export FABRIC_VERSION=${FABRIC_VERSION:=1.4.3}
export CHANNEL_NAME=${CHANNEL_NAME:=mychannel}
export NODE_WALLETS=${NODE_WALLETS:=${SCRIPT_DIR}/../node-sdk/wallet}
export DOCKER_COMPOSE_OPTS=${DOCKER_COMPOSE_OPTS:=}

# don't rewrite paths for Windows Git Bash users
export MSYS_NO_PATHCONV=1

# default configs
export USE_FPC=${USE_FPC:=true} 
export USE_COUCHDB=${USE_COUCHDB:=false} 

if [[ $USE_FPC = false ]]; then
    export FPC_CONFIG=""
    export PEER_CMD="peer"
else
    export FPC_CONFIG="-fpc"
    export PEER_CMD=/project/src/github.com/hyperledger-labs/fabric-private-chaincode/fabric/bin/peer.sh
    # FABRIC_BIN_DIR needs to be set for FPC Peer CMD
    export FABRIC_BIN_DIR=/project/src/github.com/hyperledger/fabric/.build/bin
fi

export NETWORK_CONFIG=${SCRIPT_DIR}/../network-config
export COMPOSE_PROJECT_NAME="fabric$(echo ${FPC_CONFIG} | sed 's/[^a-zA-Z0-9]//g')"
# Note: COMPOSE_PROJECT_NAME should have only chars in [a-zA-Z0-9], see https://github.com/docker/compose/issues/4002

export DOCKER_COMPOSE_CMD="docker-compose"
export DOCKER_COMPOSE_OPTS="${DOCKER_COMPOSE_OPTS:+${DOCKER_COMPOSE_OPTS} }-f ${NETWORK_CONFIG}/docker-compose.yml"
if $USE_COUCHDB; then
	export DOCKER_COMPOSE_OPTS="${DOCKER_COMPOSE_OPTS} -f ${NETWORK_CONFIG}/docker-compose-couchdb.yml"
fi
export DOCKER_COMPOSE="${DOCKER_COMPOSE_CMD} ${DOCKER_COMPOSE_OPTS}"
