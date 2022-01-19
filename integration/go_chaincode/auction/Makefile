# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../../..
include $(TOP)/build.mk

all: test

test:
	FABRIC_LOGGING_SPEC=fpc=debug:grpc=error:comm.grpc=error:gossip=warning:info go test -v -failfast .

docker:
	docker tag fpc/ercc:main fpc/ercc:latest
	docker tag fpc/auction:main fpc/auction:latest
	docker tag fpc/fpc-auction-go:main fpc/fpc-auction-go:latest
