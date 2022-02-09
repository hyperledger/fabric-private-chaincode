# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../../..
include $(TOP)/build.mk

all: build test clean

test:
	FABRIC_LOGGING_SPEC=fpc=debug:grpc=error:comm.grpc=error:gossip=warning:info go test -v -failfast .

build:
	make -C $(FPC_PATH)/samples/chaincode/kv-test-go

clean:
	rm -rf cmd
