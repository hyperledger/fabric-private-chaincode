# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../..
include $(TOP)/build.mk

# the C++ chaincode samples (irb's chaincode is built by samples/demos/irb)
SUB_DIRS = chaincode/auction chaincode/echo chaincode/kv-test

build test clean clobber:
	$(foreach DIR, $(SUB_DIRS), $(MAKE) -C $(DIR) $@ || exit ;)
