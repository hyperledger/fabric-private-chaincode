#!/bin/bash
#
# Copyright IBM Corp All Rights Reserved
# Copyright Intel Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#
# Exit on first error, print all commands.
set -e

export SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. ${SCRIPT_DIR}/lib/common.sh


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
${DOCKER_COMPOSE} up -d orderer.example.com peer0.org1.example.com ca.example.com
${DOCKER_COMPOSE} ps

# wait for Hyperledger Fabric to start
# incase of errors when running later commands, issue export FABRIC_START_TIMEOUT=<larger number>
export FABRIC_START_TIMEOUT=10
sleep ${FABRIC_START_TIMEOUT}

# Create the channel
docker exec -e "CORE_PEER_LOCALMSPID=Org1MSP" -e "CORE_PEER_MSPCONFIGPATH=/etc/hyperledger/msp/users/Admin@org1.example.com/msp" peer0.org1.example.com env TERM=${TERM} ${PEER_CMD} channel create -o orderer.example.com:7050 -c ${CHANNEL_NAME} -f /etc/hyperledger/configtx/channel.tx

# Join peer0.org1.example.com to the channel.
docker exec -e "CORE_PEER_LOCALMSPID=Org1MSP" -e "CORE_PEER_MSPCONFIGPATH=/etc/hyperledger/msp/users/Admin@org1.example.com/msp" peer0.org1.example.com env TERM=${TERM} ${PEER_CMD} channel join -b ${CHANNEL_NAME}.block

