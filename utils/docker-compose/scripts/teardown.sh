#!/bin/bash
#
# Copyright IBM Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#
# Exit on first error, print all commands.
set -e

export SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. ${SCRIPT_DIR}/lib/common.sh


# Shut down the Docker containers for the system tests and remove temporary volumes
${DOCKER_COMPOSE} stop
${DOCKER_COMPOSE} kill && ${DOCKER_COMPOSE} down

if [ "$1" == '--clean-slate' ]; then
        echo "removing state of CA etc."
	${DOCKER_COMPOSE} kill && ${DOCKER_COMPOSE} down --volumes

	# CA state got destroyed, so corresponding wallets are obsolete
        if [ -d "${NODE_WALLETS}" ]; then
		echo "deleting obsolet wallets in '${NODE_WALLETS}'"
		rm -rf "${NODE_WALLETS}";
	fi

	echo "removing generated fabric configuration and credentials"
	rm -fr ${FABRIC_CFG_PATH}/config/*
	rm -fr ${FABRIC_CFG_PATH}/crypto-config/*

	# remove the local state
	rm -f ~/.hfc-key-store/*

	# remove chaincode docker images
	# (Peer should usually clean them up, but for case it has crashed ...)
	# To minimize collateral damage we explicitly exclude the dev container
 	# and restrict to containers matching the net-id from core.yaml ..
	echo "removing running containers and left-over chaincode images"
	containers=$(docker ps -a | grep -v fpc-development- | grep "${NET_ID}-" | awk '{ print $1 }')
	if [ ! -z "$containers" ]; then
		docker rm --force ${containers}
	fi
	images=$(docker images ${NET_ID}-* -q)
	if [ ! -z "${images}" ]; then
		docker rmi "${images}"
	fi
fi
