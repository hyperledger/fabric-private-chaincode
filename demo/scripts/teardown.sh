#!/bin/bash
#
# Copyright IBM Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#

set -e

help(){
   echo "$(basename $0) [options]
   This script, by default, will teardown all of the auction demo components, but will not prune
   volumes or delete the chaincode images that have been created for the mockcc and the auctioncc
   options:
       --clean-slate:
           As part of the teardown, prune all unused docker volumes and delete mockcc and auctioncc
           images.
       --help,-h:
           Print this help screen.
    "
}

CLEAN_SLATE=false
for var in "$@"; do
    case "$var" in
        "--clean-slate")
            CLEAN_SLATE=true
            ;;
        "-h"|"--help")
            help
            exit
            ;;
        *)
            echo "Invalid option: ${var}"
            help
            exit
            ;;
    esac
    shift
done

DEMO_SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
DEMO_DOCKER_COMPOSE=${DEMO_SCRIPT_DIR}/../docker-compose.yml

export FPC_VERSION=${FPC_VERSION:=latest}

# SCRIPT_DIR is the docker compose script dir that needs to be defined to source environment variables from the FPC Network
SCRIPT_DIR=${DEMO_SCRIPT_DIR}/../../utils/docker-compose/scripts
. ${SCRIPT_DIR}/lib/common.sh

if $CLEAN_SLATE; then
	DOWN_OPT="--rmi all"
else
	DOWN_OPT=""
fi
COMPOSE_IGNORE_ORPHANS=true ${DOCKER_COMPOSE_CMD} -f ${DEMO_DOCKER_COMPOSE} kill && COMPOSE_IGNORE_ORPHANS=true docker-compose -f ${DEMO_DOCKER_COMPOSE} down ${DOWN_OPT}

if $CLEAN_SLATE; then
    "${SCRIPT_DIR}/teardown.sh" --clean-slate

    image=$(go run ${DEMO_SCRIPT_DIR}/../../utils/fabric/get-fabric-container-name.go --cc-name mockcc --peer-id peer0.org1.example.com --net-id dev_test --cc-version 1.0)
    docker inspect $image > /dev/null 2>&1 && docker rmi "${image}" || echo "No mockcc images available to remove"

    image=$(go run ${DEMO_SCRIPT_DIR}/../../utils/fabric/get-fabric-container-name.go --cc-name ecc_auctioncc --peer-id peer0.org1.example.com --net-id dev_test --cc-version 1.0)
    docker inspect $image > /dev/null 2>&1 && docker rmi "${image}" || echo "No auctioncc images available to remove"
fi


"${SCRIPT_DIR}/teardown.sh"
