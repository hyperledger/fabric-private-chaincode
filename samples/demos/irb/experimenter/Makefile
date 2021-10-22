# Copyright 2021 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../../../..
include $(TOP)/build.mk

all: docker

docker:
	DOCKER_BUILDKIT=0 docker build -f Dockerfile -t irb-experimenter-worker ..

run: docker
	$(MAKE) -C worker stop-docker run-docker

test:
	$(GO) test -v ./...
