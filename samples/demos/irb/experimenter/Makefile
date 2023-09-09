# Copyright 2021 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../../../..
include $(TOP)/build.mk

all: build

protos:
	$(MAKE) -C ../protos

build: protos
	docker build -f Dockerfile -t irb-experimenter-worker ..

run: build
	$(MAKE) -C worker stop-docker run-docker

test:
	$(GO) test -v ./...
