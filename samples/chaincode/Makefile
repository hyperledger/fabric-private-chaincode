# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../..
include $(TOP)/build.mk

SUB_DIRS = auction auction-go echo kv-test kv-test-go

all build test clean clobber:
	$(foreach DIR, $(SUB_DIRS), $(MAKE) -C $(DIR) $@ || exit ;)
