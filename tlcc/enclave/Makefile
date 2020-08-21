# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

.PHONY: all

all: build test

build:
	$(GO) build $(GOTAGS)

test:
	LD_LIBRARY_PATH=./lib:${LD_LIBRARY_PATH} $(GO) test $(GOTAGS) -test.v

clean:
	$(GO) clean

