# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../../..
include $(TOP)/build.mk

COMPONENTS = chaincode experimenter

all: build test

build:
	$(foreach DIR, $(COMPONENTS), $(MAKE) -C $(DIR) $@ || exit;)

test:
	$(GO) test -v ./...

clean:
	$(GO) clean
	rm -rf cmd users
	$(foreach DIR, $(COMPONENTS), $(MAKE) -C $(DIR) $@ || exit;)
