# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ..
include $(TOP)/build.mk

EXAMPLES = auction echo

build clean:
	$(foreach DIR, $(EXAMPLES), $(MAKE) -C $(DIR) $@ || exit;)

