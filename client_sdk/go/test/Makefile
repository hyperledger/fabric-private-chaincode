# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../../..
include $(TOP)/build.mk

build:
	$(GO) build

test:
	$(GO) test $(GOTAGS) -v ./...

clean:
	$(GO) clean
	rm -rf keystore wallet

run:
	$(GO) run .