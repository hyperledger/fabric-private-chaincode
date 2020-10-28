# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ..
include $(TOP)/build.mk

build: ecc

ecc: main.go
	$(GO) build $(GOTAGS) -o ecc main.go

test: build
	$(GO) test $(GOTAGS) -v ./...

clean:
	$(GO) clean
	$(RM) ecc

docker:
	${DOCKER} build --no-cache -t fpc/ecc .
