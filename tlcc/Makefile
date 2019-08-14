# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ..
include $(TOP)/build.mk

LD_LIBRARY_PATH := LD_LIBRARY_PATH=${LD_LIBRARY_PATH}:$$(pwd)/enclave/lib
SRC := tlcc.go
OUTPUT := tlcc.so

$(OUTPUT): $(SRC)
	$(LD_LIBRARY_PATH) $(GO) build -o $@ -buildmode=plugin $<

build: $(OUTPUT)

test: build
	$(LD_LIBRARY_PATH) $(GO) $@ -test.v ./enclave

clean:
	$(GO) clean
	rm -rf ledgersData
