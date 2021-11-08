# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ..
include $(TOP)/build.mk

SAMPLES = application chaincode demos #deployment

build clean:
	$(foreach DIR, $(SAMPLES), $(MAKE) -C $(DIR) $@ || exit;)

