# Copyright 2021 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../../../..
include $(TOP)/build.mk

RUNNING_CONTAINER="$(shell docker ps --filter name=irb-storage-service --quiet)"


all: pull-redis

.PHONY: pull-redis
pull-redis:
	docker pull redis

.PHONY: run
run:
	@if [ -z "${RUNNING_CONTAINER}" ]; then \
		docker run --name irb-storage-service --network host -d redis; \
	fi

.PHONY: mac-run
mac-run:
	docker run --name irb-storage-service -p 6379:6379 -d redis; \

.PHONY: stop
stop:
	@if [ ! -z "${RUNNING_CONTAINER}" ]; then \
		docker stop irb-storage-service ; \
		docker rm irb-storage-service ; \
	fi

