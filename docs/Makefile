# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ..
include $(TOP)/build.mk

SUB_DIRS = design

build test clean clobber:
	$(foreach DIR, $(SUB_DIRS), $(MAKE) -C $(DIR) $@ || exit;)
