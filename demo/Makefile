# Copyright 2019 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

TOP = ..
include $(TOP)/build.mk

SUB_DIRS = chaincode client

build test clean:
	$(foreach DIR, $(SUB_DIRS), $(MAKE) -C $(DIR) $@ || exit;)

clobber:
	./scripts/teardown.sh --clean-slate
	$(foreach DIR, $(SUB_DIRS), $(MAKE) -C $(DIR) $@ || exit;)
