#!/usr/bin/env bash

# Copyright IBM Corp. All Rights Reserved.
# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

if [[ -z "${FPC_PATH}" ]]; then
  echo "Error: FPC_PATH not set"
  exit 1
fi

backup() {
  FILE=$1
  BACKUP="${FILE}.backup"
  echo "backup ${FILE} ..."
  if [[ -e "${BACKUP}" ]]; then
    cp "${BACKUP}" "${FILE}"
  else
    cp "${FILE}" "${BACKUP}"
  fi
}

##############################################################################################
# Get the fabric samples repo and fetch binaries and container images
##############################################################################################

echo "Prepare fabric samples test-network for FPC"

FABRIC_SAMPLES=${FPC_PATH}/samples/deployment/test-network/fabric-samples

if [ ! -d "${FABRIC_SAMPLES}" ]; then
  echo "Fabric samples not found! cloning now!"
  git clone https://github.com/hyperledger/fabric-samples "${FABRIC_SAMPLES}"
  pushd "${FABRIC_SAMPLES}"
  git checkout -b "works" 50b69f6
  popd
fi

if [ ! -d "${FABRIC_SAMPLES}/bin" ]; then
  echo "Error: no fabric binaries found. Let's run 'network.sh prereq'"
  pushd "${FABRIC_SAMPLES}/test-network"
  ./network.sh prereq
  popd
fi


##############################################################################################
# Resolve relative paths for docker volumes
##############################################################################################

# Also there is another issue with this approach here. When working completely inside the FPC dev-container,
# the volume mounts won't work. The reason is that the docker daemon provided by the host cannot parse volume paths.
# For instance, ${FPC_PATH} inside dev-container is `/project/src/github.com/hyperledger/fabric-private-chaincode/`.
# The correct path from the docker daemon perspective is something like
# `/Users/marcusbrandenburger/Developer/gocode/src/github.com/hyperledger/fabric-private-chaincode/` (in my case on the mac).
# For this reason we use DOCKERD_FPC_PATH to resolve relative host paths.
if [ -n "${DOCKERD_FPC_PATH+x}" ]; then
  echo "Oo we are in docker mode! we need to use the host fpc path"
  FPC_PATH_HOST=${DOCKERD_FPC_PATH}
else
  FPC_PATH_HOST=${FPC_PATH}
fi
echo "set FPC_PATH_HOST = ${FPC_PATH_HOST}"

FABRIC_SAMPLES_HOST=${FPC_PATH}/samples/deployment/test-network/fabric-samples
DOCKERD_FABRIC_SAMPLES_HOST=${DOCKERD_FPC_PATH}samples/deployment/test-network/fabric-samples
TEST_NETWORK_HOST=${FABRIC_SAMPLES_HOST}/test-network
DOCKERD_TEST_NETWORK_HOST=${DOCKERD_FABRIC_SAMPLES_HOST}/test-network

echo "set DOCKERD_TEST_NETWORK_HOST = ${DOCKERD_TEST_NETWORK_HOST}"
echo "set TEST_NETWORK_HOST = ${TEST_NETWORK_HOST}"
echo "Resolving relative docker volume paths"

# replace "../" with absolute path
find "${TEST_NETWORK_HOST}/compose" -maxdepth 1 \( -name '*compose*.yaml' -o -name '*compose*.yml' \) -exec sed -i -E 's+(- )(\.\.)(/)+\1'"${DOCKERD_TEST_NETWORK_HOST}"'\3+g' {} \;
# replace "./" with absolute path
find "${TEST_NETWORK_HOST}/compose" -mindepth 2 -maxdepth 2 \( -name '*compose*.yaml' -o -name '*compose*.yml' \) -exec sed -i -E 's+(- )(\.)(/)+\1'"${DOCKERD_TEST_NETWORK_HOST}/compose"'\3+g' {} \;


##############################################################################################
# setup blockchain explorer
##############################################################################################

echo "Preparing blockchain explorer"
BE_PATH="${FPC_PATH}/samples/deployment/test-network/blockchain-explorer"
BE_PATH_HOST="${FPC_PATH_HOST}/samples/deployment/test-network/blockchain-explorer"
BE_CONFIG=${BE_PATH}/config.json
BE_CONNECTIONS_PROFILE=${BE_PATH}/connection-profile/test-network.json
BE_DOCKER_COMPOSE="${BE_PATH}/docker-compose.yaml"

if [[ -f "${BE_CONFIG}" ]] && [[ -f "${BE_CONNECTIONS_PROFILE}" ]] && [[ -f "${BE_DOCKER_COMPOSE}" ]]; then
  echo "Blockchain explorer files already exists"
  echo "Any blockchain explorer configuration are going to be overwritten."
  read -p "Are you sure? y/n " -n 1 -r
  echo ""
  if [[ $REPLY =~ ^[Nn]$ ]]; then
    echo "Skipping blockchain explorer setup"
    echo "Setup done!"
    exit 0
  fi
fi

# create folders
mkdir -p "${BE_PATH}/connection-profile"

# download configuration files
wget -O "${BE_CONFIG}" https://raw.githubusercontent.com/hyperledger/blockchain-explorer/main/examples/net1/config.json
wget -O "${BE_CONNECTIONS_PROFILE}" https://raw.githubusercontent.com/hyperledger/blockchain-explorer/main/examples/net1/connection-profile/test-network.json
wget -O "${BE_DOCKER_COMPOSE}" https://raw.githubusercontent.com/hyperledger/blockchain-explorer/main/docker-compose.yaml

cat > "${BE_PATH}/.env" << EOF
EXPLORER_CONFIG_FILE_PATH=${BE_PATH}/config.json
EXPLORER_PROFILE_DIR_PATH=${BE_PATH}/connection-profile
FABRIC_CRYPTO_PATH=${TEST_NETWORK_HOST}/organizations
EOF


##############################################################################################

echo "Setup done!"
