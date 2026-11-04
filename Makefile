# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = .
include $(TOP)/build.mk

# The Go side builds anywhere Go runs. The C side (C/C++ code, enclaves, the
# ecc binary, the shell integration suites) needs the SGX toolchain and is
# built only with FPC_CGO=1, which the dev container sets.
ifeq ($(FPC_CGO),1)
# the C/C++ code lives in sgx/; its parts are listed where they were built
# before the move, as the order matters (protos before sgx/common, the enclave
# before ecc, ecc before the C++ samples' images)
SUB_DIRS = protos sgx/common internal ercc sgx/ecc_enclave ecc fabric client_sdk samples sgx/samples utils integration sgx/integration # docs
else
# fabric fetches the Fabric release binaries the FSC test networks run on
SUB_DIRS = internal ercc fabric client_sdk samples utils
endif

FPC_SDK_DEP_DIRS = protos sgx/common utils/fabric sgx/ecc_enclave ecc
FPC_PEER_DEP_DIRS = protos sgx/common ercc fabric sgx/ecc_enclave ecc
# FPC_PEER_DEP_DIRS has to include protos, ecc, ecc_enclave, common and ercc only if we run chaincode in external builder directly on host and not indirectly via docker
FPC_PEER_CLI_WRAPPER_DEP_DIRS = utils/fabric


.PHONY: license

build: godeps

build clean clobber:
	$(foreach DIR, $(SUB_DIRS), $(MAKE) -C $(DIR) $@ || exit;)

# Tests are split into unit tests and suites. Each suite starts its own
# network and runs as a separate CI job; `make suites` lists them for the CI
# matrix. `make test` runs everything, sequentially, as before.
ifeq ($(FPC_CGO),1)
UNIT_TEST_DIRS = protos sgx/common internal ercc sgx/ecc_enclave ecc fabric client_sdk samples/application samples/chaincode sgx/samples utils
SUITES = auction echo kv deployment client_sdk go_echo go_auction go_kv go_simulation crash irb test_network
else
# the ecc unit tests use the mock enclave, so they run without the C side
UNIT_TEST_DIRS = internal ercc ecc client_sdk samples/application samples/chaincode utils
SUITES =
endif

.PHONY: test unit-test suites $(addprefix suite-,$(SUITES))

test:
	$(MAKE) unit-test
	$(foreach SUITE, $(SUITES), $(MAKE) suite-$(SUITE) || exit;)

unit-test:
	$(foreach DIR, $(UNIT_TEST_DIRS), $(MAKE) -C $(DIR) test || exit;)
	cd extension/fsc && $(GO) test ./...

suites:
	@echo $(SUITES)

suite-auction:
	$(MAKE) -C sgx/integration auction_test
suite-echo:
	$(MAKE) -C sgx/integration echo_test
suite-kv:
	$(MAKE) -C sgx/integration kv_test
suite-deployment:
	$(MAKE) -C sgx/integration deployment_test
suite-client_sdk:
	$(MAKE) -C sgx/integration client_sdk_test
suite-crash:
	$(MAKE) -C sgx/integration crash_test
suite-go_echo:
	$(MAKE) -C integration/go_chaincode deps
	$(MAKE) -C integration/go_chaincode/echo
suite-go_auction:
	$(MAKE) -C integration/go_chaincode deps
	$(MAKE) -C integration/go_chaincode/auction
suite-go_kv:
	$(MAKE) -C integration/go_chaincode deps
	$(MAKE) -C integration/go_chaincode/kv_test
suite-go_simulation:
	$(MAKE) -C integration/go_chaincode deps
	$(MAKE) -C integration/go_chaincode/simulation
suite-irb:
	$(MAKE) -C samples/demos/irb build test
suite-test_network:
	$(MAKE) -C samples/deployment test

checks: linter license

license:
	@echo "License: Running licence checks.."
	@scripts/check_license.sh

linter: gotools
	@echo "LINT: Running code checks for Go files..."
	@./scripts/golinter.sh ${FPC_PATH}
ifeq ($(FPC_CGO),1)
	@echo "LINT: Running code checks for Cpp/header files..."
	@./scripts/cpplinter.sh ${FPC_PATH}
endif
	@echo "LINT completed."

gotools:
	# install go tools if not present
	# (for faster docker-build, also replicte these commands
	#  in 'utils/docker/base-dev/Dockerfile')
	$(GO) install golang.org/x/tools/cmd/goimports
	$(GO) install google.golang.org/protobuf/cmd/protoc-gen-go
	$(GO) install honnef.co/go/tools/cmd/staticcheck@2023.1.3
	$(GO) install github.com/client9/misspell/cmd/misspell

godeps: gotools
	$(GO) mod download

gotidy:
	./scripts/gomate.sh tidy

fpc-sdk: godeps
	$(foreach DIR, $(FPC_SDK_DEP_DIRS), $(MAKE) -C $(DIR) build || exit;)

fpc-peer: godeps
	$(foreach DIR, $(FPC_PEER_DEP_DIRS), $(MAKE) -C $(DIR) build || exit;)

fpc-peer-cli: godeps
	$(foreach DIR, $(FPC_PEER_CLI_WRAPPER_DEP_DIRS), $(MAKE) -C $(DIR) build || exit;)

docker:
	$(MAKE) -C utils/docker

# runs a make target (default: all) in the dev container, which has the SGX
# toolchain and sets FPC_CGO=1, e.g. `make in-container TARGET=suite-echo`
in-container:
	$(MAKE) -C utils/docker run-dev DOCKER_DEV_OPTIONAL_CMD='make $(or $(TARGET),all)'
