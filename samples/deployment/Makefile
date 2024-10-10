# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../..
include $(TOP)/build.mk

DEPLOYMENT_SAMPLES = test-network

build test clean clobber:
	$(foreach DIR, $(DEPLOYMENT_SAMPLES), $(MAKE) -C $(DIR) $@ || exit;)

