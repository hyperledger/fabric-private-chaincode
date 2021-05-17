# Copyright 2021 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../../../../..
include $(TOP)/build.mk

WORKER_CLIENT=worker-client

all: build

$(WORKER_CLIENT):
	go build -o $@

build: $(WORKER_CLIENT)

clean:
	rm -rf $(WORKER_CLIENT)
