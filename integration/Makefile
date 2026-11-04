# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

TOP = ..
include $(TOP)/build.mk

# the FPC-go integration tests; the C++ suites are in sgx/integration
all: test

test: ecc_go_test

ecc_go_test:
	$(MAKE) -C go_chaincode
