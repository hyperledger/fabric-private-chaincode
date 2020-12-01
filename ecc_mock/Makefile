# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ..
include $(TOP)/build.mk

build: ecc

ecc: ecc_dependencies
	$(GO) build $(GOTAGS) -o ecc main.go

ecc_dependencies:
	# hard to list explicitly, so just leave empty target,
	# which forces ecc to always be built

test: build
	$(GO) test $(GOTAGS) -v ./...

clean:
	$(GO) clean
	$(RM) ecc

docker: ecc
	${DOCKER} build --no-cache -t fpc/ecc .
