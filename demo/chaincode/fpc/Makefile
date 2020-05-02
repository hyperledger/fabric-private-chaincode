# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

#ifneq ($(FPC_PATH), '')
TOP = ../../../

include $(TOP)/build.mk

BUILD_DIR := _build

$(BUILD_DIR):
	@if [ ! -d $(BUILD_DIR) ]; then \
		mkdir -p $(BUILD_DIR) && \
		cd $(BUILD_DIR) && \
		cmake ./..; \
	fi

build: $(BUILD_DIR)
	$(MAKE) --directory=$<

clean:
	rm -rf $(BUILD_DIR)


HW_EXTENSION=$(shell if [ "${SGX_MODE}" = "HW" ]; then echo "-hw"; fi)

FPC_DOCKER_NAMESPACE := hyperledger/fabric-private-chaincode
FPC_DOCKER_CC_BUILDER_NAME = $(FPC_DOCKER_NAMESPACE)-cc-builder$(HW_EXTENSION)

docker-build: clean
	$(DOCKER) image inspect $(FPC_DOCKER_CC_BUILDER_NAME) > /dev/null 2>&1 \
		|| { cd $(TOP)/utils/docker && make cc-builder; }
	$(DOCKER) run \
		-u $$(id -u):$$(id -g)\
		-v "$(DOCKERD_FPC_PATH)/demo/chaincode/fpc":/project/src/github.com/hyperledger-labs/fabric-private-chaincode/demo/chaincode/fpc\
		-w /project/src/github.com/hyperledger-labs/fabric-private-chaincode/demo/chaincode/fpc\
		$(FPC_DOCKER_CC_BUILDER_NAME)\
		sh -c 'make build'

test: build
	./test.sh
