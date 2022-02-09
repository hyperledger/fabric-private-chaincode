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

build: ercc-container ecc-container

ercc-container:
	make -C ${FPC_PATH}/ercc all docker

ecc-container:
	if [ "${SGX_MODE}" = "HW" ]; then \
		export HW_EXTENSION="-hw" ; \
	fi && \
	make -C ${CC_PATH} && \
	make -C ${FPC_PATH}/ecc DOCKER_IMAGE=fpc/fpc-${CC_ID}$${HW_EXTENSION} DOCKER_ENCLAVE_SO_PATH=${CC_PATH}/_build/lib all docker

ercc-ecc-start:
	if [ ! -z "${DOCKERD_FPC_PATH}" ]; then \
  		export SGX_CREDENTIALS_PATH="${DOCKERD_FPC_PATH}/integration/config/ias"; \
  	fi && \
	if [ "${SGX_MODE}" = "HW" ]; then \
		export HW_EXTENSION="-hw" \
		    AESMD_PATH="/var/run/aesmd" \
		    SGX_DEVICE_PATH=$$(if [ -e "/dev/isgx" ]; then echo "/dev/isgx"; elif [ -e "/dev/sgx/enclave" ]; then echo "/dev/sgx/enclave"; else echo "none"; fi) && \
		[ "$${SGX_DEVICE_PATH}" != "none" ] || ( echo "ERROR: SGX_MODE is HW but no sgx device found"; exit 1; ) \
	fi && \
	env CC_ID=${CC_ID} FPC_VERSION=${FPC_VERSION} docker-compose up

ercc-ecc-stop:
	env CC_ID=${CC_ID} FPC_VERSION=${FPC_VERSION} docker-compose down
