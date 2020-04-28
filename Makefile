# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = .
include $(TOP)/build.mk

# TODO bring back demo
# SUB_DIRS = utils ercc ecc_enclave ecc tlcc_enclave tlcc examples integration demo # docs
SUB_DIRS = utils ercc ecc_enclave ecc tlcc_enclave tlcc plugins fabric examples integration
FPC_SDK = utils/fabric ecc_enclave ecc

.PHONY: license

build: godeps

build test clean clobber:
	$(foreach DIR, $(SUB_DIRS), $(MAKE) -C $(DIR) $@ || exit;)

checks: linter license

license:
	@echo "License: Running licence checks.."
	@scripts/check_license.sh

linter: gotools
	@echo "LINT: Running code checks for Go files..."
	@cd $$(/bin/pwd) && ./scripts/golinter.sh
	@echo "LINT: Running code checks for Cpp/header files..."
	@cd $$(/bin/pwd) && ./scripts/cpplinter.sh
	@echo "LINT completed."

gotools:
	# install goimports if not present
	cd utils/tools && $(GO) install golang.org/x/tools/cmd/goimports

godeps: gotools
	$(GO) mod download

fpc-sdk: godeps
	$(foreach DIR, $(FPC_SDK), $(MAKE) -C $(DIR) build || exit;)
