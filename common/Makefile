# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

TOP = ..
include $(TOP)/build.mk

BUILD_DIR := _build

SUB_DIRS = crypto

$(BUILD_DIR):
	@if [ ! -d $(BUILD_DIR) ]; then \
		mkdir -p $(BUILD_DIR) && \
		cd $(BUILD_DIR) && \
		cmake ./..; \
	fi

build test: $(BUILD_DIR)
	$(foreach DIR, $(SUB_DIRS), $(MAKE) -C $(DIR) $@ || exit;)
	$(MAKE) --directory=$<

clean clobber:
	$(foreach DIR, $(SUB_DIRS), $(MAKE) -C $(DIR) $@ || exit;)
	rm -rf $(BUILD_DIR)
