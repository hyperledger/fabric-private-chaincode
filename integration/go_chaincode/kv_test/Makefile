# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../../..
include $(TOP)/build.mk

all: test

test:
	FABRIC_LOGGING_SPEC=fpc=debug:grpc=error:comm.grpc=error:gossip=warning:info go test -v -failfast .

build:
	make -C $(FPC_PATH)/samples/chaincode/kv-test-go

docker:
	docker tag fpc/ercc:main fpc/ercc:latest
	docker tag fpc/auction:main fpc/auction:latest

clean:
	rm -rf cmd
