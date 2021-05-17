# Copyright 2021 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../../../..
include $(TOP)/build.mk

all: pull-redis

.PHONY: pull-redis
pull-redis:
	docker pull redis

.PHONY: run
run:
	docker run --name irb-storage-service --network host -d redis

.PHONY: stop
stop:
	docker stop irb-storage-service
	docker rm irb-storage-service
