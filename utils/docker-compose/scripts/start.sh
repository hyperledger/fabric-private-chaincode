#!/bin/bash
#
# Copyright IBM Corp All Rights Reserved
# Copyright 2019 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0
#
# Exit on first error, print all commands.
set -e

export SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. ${SCRIPT_DIR}/lib/common.sh

# test pre-conditions and try to remediate
#
if [[ ! $USE_FPC = false ]]; then
    # - existance of FPC peer
    FPC_PEER_NAME="hyperledger/fabric-peer-fpc$(if [ "${SGX_MODE}" = "HW" ]; then echo "-hw"; fi):${FPC_VERSION}" 
    if [ -z "$(docker images -q ${FPC_PEER_NAME})" ]; then
	echo "FPC peer container image '${FPC_PEER_NAME}' does not exist, try to build it ..."
	# if it doesn't exist, build it: note this can take quite some time!!
	pushd "${FPC_PATH}/utils/docker" || die "can't go to peer build location"
	make SGX_MODE=${SGX_MODE} peer || die "can't build peer"
	popd
    fi
    # - existance of boilerplate
    BOILERPLATE_NAME="hyperledger/fabric-private-chaincode-boilerplate-ecc$(if [ "${SGX_MODE}" = "HW" ]; then echo "-hw"; fi):${FPC_VERSION}"
    if [ -z "$(docker images -q ${BOILERPLATE_NAME})" ]; then
	echo "FPC boilerplate container image '${BOILERPLATE_NAME}' does not exist, try to build it ..."
	pushd "${FPC_PATH}/" || die "can't go to fpc-sdk and boilerplate build location"
	make SGX_MODE=${SGX_MODE} fpc-sdk || die "can't build fpc sdk"
	popd
	pushd "${FPC_PATH}/utils/docker" || die "can't go to docker build location"
	make SGX_MODE=${SGX_MODE} || die "can't build docker base images"
	popd
	pushd "${FPC_PATH}/ecc" || die "can't go to fpc-sdk and boilerplate build location"
	make SGX_MODE=${SGX_MODE} docker-boilerplate-ecc || die "can't build boilerplate"
	popd
    fi
fi
# - generated crypto-config files
#   test that we have generated crypto-config. Otherwise below up
#   will create empty files as root which is a PITA if you are running
#   this as non-root
if [ ! -d "${NETWORK_CONFIG}/crypto-config/ordererOrganizations" ]; then
    echo "Could not find crypto configuration, try to generate it..."
    "${FPC_PATH}/utils/docker-compose/scripts/generate.sh"
fi


# The following echo statements are here so users know the environment variables being used and
# can use them with docker-compose.yml directly if desired.
cat <<EOF

# use below environment definition if you want to use additional docker-compose command such as
# '\${DOCKER_COMPOSE} ps' to get status of docker compose network or
# '\${DOCKER_COMPOSE} logs peer0.org1.example.com' for peer logs, e.g., to investigate start errors
export \\
 COMPOSE_PROJECT_NAME="${COMPOSE_PROJECT_NAME}"\\
 FABRIC_VERSION="${FABRIC_VERSION}"\\
 PEER_CMD="${PEER_CMD}"\\
 FPC_CONFIG="${FPC_CONFIG}"\\
 CHANNEL_NAME="${CHANNEL_NAME}"\\
 DOCKER_COMPOSE="${DOCKER_COMPOSE}"

EOF

${DOCKER_COMPOSE} down
${DOCKER_COMPOSE} up -d
${DOCKER_COMPOSE} ps

# wait for Hyperledger Fabric to start
# incase of errors when running later commands, issue export FABRIC_START_TIMEOUT=<larger number>
export FABRIC_START_TIMEOUT=20
sleep ${FABRIC_START_TIMEOUT}


# Command to execute peer cli inside peer container 
# Note: by default the peer runs with peer credentials, so we have to override CORE_PEER_MSPCONFIGPATH env-var that the peer in cli-mode doesn't use peer but admin credentials
REMOTE_PEER_CMD="docker exec -e CORE_PEER_LOCALMSPID=Org1MSP -e CORE_PEER_MSPCONFIGPATH=/etc/hyperledger/msp/users/Admin@org1.example.com/msp peer0.org1.example.com env TERM=${TERM} ${PEER_CMD}"

# Create the channel
${REMOTE_PEER_CMD} channel create -o orderer.example.com:7050 -c ${CHANNEL_NAME} -f /etc/hyperledger/configtx/channel.tx

# Join peer0.org1.example.com to the channel.
${REMOTE_PEER_CMD}  channel join -b ${CHANNEL_NAME}.block

