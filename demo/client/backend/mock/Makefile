# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../../../..
include $(TOP)/build.mk

CC_NAME = auction_mock_chaincode
DOCKER_IMAGE = auction_mock_backend
APP_SRC = ${PWD}

run-fpc: build
	env LD_LIBRARY_PATH=../../../../ecc_enclave/_build/lib:${LD_LIBRARY_PATH} ./mock --debug

build:
	$(GO) build

clean:
	rm -rf mock enclave
