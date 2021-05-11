# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../../../..
include $(TOP)/build.mk

BUILD_DIR := _build

$(BUILD_DIR):
	@if [ ! -d $(BUILD_DIR) ]; then \
		mkdir -p $(BUILD_DIR) && \
		cd $(BUILD_DIR) && \
		cmake ./..; \
	fi

build: protos $(BUILD_DIR)
	$(MAKE) --directory=$(BUILD_DIR)

clean: protos-clean
	rm -rf $(BUILD_DIR)


PROTO_BUILD_DIR := _protos
PROTO_PATH := ../protos
PROTO_FILES := irb.proto

.PHONY: protos
protos: $(PROTO_BUILD_DIR)
	$(PROTOC_CMD) \
		--plugin=protoc-gen-nanopb=$(NANOPB_PATH)/generator/protoc-gen-nanopb-py2 \
		--proto_path=$(PROTO_PATH) --nanopb_out=$(PROTO_BUILD_DIR) $(PROTO_FILES)

$(PROTO_BUILD_DIR):
	mkdir -p $@

.PHONY: protos-clean
protos-clean:
	rm -rf $(PROTO_BUILD_DIR)
