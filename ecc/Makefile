# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ..
include $(TOP)/build.mk

CC_NAME ?= ecc
NET_ID ?=dev
PEER_ID ?=jdoe
CAAS_PORT ?= 9999

DOCKER_CONTAINER_ID?=$$(docker ps | grep -- ${NET_ID}-${PEER_ID}-$(CC_NAME)- | awk '{print $$1;}')
# the following are the required docker build parameters
DOCKER_IMAGE ?= $$(docker images | grep -- ${NET_ID}-${PEER_ID}-$(CC_NAME)- | awk '{print $$1;}')
HW_EXTENSION=$(shell if [ "${SGX_MODE}" = "HW" ]; then echo "-hw"; fi)
DOCKER_BOILERPLATE_ECC_IMAGE ?= hyperledger/$(PROJECT_NAME)-boilerplate-ecc${HW_EXTENSION}
ECC_ENCLAVE_BUILD = ../ecc_enclave/_build
ERR_MSG = "ecc_enclave build does not exist!"
DOCKER_ENCLAVE_SO_PATH ?= $(ENCLAVE_SO_PATH)

build: hw-check ecc sym-links

hw-check:
ifeq ($(SGX_MODE)$(FPC_CGO),HW)
	@echo "SGX_MODE=HW requires FPC_CGO=1 (hardware mode is only supported inside the dev container)" >&2; exit 1
endif

ecc_enclave: sym-links
	@if [ ! -d $(ECC_ENCLAVE_BUILD)/lib ]; then echo $(ERR_MSG); exit 1; fi
	@if [ ! -d $(ECC_ENCLAVE_BUILD)/include ]; then echo $(ERR_MSG); exit 1; fi

sym-links:
	ln -sfn ../../../ecc_enclave/_build/lib chaincode/enclave/ecc-enclave-lib
	ln -sfn ../../../ecc_enclave/_build/include chaincode/enclave/ecc-enclave-include

# the ecc binary is what C chaincode runs with (the external builder copies it
# into the chaincode); without fpc_cgo it would contain the mock enclave
ecc: ecc_dependencies
ifneq ($(FPC_CGO),1)
	@echo "ecc: building the ecc binary needs the C enclave; run with FPC_CGO=1 (e.g. make in-container)" >&2; exit 1
endif
	# build the package (not main.go alone): main picks the enclave from the
	# fpc_cgo-tagged enclave_cgo.go or enclave_mock.go
	$(GO) build $(GOTAGS) -o ecc .

ecc_dependencies:
	# hard to list explicitly, so just leave empty target,
	# which forces ecc to always be built

# unit tests always run with the mock enclave, i.e., without fpc_cgo; they
# don't need the ecc binary, so `test` doesn't depend on `build`
ECC_TEST_TAG_LIST = $(filter-out fpc_cgo,$(GO_TAG_LIST))
test:
	$(GO) test $(if $(strip $(ECC_TEST_TAG_LIST)),-tags $(subst $(space),$(comma),$(strip $(ECC_TEST_TAG_LIST)))) $(GOTESTFLAGS) ./...

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

docker: docker-fpc-app

docker-boilerplate-ecc: ecc
	$(DOCKER) build $(DOCKER_BUILD_OPTS) -t $(DOCKER_BOILERPLATE_ECC_IMAGE):$(FPC_VERSION) -f Dockerfile.boilerplate-ecc\
		--build-arg SGX_MODE=$(SGX_MODE)\
		--build-arg CAAS_PORT=$(CAAS_PORT)\
		..

docker-fpc-app: docker-boilerplate-ecc
	if [ -z "$(DOCKER_IMAGE)" ]; then\
		echo "\033[0;31mERROR: cannot override $(CC_NAME) docker image - not found\033[0m";\
		exit 1;\
	fi
	echo "\033[1;33mWARNING: overriding $(DOCKER_IMAGE) docker image\033[0m"
	$(DOCKER) build $(DOCKER_BUILD_OPTS) -t $(DOCKER_IMAGE):$(FPC_VERSION) -t $(DOCKER_IMAGE):latest -f Dockerfile.fpc-app \
                --build-arg HW_EXTENSION=$(HW_EXTENSION)\
 		$(DOCKER_ENCLAVE_SO_PATH)

docker-run:
	if [ "$(SGX_MODE)" = "HW" ]; then \
		SGX_DEVICE_PATH=$(if [ -e "/dev/isgx" ]; then echo "/dev/isgx"; elif [ -e "/dev/sgx/enclave" ]; then echo "/dev/sgx/enclave"; else echo "ERROR: NO SGX DEVICE FOUND"; fi);\
		DOCKER_SGX_ARGS="--device $${SGX_DEVICE_PATH} -v /var/run/aesmd:/var/run/aesmd";\
	fi;\
	$(DOCKER) run \
		-it \
		$${DOCKER_SGX_ARGS} \
		--name $(CC_NAME) \
		-e "CORE_CHAINCODE_LOGGING_LEVEL=DEBUG" \
		-e "CORE_CHAINCODE_LOGGING_SHIM=INFO" \
		-e "CORE_PEER_ADDRESS=172.17.0.1:7051" \
		-e "CORE_CHAINCODE_ID_NAME=$(CC_NAME):0" \
		$(DOCKER_IMAGE):$(FPC_VERSION) \
		chaincode

docker-stop:
	-if [ ! -z "$(DOCKER_CONTAINER_ID)" ]; then docker stop $(DOCKER_CONTAINER_ID); fi
	-if [ ! -z "$(DOCKER_CONTAINER_ID)" ]; then docker rm $(DOCKER_CONTAINER_ID); fi

docker-clean: docker-stop
	-if [ ! -z "$(DOCKER_IMAGE)" ]; then docker rmi -f $(DOCKER_IMAGE); fi
	for tag in \
	    ${FPC_VERSION} \
	    latest \
	; do \
	    if [ ! -z "$$(docker images -q ${DOCKER_BOILERPLATE_ECC_IMAGE}:$${tag})" ]; then \
		docker rmi $(DOCKER_BOILERPLATE_ECC_IMAGE):$${tag}; \
	    fi \
	done

clobber: docker-clean
clean: docker-clean
	$(GO) clean
	$(RM) ecc enclave/ecc-enclave-lib enclave/ecc-enclave-include
