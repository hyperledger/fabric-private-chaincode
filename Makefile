# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = .
include $(TOP)/build.mk

SUB_DIRS = utils ercc ecc_enclave ecc tlcc_enclave tlcc examples integration demo # docs
PLUGINS = ercc ecc_enclave ecc tlcc_enclave tlcc
FPC_SDK = utils/fabric ecc_enclave ecc

.PHONY: license

build : godeps

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
	$(GO) get golang.org/x/tools/cmd/goimports

godeps: gotools
	$(GO) get github.com/spf13/viper
	$(GO) get golang.org/x/sync/semaphore
	$(GO) get github.com/pkg/errors
	$(GO) get github.com/golang/protobuf/proto
	$(GO) get github.com/onsi/ginkgo
	$(GO) get github.com/onsi/gomega
	$(GO) get github.com/gin-contrib/cors
	$(GO) get github.com/gin-gonic/gin
	$(GO) get github.com/dustin/go-broadcast

plugins:
	$(foreach DIR, $(PLUGINS), $(MAKE) -C $(DIR) build || exit;)

fpc-sdk: godeps
	$(foreach DIR, $(FPC_SDK), $(MAKE) -C $(DIR) build || exit;)
