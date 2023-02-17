# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = .
include $(TOP)/build.mk

SUB_DIRS = protos common internal ercc ecc_enclave ecc fabric client_sdk samples utils integration # docs

FPC_SDK_DEP_DIRS = protos common utils/fabric ecc_enclave ecc
FPC_PEER_DEP_DIRS = protos common ercc fabric ecc_enclave ecc
# FPC_PEER_DEP_DIRS has to include protos, ecc, ecc_enclave, common and ercc only if we run chaincode in external builder directly on host and not indirectly via docker
FPC_PEER_CLI_WRAPPER_DEP_DIRS = utils/fabric


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
	@./scripts/golinter.sh ${FPC_PATH}
	@echo "LINT: Running code checks for Cpp/header files..."
	@./scripts/cpplinter.sh ${FPC_PATH}
	@echo "LINT completed."

gotools:
	# install go tools if not present
	# (for faster docker-build, also replicte these commands
	#  in 'utils/docker/base-dev/Dockerfile')
	$(GO) install golang.org/x/tools/cmd/goimports
	$(GO) install google.golang.org/protobuf/cmd/protoc-gen-go
	GO111MODULE=off $(GO) get github.com/maxbrunsfeld/counterfeiter
	$(GO) install honnef.co/go/tools/cmd/staticcheck@v0.3.3
	$(GO) get -u github.com/client9/misspell/cmd/misspell

godeps: gotools
	$(GO) mod download

fpc-sdk: godeps
	$(foreach DIR, $(FPC_SDK_DEP_DIRS), $(MAKE) -C $(DIR) build || exit;)

fpc-peer: godeps
	$(foreach DIR, $(FPC_PEER_DEP_DIRS), $(MAKE) -C $(DIR) build || exit;)

fpc-peer-cli: godeps
	$(foreach DIR, $(FPC_PEER_CLI_WRAPPER_DEP_DIRS), $(MAKE) -C $(DIR) build || exit;)

report:
	@echo "Reporting CI data..."
	@cd $$(/bin/pwd) && ./scripts/report.sh

docker:
	$(MAKE) -C utils/docker

# add the ci_report target only at CI time, when coverage is enabled
ifeq (${IS_CI_RUNNING}, true)
ifeq (${CODE_COVERAGE_ENABLED}, true)
ci_report: report
endif
endif
