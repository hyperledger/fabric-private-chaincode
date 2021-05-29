# Copyright 2021 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

PID_FILE=.worker.pid

PID="$(shell cat ${PID_FILE} 2> /dev/null)"

WORKER_DOCKER_IMAGE="irb-experimenter-worker"

run:
	@if [ -z ${PID} ]; then \
	    PYTHONPATH=../../ python3 workerCLI.py &  echo -n "$$!" > ${PID_FILE} ;\
	fi

stop:
	@if [ ! -z ${PID} ]; then \
	    wget --quiet --spider localhost:5000/shutdown -O /dev/null ;\
	    rm -rf ${PID_FILE} ;\
	fi

RUNNING_CONTAINER="$(shell docker ps --filter name=${WORKER_DOCKER_IMAGE} --quiet)"

run-docker:
	@if [ -z "${RUNNING_CONTAINER}" ]; then \
        docker run --name ${WORKER_DOCKER_IMAGE} --network host -d ${WORKER_DOCKER_IMAGE}; \
	fi

run-mac:
	@if [ -z "${RUNNING_CONTAINER}" ]; then \
        docker run --name ${WORKER_DOCKER_IMAGE} -p 5000:5000 -d ${WORKER_DOCKER_IMAGE}; \
	fi

.PHONY: stop
stop-docker:
	@if [ ! -z "${RUNNING_CONTAINER}" ]; then \
        docker stop ${WORKER_DOCKER_IMAGE} ; \
        docker rm ${WORKER_DOCKER_IMAGE} ; \
	fi

