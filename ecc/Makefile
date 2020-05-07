# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ..
include $(TOP)/build.mk

CC_NAME ?= ecc
NET_ID ?=dev
PEER_ID ?=jdoe

VSCC_SRC = vscc/ecc_validation_plugin.go vscc/ecc_validation_logic.go
VSCC_OUT = ecc-vscc.so

DOCKER_CONTAINER_ID?=$$(docker ps | grep -- ${NET_ID}-${PEER_ID}-$(CC_NAME)- | awk '{print $$1;}')
# the following are the required docker build parameters
DOCKER_IMAGE ?= $$(docker images | grep -- ${NET_ID}-${PEER_ID}-$(CC_NAME)- | awk '{print $$1;}')
BOILERPLATE_EXTENSION=$(shell if [ "${SGX_MODE}" = "HW" ]; then echo "-hw"; fi)
DOCKER_BOILERPLATE_ECC_IMAGE ?= hyperledger/$(PROJECT_NAME)-boilerplate-ecc$(BOILERPLATE_EXTENSION)
DOCKER_ENCLAVE_SO_PATH ?= $(ENCLAVE_SO_PATH)

all: build
# test target disable because it requires revision
#all: build test

build: vscc-plugin sym-links

ECC_ENCLAVE_BUILD = ../ecc_enclave/_build
ERR_MSG = "ecc_enclave build does not exist!"

ecc_enclave: sym-links
	@if [ ! -d $(ECC_ENCLAVE_BUILD)/lib ]; then echo $(ERR_MSG); exit 1; fi
	@if [ ! -d $(ECC_ENCLAVE_BUILD)/include ]; then echo $(ERR_MSG); exit 1; fi

sym-links:
	ln -sfn ../../ecc_enclave/_build/lib enclave/ecc-enclave-lib
	ln -sfn ../../ecc_enclave/_build/include enclave/ecc-enclave-include

ecc: ecc_enclave
	$(GO) build -o ecc cmd/main.go

vscc-plugin: $(VSCC_SRC)
	$(GO) build -o $(VSCC_OUT) -buildmode=plugin $^

#test:
#	$(LD_LIBRARY_PATH) $(GO) $@ -test.v ./...

#.PHONY: stress
#stress:
#	$(LD_LIBRARY_PATH) $(GO) test -v -run TestEnclaveChaincode_Invoke_Auction

#.PHONY: debug
#debug:
#	$(LD_LIBRARY_PATH) $(GO) test -c
#	$(LD_LIBRARY_PATH) sgx-gdb ecc.test -d $(GOROOT)

clean: docker-clean
	$(GO) clean
	rm -rf ecc enclave/ecc-enclave-lib enclave/ecc-enclave-include

# Note:
# - docker images are not necessarily rebuild if they exist but are outdated.
#   To force rebuild you have two options
#   - do a 'make clobber' first. This ensures you will have the uptodate images
#     but is a broad and slow brush
#   - to just fore rebuilding an image, call `make` with FORCE_REBUILD defined
DOCKER_BUILD_OPTS ?=
ifdef FORCE_REBUILD
	DOCKER_BUILD_OPTS += --no-cache
endif
DOCKER_BUILD_OPTS += --build-arg FPC_VERSION=$(FPC_VERSION)

docker-boilerplate-ecc: ecc
	$(DOCKER) build $(DOCKER_BUILD_OPTS) -t $(DOCKER_BOILERPLATE_ECC_IMAGE) -f Dockerfile.boilerplate-ecc\
		--build-arg SGX_MODE=$(SGX_MODE)\
		..
	$(DOCKER) tag $(DOCKER_BOILERPLATE_ECC_IMAGE) $(DOCKER_BOILERPLATE_ECC_IMAGE):${FPC_VERSION}

docker-fpc-app: docker-boilerplate-ecc
	if [ -z "$(DOCKER_IMAGE)" ]; then\
		echo "\033[0;31mERROR: cannot override $(CC_NAME) docker image - not found\033[0m";\
		exit 1;\
	fi
	echo "\033[1;33mWARNING: overriding $(DOCKER_IMAGE) docker image\033[0m"
	$(DOCKER) build $(DOCKER_BUILD_OPTS) -t $(DOCKER_IMAGE) -f Dockerfile.fpc-app \
		--build-arg BOILERPLATE_EXTENSION=$(BOILERPLATE_EXTENSION)\
		--build-arg enclave_so_path=$(DOCKER_ENCLAVE_SO_PATH)\
		..

docker-run:
	if [ "$(SGX_MODE)" = "HW" ]; then \
		SGX_DEVICE_PATH=$(if [ -e "/dev/isgx" ]; then echo "/dev/isgx"; elif [ -e "/dev/sgx" ]; then echo "/dev/sgx"; else echo "ERROR: NO SGX DEVICE FOUND"; fi);\
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
		$(DOCKER_IMAGE) \
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
