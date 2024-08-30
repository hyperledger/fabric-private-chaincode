# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../..
include $(TOP)/build.mk

DEMOS = irb

build test clean:
	$(foreach DIR, $(DEMOS), $(MAKE) -C $(DIR) $@ || exit;)

