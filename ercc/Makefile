# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ..
include $(TOP)/build.mk

VSCC_SRC = vscc/ercc_validation_plugin.go vscc/ercc_validation_logic.go
VSCC_OUT = ercc-vscc.so
DECO_SRC = attestation/ias_credentials/decoration.go
DECO_OUT = ercc-decorator.so

SUBDIRS = attestation

.PHONY: $(SUBDIRS)

$(SUBDIRS):
	$(MAKE) -C $(SUBDIRS)

$(SUBDIRS)-clean:
	$(MAKE) -C $(SUBDIRS) clean

build: $(SUBDIRS) ercc vscc-plugin decorator-plugin

ercc:
	$(GO) build

vscc-plugin: $(VSCC_SRC)
	$(GO) build -o $(VSCC_OUT) -buildmode=plugin $^

decorator-plugin: $(DECO_SRC)
	$(GO) build -o $(DECO_OUT) -buildmode=plugin $^

test: build
	$(GO) test -v

clean: docker-clean $(SUBDIRS)-clean
	$(GO) $@

ERCC_NAME?=ercc
DOCKER_IMAGE?=$$(docker images | grep -- -$(ERCC_NAME)- | awk '{print $$1;}')

docker-clean:
	-if [ ! -z "$(DOCKER_IMAGE)" ]; then docker rmi -f $(DOCKER_IMAGE); fi
	-if [ ! -z "$(DOCKER_IMAGE)" ]; then echo "\033[0;33mWARNING: ercc not deleted (peer still up?)\033[0m"; fi
