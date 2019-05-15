# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ..
include $(TOP)/build.mk

VSCC_SRC = vscc/ercc_validation_plugin.go vscc/ercc_validation_logic.go
VSCC_OUT = ercc-vscc.so
DECO_SRC = attestation/ias_credentials/decoration.go
DECO_OUT = ercc-decorator.so

build: ercc vscc-plugin decorator-plugin

ercc:
	$(GO) build

vscc-plugin: $(VSCC_SRC)
	$(GO) build -o $(VSCC_OUT) -buildmode=plugin $^

decorator-plugin: $(DECO_SRC)
	$(GO) build -o $(DECO_OUT) -buildmode=plugin $^

test: build
	$(GO) test -v

clean: docker-clean
	$(GO) $@

ERCC_NAME?=ercc
DOCKER_IMAGE?=$$(docker images | grep -- -$(ERCC_NAME)- | awk '{print $$1;}')

docker-clean:
	-docker rmi $(DOCKER_IMAGE)
