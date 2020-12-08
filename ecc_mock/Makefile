# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ..
include $(TOP)/build.mk

build: ecc sym-links

ECC_ENCLAVE_BUILD = ../ecc_enclave/_build
ERR_MSG = "ecc_enclave build does not exist!"
DOCKER_ENCLAVE_SO_PATH ?= $(ENCLAVE_SO_PATH)

ecc_enclave: sym-links
	@if [ ! -d $(ECC_ENCLAVE_BUILD)/lib ]; then echo $(ERR_MSG); exit 1; fi
	@if [ ! -d $(ECC_ENCLAVE_BUILD)/include ]; then echo $(ERR_MSG); exit 1; fi

sym-links:
	ln -sfn ../../../ecc_enclave/_build/lib chaincode/enclave/ecc-enclave-lib
	ln -sfn ../../../ecc_enclave/_build/include chaincode/enclave/ecc-enclave-include

ecc: ecc_dependencies
	$(GO) build $(GOTAGS) -o ecc main.go

ecc_dependencies:
	# hard to list explicitly, so just leave empty target,
	# which forces ecc to always be built

test: build
	$(GO) test $(GOTAGS) -v ./...

clean:
	$(GO) clean
	$(RM) ecc enclave/ecc-enclave-lib enclave/ecc-enclave-include

docker: ecc
	$(DOCKER) build --no-cache \
		-t fpc/ecc .
