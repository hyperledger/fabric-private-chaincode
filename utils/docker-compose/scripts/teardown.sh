#!/bin/bash
#
# Copyright IBM Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#
# Exit on first error, print all commands.
set -e

export SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
export COMPOSE_PROJECT_NAME=fabric-fpc
export NETWORK_CONFIG=${SCRIPT_DIR}/../network-config

# Shut down the Docker containers for the system tests.
docker-compose -f $NETWORK_CONFIG/docker-compose.yml stop
docker-compose -f $NETWORK_CONFIG/docker-compose.yml kill && docker-compose -f $NETWORK_CONFIG/docker-compose.yml down

# remove the local state
rm -f ~/.hfc-key-store/*

# remove chaincode docker images
docker rm $(docker ps -aq)
docker rmi $(docker images dev-* -q)

# Your system is now clean
docker volume prune
