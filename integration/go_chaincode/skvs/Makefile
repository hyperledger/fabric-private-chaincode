# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../../..
include $(TOP)/build.mk

CHAINCODE_PATH=$(FPC_PATH)/samples/chaincode/secret-keeper-go

all: build test clean

test:
	FABRIC_LOGGING_SPEC=fpc=debug:grpc=error:comm.grpc=error:gossip=warning:info go test -v -failfast .

build:
	make -C $(CHAINCODE_PATH) ECC_MAIN_FILES=$(CHAINCODE_PATH)/cmd/skvs/main.go with_go env docker
	cp $(CHAINCODE_PATH)/mrenclave .

clean:
	rm -rf cmd
	rm mrenclave
