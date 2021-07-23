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

echo "Prepare fabric samples test-network for FPC"

FABRIC_SAMPLES=${FPC_PATH}/samples/deployment/test-network/fabric-samples
CORE_PATH=${FABRIC_SAMPLES}/config/core.yaml
DOCKER_PATH=${FABRIC_SAMPLES}/test-network/docker
DOCKER_COMPOSE_TEST_NET=${DOCKER_PATH}/docker-compose-test-net.yaml
DOCKER_COMPOSE_CA=${DOCKER_PATH}/docker-compose-ca.yaml

if [ ! -d "${FABRIC_SAMPLES}/bin" ]; then
  echo "Error: no fabric binaries found, see README.md"
  exit 1
fi


###############################################
# Adding FPC support to core.yaml
###############################################

echo "Adding FPC external builder to core.yaml"
# Create a backup copy of `core.yaml` first and always work from there.
backup ${CORE_PATH}
yq m -i -a=append ${CORE_PATH} core_ext.yaml


###############################################
# Resolve relative paths for docker volumes
###############################################

# Also there is another issue with this approach here. When working completely inside the FPC dev-container,
# the volume mounts won't work. The reason is that the docker daemon provided by the host cannot parse volume paths.
# For instance, ${FPC_PATH} inside dev-container is `/project/src/github.com/hyperledger/fabric-private-chaincode/`.
# The correct path from the docker daemon perspective is something like
# `/Users/marcusbrandenburger/Developer/gocode/src/github.com/hyperledger/fabric-private-chaincode/` (in my case on the mac).
# For this reason we use DOCKERD_FPC_PATH to resolve relative host paths.
if [ ! -z ${DOCKERD_FPC_PATH+x} ]; then
  echo "Oo we are in docker mode! we need to use the host fpc path"
  FPC_PATH_HOST=${DOCKERD_FPC_PATH}
else
  FPC_PATH_HOST=${FPC_PATH}
fi
echo "set FPC_PATH_HOST = ${FPC_PATH_HOST}"

FABRIC_SAMPLES_HOST=${FPC_PATH_HOST}/samples/deployment/test-network/fabric-samples

echo "Resolving relative docker volume paths in ..."

echo "${DOCKER_COMPOSE_TEST_NET}"
backup ${DOCKER_COMPOSE_TEST_NET}
sed -i "s+\.\./+${FABRIC_SAMPLES_HOST}/test-network/+g" "${DOCKER_COMPOSE_TEST_NET}"

echo "${DOCKER_COMPOSE_CA}"
backup ${DOCKER_COMPOSE_CA}
sed -i "s+\.\./+${FABRIC_SAMPLES_HOST}/test-network/+g" "${DOCKER_COMPOSE_CA}"

echo "${DOCKER_COMPOSE_TEST_NET}"
peers=("peer0.org1.example.com" "peer0.org2.example.com")
for p in "${peers[@]}"; do
  yq w -i ${DOCKER_COMPOSE_TEST_NET} "services.\"$p\".volumes[+]" "${FPC_PATH_HOST}:/opt/gopath/src/github.com/hyperledger/fabric-private-chaincode"
  yq w -i ${DOCKER_COMPOSE_TEST_NET} "services.\"$p\".volumes[+]" "${FABRIC_SAMPLES_HOST}/config/core.yaml:/etc/hyperledger/fabric/core.yaml"
done

###############################################
# setup blockchain explorer
###############################################

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
mkdir -p ${BE_PATH}/connection-profile

# download configuration files
wget -O ${BE_CONFIG} https://raw.githubusercontent.com/hyperledger/blockchain-explorer/main/examples/net1/config.json
wget -O ${BE_CONNECTIONS_PROFILE} https://raw.githubusercontent.com/hyperledger/blockchain-explorer/main/examples/net1/connection-profile/test-network.json
wget -O ${BE_DOCKER_COMPOSE} https://raw.githubusercontent.com/hyperledger/blockchain-explorer/main/docker-compose.yaml

# prepare BE docker compose file to be used in FPC docker dev environment and with localhost
echo "Resolving relative volume paths in ..."
echo "${BE_DOCKER_COMPOSE}"
sed -i "s+./examples/net1+${BE_PATH_HOST}+g" "${BE_DOCKER_COMPOSE}"
sed -i "s+/fabric-path/fabric-samples+${FABRIC_SAMPLES_HOST}+g" "${BE_DOCKER_COMPOSE}"

###############################################

echo "Setup done!"
