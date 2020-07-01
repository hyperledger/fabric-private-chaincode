# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ..
include $(TOP)/build.mk

ATTESTATION_DIR = ./attestation/
ATTESTATION_PROVIDER = $(ATTESTATION_DIR)/provider.go

build: attestation ercc

$(ATTESTATION_PROVIDER):
	$(MAKE) -C $(ATTESTATION_DIR)

attestation-clean:
	$(MAKE) -C $(ATTESTATION_DIR) clean

ercc: $(ATTESTATION_PROVIDER)
	# ercc's binary is created here and referenced by external launcher for run-on-host mode
	$(GO) build

test: build
	$(GO) test -v ./...

clean: docker-clean attestation-clean
	$(GO) $@

ERCC_NAME?=ercc
DOCKER_IMAGE?=$$(docker images | grep -- -$(ERCC_NAME)- | awk '{print $$1;}')

docker-clean:
	-if [ ! -z "$(DOCKER_IMAGE)" ]; then docker rmi -f $(DOCKER_IMAGE); fi
	-if [ ! -z "$(DOCKER_IMAGE)" ]; then echo "\033[0;33mWARNING: ercc not deleted (peer still up?)\033[0m"; fi
