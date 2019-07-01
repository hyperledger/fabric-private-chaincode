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

DOCKER_IMAGE ?= $$(docker images | grep -- ${NET_ID}-${PEER_ID}-$(CC_NAME)- | awk '{print $$1;}')
DOCKER_CONTAINER_ID?=$$(docker ps | grep -- ${NET_ID}-${PEER_ID}-$(CC_NAME)- | awk '{print $$1;}')
DOCKER_BOILERPLATE_ECC_IMAGE ?= $(PROJECT_NAME)/boilerplate-ecc

all: build
# test target disable because it requires revision
#all: build test
build: docker-boilerplate-ecc vscc-plugin

ECC_ENCLAVE_BUILD = ../ecc_enclave/_build
ERR_MSG = "ecc_enclave build does not exist!"

ecc_enclave: sym-links
	@if [ ! -d $(ECC_ENCLAVE_BUILD)/lib ]; then echo $(ERR_MSG); exit 1; fi
	@if [ ! -d $(ECC_ENCLAVE_BUILD)/include ]; then echo $(ERR_MSG); exit 1; fi

sym-links:
	# these symbolic links are necessary now to build ecc with the generated mrenclave.go
	# once this is ready from the application build. Later, when the go file will be separated
	# from the MRENCLAVE value, the relative symbolic links can (will) be removed.
	ln -sfn ../../${ENCLAVE_SO_PATH}/../ enclave/app-lib
	[ -f enclave/app-lib/lib/mrenclave.go ] || exit 1
	ln -sfn app-lib/lib/mrenclave.go enclave/mrenclave.go
	ln -sfn ../../ecc_enclave/_build/lib enclave/ecc-enclave-lib
	ln -sfn ../../ecc_enclave/_build/include enclave/ecc-enclave-include

ecc: ecc_enclave enclave/app-lib/lib/mrenclave.go
	$(GO) get golang.org/x/sync/semaphore
	$(GO) get github.com/pkg/errors
	$(GO) get github.com/golang/protobuf/proto
	$(GO) get github.com/onsi/ginkgo
	$(GO) get github.com/onsi/gomega
	$(GO) build

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

docker-boilerplate-ecc:
	$(DOCKER) build -t $(DOCKER_BOILERPLATE_ECC_IMAGE) -f Dockerfile.boilerplate-ecc ..

docker-fpc-app: ecc
	if [ -z "$(DOCKER_IMAGE)" ]; then\
		echo "\033[0;31mERROR: cannot override $(CC_NAME) docker image - not found\033[0m";\
		exit 1;\
	fi
	echo "\033[1;33mWARNING: overriding $(DOCKER_IMAGE) docker image\033[0m"
	$(DOCKER) build -t $(DOCKER_IMAGE) -f Dockerfile.fpc-app --build-arg enclave_so_path=$(ENCLAVE_SO_PATH) ..

docker-run:
	$(DOCKER) run \
		-it \
		--device /dev/isgx \
		-v /var/run/aesmd:/var/run/aesmd \
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
