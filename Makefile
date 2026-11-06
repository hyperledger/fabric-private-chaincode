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
# network and runs as a separate CI job; `make suites` lists the ones CI runs
# on this side (natively, or in the dev container with FPC_CGO=1).
#
# The Go side (GO_UNIT_TEST_DIRS, GO_SUITES) needs only Go and Docker. The
# C side (CGO_UNIT_TEST_DIRS, CGO_SUITES) needs the SGX toolchain, i.e. the
# dev container: the C/C++ code, plus internal and ercc, whose fpc_cgo tests
# (PDO crypto, attestation) only run there. go_echo, go_auction and go_kv are
# in both: with FPC_CGO=1 they build the chaincode with ego, for SGX.
# go_simulation and test_network run their chaincode without SGX either way.
GO_UNIT_TEST_DIRS = internal ercc ecc client_sdk samples/application samples/chaincode utils
CGO_UNIT_TEST_DIRS = protos sgx/common internal ercc sgx/ecc_enclave sgx/samples
GO_SUITES = go_echo go_auction go_kv go_simulation test_network
CGO_SUITES = auction echo kv deployment client_sdk go_echo go_auction go_kv crash irb
ifeq ($(FPC_CGO),1)
SUITES = $(CGO_SUITES)
else
SUITES = $(GO_SUITES)
endif

.PHONY: test unit-test cgo-unit-test cpplinter suites $(addprefix suite-,$(GO_SUITES) $(CGO_SUITES))

# natively unit tests only; with FPC_CGO=1 also every suite of both sides
test:
	$(MAKE) unit-test
	$(if $(filter 1,$(FPC_CGO)),$(foreach SUITE, $(sort $(CGO_SUITES) $(GO_SUITES)), $(MAKE) suite-$(SUITE) || exit;))

# with FPC_CGO=1 both sides; CI runs cgo-unit-test in the dev container and
# unit-test natively
unit-test: $(if $(filter 1,$(FPC_CGO)),cgo-unit-test)
	$(foreach DIR, $(filter-out $(if $(filter 1,$(FPC_CGO)),$(CGO_UNIT_TEST_DIRS)),$(GO_UNIT_TEST_DIRS)), $(MAKE) -C $(DIR) test || exit;)
	cd extension/fsc && $(GO) test ./...

cgo-unit-test:
	$(foreach DIR, $(CGO_UNIT_TEST_DIRS), $(MAKE) -C $(DIR) test || exit;)

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

linter: gotools $(if $(filter 1,$(FPC_CGO)),cpplinter)
	@echo "LINT: Running code checks for Go files..."
	@./scripts/golinter.sh ${FPC_PATH}
	@echo "LINT completed."

# CI runs it on its own in the dev container; linter runs it with FPC_CGO=1
cpplinter:
	@echo "LINT: Running code checks for Cpp/header files..."
	@./scripts/cpplinter.sh ${FPC_PATH}

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
