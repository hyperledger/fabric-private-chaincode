# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../../..
include $(TOP)/build.mk

ifeq ($(CC_ID),)
 $(error CC_ID is not set)
endif
ifeq ($(CC_PATH),)
 $(error CC_PATH is not set)
endif


# Define default value for CONTAINER_CLI if not set
CONTAINER_CLI ?= docker

# Check if ${CONTAINER_CLI}-compose command exists
ifeq ($(shell command -v $(CONTAINER_CLI)-compose > /dev/null 2>&1 && echo yes), yes)
    CONTAINER_CLI_COMPOSE ?= $(CONTAINER_CLI)-compose
else
    CONTAINER_CLI_COMPOSE ?= $(CONTAINER_CLI) compose
endif

# Define a target to display the chosen container CLI tools
infoln:
	@echo "Using $(CONTAINER_CLI) and $(CONTAINER_CLI_COMPOSE)"

build: ercc-container ecc-container

ercc-container:
	make -C ${FPC_PATH}/ercc all docker

ecc-container:
	if [ "${SGX_MODE}" = "HW" ]; then \
		export HW_EXTENSION="-hw" ; \
	fi && \
	if [ "${WITH_GO}" = "YES" ]; then \
		export ECC_PATH=${CC_PATH}/mrenclave ; \
	else \
	    export ECC_PATH=${CC_PATH}/_build/lib ; \
	fi && \
	make -C ${CC_PATH} && \
	make -C ${FPC_PATH}/ecc DOCKER_IMAGE=fpc/fpc-${CC_ID}$${HW_EXTENSION} DOCKER_ENCLAVE_SO_PATH=$${ECC_PATH} all docker

ercc-ecc-start: infoln
	if [ ! -z "${DOCKERD_FPC_PATH}" ]; then \
  		export SGX_CREDENTIALS_PATH="${DOCKERD_FPC_PATH}/integration/config/ias"; \
  	fi && \
	if [ "${SGX_MODE}" = "HW" ]; then \
		export HW_EXTENSION="-hw" \
		    AESMD_PATH="/var/run/aesmd" \
		    SGX_DEVICE_PATH=$$(if [ -e "/dev/isgx" ]; then echo "/dev/isgx"; elif [ -e "/dev/sgx/enclave" ]; then echo "/dev/sgx/enclave"; else echo "none"; fi) && \
		[ "$${SGX_DEVICE_PATH}" != "none" ] || ( echo "ERROR: SGX_MODE is HW but no sgx device found"; exit 1; ) \
	fi && \
	env CC_ID=${CC_ID} FPC_VERSION=${FPC_VERSION} ${CONTAINER_CLI_COMPOSE} up

ercc-ecc-stop:
	env CC_ID=${CC_ID} FPC_VERSION=${FPC_VERSION} ${CONTAINER_CLI_COMPOSE} down
