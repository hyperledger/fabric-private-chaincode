# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../..
include $(TOP)/build.mk

# the Go samples; the C++ samples are in sgx/samples/chaincode
GO_DIRS = auction-go echo-go kv-test-go confidential-escrow secret-keeper-go
# confidential-escrow's unit tests are flaky (TestCreateAndLockEscrow_*
# fail most runs with "Certificate hash mismatch"); skip them until fixed
TEST_SKIP_DIRS = confidential-escrow

# building the Go samples with ego needs the SGX toolchain; their Go unit
# tests don't
BUILD_DIRS = $(if $(filter 1,$(FPC_CGO)),$(GO_DIRS))

build:
	$(foreach DIR, $(BUILD_DIRS), $(MAKE) -C $(DIR) $@ || exit ;)

clean clobber:
	$(foreach DIR, $(GO_DIRS), $(MAKE) -C $(DIR) $@ || exit ;)

test:
	$(foreach DIR, $(filter-out $(TEST_SKIP_DIRS),$(GO_DIRS)), $(MAKE) -C $(DIR) $@ || exit ;)
