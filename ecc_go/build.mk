# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

#TOP = ../
include $(TOP)/build.mk

CAAS_PORT ?= 9999

# the following are the required docker build parameters
HW_EXTENSION=$(shell if [ "${SGX_MODE}" = "HW" ]; then echo "-hw"; fi)

DOCKER_IMAGE ?= fpc/$(CC_NAME)${HW_EXTENSION}
DOCKER_FILE ?= $(FPC_PATH)/ecc_go/Dockerfile
EGO_CONFIG_FILE ?= $(FPC_PATH)/ecc_go/enclave.json
ECC_BINARY ?= ecc
ECC_BUNDLE ?= $(ECC_BINARY)-bundle

build: ecc docker env

ecc: ecc_dependencies
	ego-go build $(GOTAGS) -o $(ECC_BINARY) main.go
	cp $(EGO_CONFIG_FILE) .
	ego sign
	ego uniqueid $(ECC_BINARY) > mrenclave
	ego bundle $(ECC_BINARY) $(ECC_BUNDLE)

.PHONY: with_go
with_go: ecc_dependencies
	$(GO) build $(GOTAGS) -o $(ECC_BUNDLE) main.go
	echo "fake_mrenclave" > mrenclave

ecc_dependencies:
	# hard to list explicitly, so just leave empty target,
	# which forces ecc to always be built

env:
	echo "export CC_NAME=$(CC_NAME)" > details.env
	echo "export FPC_MRENCLAVE=$(shell cat mrenclave)" >> details.env
	echo "export FPC_CHAINCODE_IMAGE=$(DOCKER_IMAGE):latest" >> details.env

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
DOCKER_BUILD_OPTS += --build-arg FPC_CCENV_IMAGE=$(FPC_CCENV_IMAGE)
DOCKER_BUILD_OPTS += --build-arg SGX_MODE=$(SGX_MODE)
DOCKER_BUILD_OPTS += --build-arg CAAS_PORT=$(CAAS_PORT)

docker:
	$(DOCKER) build $(DOCKER_BUILD_OPTS) \
		$(shell if [ "${SGX_MODE}" = "SIM" ]; then echo "--build-arg OE_SIMULATION=1"; fi) \
		-t $(DOCKER_IMAGE):$(FPC_VERSION) \
		-f $(DOCKER_FILE) \
		. \
	&& $(DOCKER) tag $(DOCKER_IMAGE):$(FPC_VERSION) $(DOCKER_IMAGE):latest

clean:
	$(GO) clean
	rm -f $(ECC_BINARY) $(ECC_BUNDLE) enclave.json coverage.out mrenclave public.pem private.pem details.env
