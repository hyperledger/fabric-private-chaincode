# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

#ifneq ($(FPC_PATH), '')
ifdef FPC_PATH
	TOP = $(FPC_PATH)
else
	TOP = ../../../
	export FPC_PATH = $(abspath $(TOP))
endif

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
	-rm -rf $(BUILD_DIR)

docker-build: clean
	$(DOCKER) image inspect hyperledger/fabric-private-chaincode-cc-builder > /dev/null 1>&1 || { cd $(TOP)/utils/docker && make cc-builder; }
	$(DOCKER) run -u $$(id -u):$$(id -g) -v ${PWD}:/project/src/github.com/hyperledger-labs/fabric-private-chaincode/demo/chaincode/fpc -w /project/src/github.com/hyperledger-labs/fabric-private-chaincode/demo/chaincode/fpc hyperledger/fabric-private-chaincode-cc-builder sh -c make build

test: build
	./test.sh
