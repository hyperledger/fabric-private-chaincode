# Copyright 2021 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

PID_FILE=.worker.pid

PID="$(shell cat ${PID_FILE} 2> /dev/null)"

WORKER_DOCKER_IMAGE="irb-experimenter-worker"

REDIS_HOST ?= "localhost"
REDIS_PORT ?= "6379"

run:
	@if [ -z ${PID} ]; then \
	    PYTHONPATH=../../ python3 workerCLI.py &  echo -n "$$!" > ${PID_FILE} ;\
	fi

stop:
	@if [ ! -z ${PID} ]; then \
	    wget --quiet --spider localhost:5000/shutdown -O /dev/null ;\
	    rm -rf ${PID_FILE} ;\
	fi

RUNNING_CONTAINER="$(shell docker ps -a --filter name=${WORKER_DOCKER_IMAGE} --quiet)"

run-docker:
	@if [ -z "${RUNNING_CONTAINER}" ]; then \
        docker run --name ${WORKER_DOCKER_IMAGE} \
        	--env REDIS_HOST=${REDIS_HOST} --env REDIS_PORT=${REDIS_PORT} \
        	--network host \
        	-d ${WORKER_DOCKER_IMAGE}; \
	fi

mac-run-docker:
	@if [ -z "${RUNNING_CONTAINER}" ]; then \
        docker run --name ${WORKER_DOCKER_IMAGE} \
			--env REDIS_HOST=${REDIS_HOST} --env REDIS_PORT=${REDIS_PORT} \
        	-p 5000:5000 \
        	-d ${WORKER_DOCKER_IMAGE}; \
	fi

.PHONY: stop
stop-docker:
	@if [ ! -z "${RUNNING_CONTAINER}" ]; then \
        docker stop ${WORKER_DOCKER_IMAGE} ; \
        docker rm ${WORKER_DOCKER_IMAGE} ; \
	fi

