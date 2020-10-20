# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ..
include $(TOP)/build.mk

build: ercc

ercc:
	# ERCC's binary is created here.
	# The binary is then referenced by the fpc-peer cli, inserted in the ERCC package,
	# and eventually run by the external launcher.
	$(GO) build $(GOTAGS) -o ercc main.go

test: build
	$(GO) test $(GOTAGS) -v ./...

clean: docker-clean
	$(GO) clean
	rm -f ercc

generate:
	$(GO) generate ./...

DOCKER_IMAGE?=fpc/ercc

docker: ercc
	$(DOCKER) build -t $(DOCKER_IMAGE) .

docker-clean:
	$(DOCKER) rmi -f $(DOCKER_IMAGE)
