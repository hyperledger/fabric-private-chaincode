# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../../../..
include $(TOP)/build.mk

BUILD_DIR := _build
DOCKER_IMAGE_NAME := fpc/irb-experiment

all: docker

$(BUILD_DIR):
	@if [ ! -d $(BUILD_DIR) ]; then \
		mkdir -p $(BUILD_DIR) && \
		cd $(BUILD_DIR) && \
		cmake ./..; \
	fi

build: protos $(BUILD_DIR)
	$(MAKE) --directory=$(BUILD_DIR)

clean: protos-clean
	rm -rf $(BUILD_DIR)

docker: build
	if [ "${SGX_MODE}" = "HW" ]; then \
		export HW_EXTENSION="-hw" ; \
	fi && \
	make -C ${FPC_PATH}/ecc DOCKER_IMAGE=${DOCKER_IMAGE_NAME}$${HW_EXTENSION} DOCKER_ENCLAVE_SO_PATH=${FPC_PATH}/samples/demos/irb/chaincode/${BUILD_DIR}/lib all docker

PROTO_BUILD_DIR := _protos
PROTO_PATH := ../protos
PROTO_FILES := irb.proto

.PHONY: protos
protos: $(PROTO_BUILD_DIR)
	$(PROTOC_CMD) \
		--plugin=protoc-gen-nanopb=$(NANOPB_PATH)/generator/protoc-gen-nanopb-py2 \
		--proto_path=$(PROTO_PATH) "--nanopb_out=-f ${PROTO_PATH}/irb.options:$(PROTO_BUILD_DIR)" $(PROTO_FILES)

$(PROTO_BUILD_DIR):
	mkdir -p $@

.PHONY: protos-clean
protos-clean:
	rm -rf $(PROTO_BUILD_DIR)
