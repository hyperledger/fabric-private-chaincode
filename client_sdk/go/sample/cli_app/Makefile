# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../../../..
include $(TOP)/build.mk

build:
	$(GO) build $(GOTAGS) -o fpcclient

docker:
	$(DOCKER) build $(DOCKER_BUILD_OPTS) \
    	--no-cache \
    	-t fpc/fpcclient:latest \
    	-f Dockerfile \
    	.
