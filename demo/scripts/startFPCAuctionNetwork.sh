#!/bin/bash

#
# Copyright IBM Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#

export DEMO_SCRIPTS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
export FPC_ROOT=${DEMO_SCRIPTS_DIR}/../..
export DEMO_ROOT=${DEMO_SCRIPTS_DIR}/..

# SCRIPT_DIR is the docker compose script dir that needs to be defined to source environment variables from the FPC Network
export SCRIPT_DIR=${FPC_ROOT}/utils/docker-compose/scripts
. ${SCRIPT_DIR}/lib/common.sh

# Cleanup any previous iterations of the demo
"${SCRIPT_DIR}/teardown.sh" --clean-slate

# Generate the necessary credentials and start the FPC network
"${SCRIPT_DIR}/generate.sh"
"${SCRIPT_DIR}/start.sh"

# Install and Instantiate Auction Chaincode
"${DEMO_SCRIPTS_DIR}/installCC.sh"

# Register Users
"${DEMO_ROOT}/client/backend/fabric-gateway/registerUsers.sh"

# Run the Auction Client, the COMPOSE_PROJECT_NAME is set in ${SCRIPT_DIR}/
# Since the new containers are going to use the same network as the FPC network, docker-compose
# typically throws a warning as it sees containers using the network. To quiet the warning set
# COMPOSE_IGNORE_ORPHANS to true.
COMPOSE_IGNORE_ORPHANS=true ${DOCKER_COMPOSE_CMD} -f ${DEMO_ROOT}/docker-compose.yml up -d auction_client auction_frontend
