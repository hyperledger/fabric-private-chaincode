# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = .
include $(TOP)/build.mk

SUB_DIRS = common utils ercc ecc_enclave ecc tlcc_enclave tlcc fabric examples integration demo # docs

FPC_SDK_DEP_DIRS = utils/fabric common ecc_enclave ecc
FPC_PEER_DEP_DIRS = common ercc tlcc_enclave tlcc fabric ecc_enclave ecc
# FPC_PEER_DEP_DIRS has to include ecc, ecc_enclave, common and ercc only if we run chaincode in external builder directly on host and not indirectly via docker
FPC_PEER_CLI_WRAPPER_DEP_DIRS = utils/fabric


.PHONY: license

build: godeps

build test clean clobber:
	$(foreach DIR, $(SUB_DIRS), $(MAKE) -C $(DIR) $@ || exit;)

checks: linter license

license:
	@echo "License: Running licence checks.."
	@scripts/check_license.sh

linter: gotools build
	@echo "LINT: Running code checks for Go files..."
	@cd $$(/bin/pwd) && ./scripts/golinter.sh
	@echo "LINT: Running code checks for Cpp/header files..."
	@cd $$(/bin/pwd) && ./scripts/cpplinter.sh
	@echo "LINT completed."

gotools:
	# install goimports if not present
	$(GO) install golang.org/x/tools/cmd/goimports

godeps: gotools
	$(GO) mod download

fpc-sdk: godeps
	$(foreach DIR, $(FPC_SDK_DEP_DIRS), $(MAKE) -C $(DIR) build || exit;)

fpc-peer: godeps
	$(foreach DIR, $(FPC_PEER_DEP_DIRS), $(MAKE) -C $(DIR) build || exit;)

fpc-peer-cli: godeps
	$(foreach DIR, $(FPC_PEER_CLI_WRAPPER_DEP_DIRS), $(MAKE) -C $(DIR) build || exit;)
