# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../../..
include $(TOP)/build.mk

CHAINCODE_PATH=$(FPC_PATH)/samples/chaincode/kv-test-go

all: build test clean

test:
	FABRIC_LOGGING_SPEC=fpc=debug:grpc=error:comm.grpc=error:gossip=warning:info go test -v -failfast .

build:
	make -C $(CHAINCODE_PATH)
	cp $(CHAINCODE_PATH)/mrenclave .

clean:
	rm -rf cmd
	rm mrenclave
