# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../..
include $(TOP)/build.mk


TEST_EXAMPLE ?= echo


build: ercc-container ecc-container

ercc-container:
	make -C ${FPC_PATH}/ercc all docker

ecc-container:
	if [ "${SGX_MODE}" = "HW" ]; then \
		export HW_EXTENSION="-hw" ; \
	fi && \
	make -C ${FPC_PATH}/ecc DOCKER_IMAGE=fpc/fpc-${TEST_EXAMPLE}$${HW_EXTENSION} DOCKER_ENCLAVE_SO_PATH=${FPC_PATH}/examples/${TEST_EXAMPLE}/_build/lib all docker


ercc-ecc-run:
	if [ "${SGX_MODE}" = "HW" ]; then \
		export HW_EXTENSION="-hw" \
		    AESMD_PATH="/var/run/aesmd" \
		    SGX_DEVICE_PATH=$$(if [ -e "/dev/isgx" ]; then echo "/dev/isgx"; elif [ -e "/dev/sgx" ]; then echo "/dev/sgx"; else echo "none"; fi) && \
		[ "$${SGX_DEVICE_PATH}" != "none" ] || ( echo "ERROR: SGX_MODE is HW but no sgx device found"; exit 1; ) \
	fi && \
	env CC_ID=${TEST_EXAMPLE} docker-compose up

ercc-ecc-stop:
	env CC_ID=${TEST_EXAMPLE} docker-compose down
