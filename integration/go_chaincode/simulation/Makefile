# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../../..
include $(TOP)/build.mk

ERCC_PATH=$(FPC_PATH)/ercc
CHAINCODE_PATH=$(FPC_PATH)/samples/chaincode/kv-test-go
CC_NAME=kv-test-no-sgx
# the non-SGX ercc and chaincode bundle are built into their own directory and
# image, so they never overwrite the SGX ones other suites use
NOSGX_OUT=$(CURDIR)/_nosgx
ERCC_IMAGE=fpc/ercc-nosgx

BUILD_OPT += FPC_CCENV_IMAGE=ubuntu:20.04
BUILD_OPT += GOOS=linux
BUILD_OPT += ERCC_GOTAGS=

all: build test clean

test:
	FPC_ERCC_IMAGE=$(ERCC_IMAGE) \
	FABRIC_LOGGING_SPEC=fpc=debug:grpc=error:comm.grpc=error:gossip=warning:info go test -v -failfast .

build: ercc ecc

ercc:
	$(BUILD_OPT) \
	ERCC_OUT=$(NOSGX_OUT)/ercc DOCKER_IMAGE=$(ERCC_IMAGE) \
	make -C $(ERCC_PATH) build docker

ecc:
	$(BUILD_OPT) \
	CC_NAME=$(CC_NAME) WITH_GO_OUT=$(NOSGX_OUT)/ecc DOCKER_CONTEXT=$(NOSGX_OUT)/ecc \
	make -C $(CHAINCODE_PATH) with_go docker

clean:
	rm -rf cmd $(NOSGX_OUT)
