# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

#TOP = ../
include $(TOP)/build.mk

CAAS_PORT ?= 9999

# the following are the required docker build parameters
HW_EXTENSION=$(shell if [ "${SGX_MODE}" = "HW" ]; then echo "-hw"; fi)

DOCKER_IMAGE ?= fpc/fpc-$(CC_NAME)-go${HW_EXTENSION}
DOCKER_FILE ?= $(FPC_PATH)/ecc_go/Dockerfile
EGO_CONFIG_FILE ?= $(FPC_PATH)/ecc_go/enclave.json

build: ecc docker

ecc: ecc_dependencies
	ego-go build $(GOTAGS) -o ecc main.go
	ego sign $(EGO_CONFIG_FILE)
	ego uniqueid ecc > mrenclave

.PHONY: with_go
with_go: ecc_dependencies
	$(GO) build $(GOTAGS) -o ecc main.go
	echo "fake_mrenclave" > mrenclave

ecc_dependencies:
	# hard to list explicitly, so just leave empty target,
	# which forces ecc to always be built

test: build
	# note that we run unit test with a mock enclave
	$(GO) test $(GOTAGS) $(GOTESTFLAGS) ./...

# Note:
# - docker images are not necessarily rebuild if they exist but are outdated.
#   To force rebuild you have two options
#   - do a 'make clobber' first. This ensures you will have the uptodate images
#     but is a broad and slow brush
#   - to just fore rebuilding an image, call `make` with DOCKER_FORCE_REBUILD defined
#   - to keep docker build quiet unless there is an error, call `make` with DOCKER_QUIET_BUILD defined
DOCKER_BUILD_OPTS ?=
ifdef DOCKER_QUIET_BUILD
	DOCKER_BUILD_OPTS += --quiet
endif
ifdef DOCKER_FORCE_REBUILD
	DOCKER_BUILD_OPTS += --no-cache
endif
DOCKER_BUILD_OPTS += --build-arg FPC_VERSION=$(FPC_VERSION)
DOCKER_BUILD_OPTS += --build-arg SGX_MODE=$(SGX_MODE)
DOCKER_BUILD_OPTS += --build-arg CAAS_PORT=$(CAAS_PORT)



docker:
	$(DOCKER) build $(DOCKER_BUILD_OPTS) -t $(DOCKER_IMAGE):$(FPC_VERSION) -f $(DOCKER_FILE)\
		$(shell if [ "${SGX_MODE}" = "SIM" ]; then echo "--build-arg OE_SIMULATION=1"; fi)\
		. &&\
	$(DOCKER) tag $(DOCKER_IMAGE):$(FPC_VERSION) $(DOCKER_IMAGE):latest

clean:
	$(GO) clean
	rm -f ecc coverage.out
