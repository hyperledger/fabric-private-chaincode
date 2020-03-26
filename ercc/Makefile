# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ..
include $(TOP)/build.mk

ATTESTATION = attestation/provider.go

build: attestation ercc

$(ATTESTATION):
	cd attestation && \
	make

attestation-clean:
	cd attestation && \
	make clean

ercc: $(ATTESTATION)
	$(GO) build
	rm ercc # validating a successful build, ercc not meant to be run directly

test: build
	$(GO) test -v ./...

clean: docker-clean attestation-clean
	$(GO) $@

ERCC_NAME?=ercc
DOCKER_IMAGE?=$$(docker images | grep -- -$(ERCC_NAME)- | awk '{print $$1;}')

docker-clean:
	-if [ ! -z "$(DOCKER_IMAGE)" ]; then docker rmi -f $(DOCKER_IMAGE); fi
	-if [ ! -z "$(DOCKER_IMAGE)" ]; then echo "\033[0;33mWARNING: ercc not deleted (peer still up?)\033[0m"; fi
