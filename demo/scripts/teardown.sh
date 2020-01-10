#!/bin/bash
#
# Copyright IBM Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#

set -e

DEMO_SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
DEMO_DOCKER_COMPOSE=${DEMO_SCRIPT_DIR}/../docker-compose.yml

# SCRIPT_DIR is the docker compose script dir that needs to be defined to source environment variables from the FPC Network
SCRIPT_DIR=${DEMO_SCRIPT_DIR}/../../utils/docker-compose/scripts
. ${SCRIPT_DIR}/lib/common.sh

COMPOSE_IGNORE_ORPHANS=true docker-compose -f ${DEMO_DOCKER_COMPOSE} kill && COMPOSE_IGNORE_ORPHANS=true docker-compose -f ${DEMO_DOCKER_COMPOSE} down


if [ "$1" == '--clean-slate' ]; then
    "${SCRIPT_DIR}/teardown.sh" --clean-slate
    images=$(docker images dev_*mockcc-* -q)
    if [ ! -z "${images}" ]; then
        docker rmi "${images}"
    fi
else
    "${SCRIPT_DIR}/teardown.sh"
fi
