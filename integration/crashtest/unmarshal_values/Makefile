TOP = ../../..
include $(TOP)/build.mk

TEST_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
BUILD_DIR := ${TEST_DIR}/_build
CC_NAME := "crash"
DOCKER_IMAGE_NAME := fpc/${CC_NAME}

all: build docker test

$(BUILD_DIR):
	@if [ ! -d $(BUILD_DIR) ]; then \
		mkdir -p $(BUILD_DIR) && \
		cd $(BUILD_DIR) && \
		cmake ./..; \
	fi

build: $(BUILD_DIR)
	$(MAKE) --directory=$<

docker: build
	if [ "${SGX_MODE}" = "HW" ]; then \
		export HW_EXTENSION="-hw" ; \
	fi && \
	make -C ${FPC_PATH}/ecc CC_NAME=${CC_NAME} DOCKER_IMAGE=${DOCKER_IMAGE_NAME}$${HW_EXTENSION} DOCKER_ENCLAVE_SO_PATH=${BUILD_DIR}/lib all docker \
	&& $(DOCKER) tag ${DOCKER_IMAGE_NAME}$${HW_EXTENSION}:$(FPC_VERSION) ${DOCKER_IMAGE_NAME}$${HW_EXTENSION}:latest

test: docker
	CC_ID=$(CC_NAME) ./test.sh

clean:
	rm -rf $(BUILD_DIR)




