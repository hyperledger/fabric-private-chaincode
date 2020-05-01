# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../../../..
include $(TOP)/build.mk

CC_NAME = auction_mock_chaincode
DOCKER_IMAGE = auction_mock_backend

run-go:
	env LD_LIBRARY_PATH=../../../../ecc_enclave/_build/lib:${LD_LIBRARY_PATH} $(GO) run .

run-fpc: build
	env LD_LIBRARY_PATH=../../../../ecc_enclave/_build/lib:${LD_LIBRARY_PATH} ./mock --debug

build:
	$(GO) build -tags fpc

clean:
	rm -rf mock enclave
