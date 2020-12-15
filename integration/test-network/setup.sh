#!/bin/bash

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

  if [[ -e "${BACKUP}" ]]; then
    cp "${BACKUP}" "${FILE}"
  else
    cp "${FILE}" "${BACKUP}"
  fi
}

FABRIC_SAMPLES=${FPC_PATH}/integration/test-network/fabric-samples/

if [ ! -d "${FABRIC_SAMPLES}/bin" ]; then
  echo "Error: environment not properly setup, see README.md"
  #cd ${FABRIC_SAMPLES} && curl -sSL https://bit.ly/2ysbOFE | bash -s -- -s
fi

# patch fabric-sample

# - config and docker-compose files (for fpc lite enablement)

CORE_PATH=${FABRIC_SAMPLES}/config/core.yaml
DOCKER_PATH=${FABRIC_SAMPLES}/test-network/docker/docker-compose-test-net.yaml

# TODO the current `setup.sh` has an issue that it is not idempotent
# In particular, when starting nodes using the ./network up ... script won't work if
# docker-compose-test-net.yaml contains redundant volume entries!
# maybe create a backup copy of `core.yaml` and `docker-compose-test-net.yaml` first and always work from there.
backup ${CORE_PATH}

yq m -i -a=append ${CORE_PATH} core_ext.yaml

backup ${DOCKER_PATH}

peers=("peer0.org1.example.com" "peer0.org2.example.com")

# Also there is another issue with this approach here. When working completely inside the FPC dev-container,
# the volume mounts won't work. The reason is that the docker daemon provided by the host cannot parse volume paths.
# For instance, ${FPC_PATH} inside dev-container is `/project/src/github.com/hyperledger-labs/fabric-private-chaincode/`.
# The correct path from the docker daemon perspective is something like
# `/Users/marcusbrandenburger/Developer/gocode/src/github.com/hyperledger-labs/fabric-private-chaincode/` (in my case on the mac).

# Using DOCKERED_FPC_PATH fixes our custom volumes but still there is an issue default mounts in `docker-compose-test-net.yaml`
# TODO need to be fixed

if [ ! -z ${DOCKERD_FPC_PATH+x} ]; then
  echo "Oo we are in docker mode! we need to use the host fpc path"
  FPC_PATH=${DOCKERD_FPC_PATH}
  FABRIC_SAMPLES=${FPC_PATH}/integration/test-network/fabric-samples/
  echo "set FPC_PATH = ${FPC_PATH}"
fi

for p in "${peers[@]}"; do
  yq w -i ${DOCKER_PATH} "services.\"$p\".volumes[+]" "${FPC_PATH}:/opt/gopath/src/github.com/hyperledger-labs/fabric-private-chaincode"
  yq w -i ${DOCKER_PATH} "services.\"$p\".volumes[+]" "${FABRIC_SAMPLES}/config/core.yaml:/etc/hyperledger/fabric/core.yaml"
done
