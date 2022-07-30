# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../../..
include $(TOP)/build.mk

ERCC_PATH=$(FPC_PATH)/ercc
CHAINCODE_PATH=$(FPC_PATH)/samples/chaincode/kv-test-go
CC_NAME=kv-test-no-sgx

BUILD_OPT += FPC_CCENV_IMAGE=ubuntu:20.04
BUILD_OPT += GOOS=linux
BUILD_OPT += ERCC_GOTAGS=

all: build test clean

test:
	FABRIC_LOGGING_SPEC=fpc=debug:grpc=error:comm.grpc=error:gossip=warning:info go test -v -failfast .

build: ercc ecc

ercc:
	$(BUILD_OPT) \
	make -C $(ERCC_PATH) build docker

ecc:
	$(BUILD_OPT) \
	CC_NAME=$(CC_NAME) \
	make -C $(CHAINCODE_PATH) with_go docker

clean:
	rm -rf cmd
