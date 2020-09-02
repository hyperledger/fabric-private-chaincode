# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ..
include $(TOP)/build.mk

PROTO_C_BUILD_DIR := $(TOP)/common/protos
PROTO_GO_BUILD_DIR := $(TOP)/internal/protos

.PHONY: protos
protos: $(PROTO_C_BUILD_DIR) $(PROTO_GO_BUILD_DIR)
	./generate_protos.sh

$(PROTO_GO_BUILD_DIR):
	mkdir -p $@

$(PROTO_C_BUILD_DIR):
	mkdir -p $@

.PHONY: protos-clean
protos-clean:
	rm -rf $(PROTO_C_BUILD_DIR)
	rm -rf $(PROTO_GO_BUILD_DIR)

build: protos

clean: protos-clean
