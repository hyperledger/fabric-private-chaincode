# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

.PHONY: all

all: build test

build:
	go build

test:
	LD_LIBRARY_PATH=./lib:${LD_LIBRARY_PATH} go test -test.v

clean:
	go clean

