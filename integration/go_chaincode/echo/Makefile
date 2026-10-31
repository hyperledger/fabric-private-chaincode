# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../../..
include $(TOP)/build.mk

CHAINCODE_PATH=$(FPC_PATH)/samples/chaincode/echo-go

all: build test clean

# build the chaincode image this test uses (fpc/echo-go), rather than relying
# on whatever image another test left behind
build:
	make -C $(CHAINCODE_PATH)

test:
	FABRIC_LOGGING_SPEC=fpc=debug:grpc=error:comm.grpc=error:gossip=warning:info go test -v -failfast .

clean:
	rm -rf cmd
