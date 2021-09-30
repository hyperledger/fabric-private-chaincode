# Copyright 2021 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../../../..
include $(TOP)/build.mk

all: build

build:
	DOCKER_BUILDKIT=0 docker build -f Dockerfile -t irb-experimenter-worker ..

run: build
	make -C worker stop-docker run-docker

test:
	go run test/worker-client.go
