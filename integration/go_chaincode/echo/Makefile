# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../../..
include $(TOP)/build.mk

all: test clean

test:
	FABRIC_LOGGING_SPEC=fpc=debug:grpc=error:comm.grpc=error:gossip=warning:info go test -v -failfast .

clean:
	rm -rf cmd
