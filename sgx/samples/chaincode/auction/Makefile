# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../../..
include $(TOP)/build.mk

BUILD_DIR := _build
CC_NAME := "auction"
DOCKER_IMAGE_NAME := fpc/${CC_NAME}

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

docker: build
	if [ "${SGX_MODE}" = "HW" ]; then \
		export HW_EXTENSION="-hw" ; \
	fi && \
	make -C ${FPC_PATH}/ecc CC_NAME=${CC_NAME} DOCKER_IMAGE=${DOCKER_IMAGE_NAME}$${HW_EXTENSION} DOCKER_ENCLAVE_SO_PATH=$(PWD)/${BUILD_DIR}/lib all docker
