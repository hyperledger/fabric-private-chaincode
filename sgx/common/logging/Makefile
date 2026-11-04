# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../..
include $(TOP)/build.mk

BUILD_DIR := _build

$(BUILD_DIR):
	@if [ ! -d $(BUILD_DIR) ]; then \
		mkdir -p $(BUILD_DIR) && \
		cd $(BUILD_DIR) && \
		cmake ./.. ; \
	fi

build: $(BUILD_DIR)
	$(MAKE) --directory=$<

test: build
	$(MAKE) -C $(BUILD_DIR) test

clean:
	rm -rf $(BUILD_DIR)
