# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../../..
include $(TOP)/build.mk

.PHONY: ercc-ecc-start
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
	if [ ! -z "${EXTRA_COMPOSE_FILE}" ]; then \
		env FPC_VERSION=${FPC_VERSION} ${DOCKER_COMPOSE} -f compose.yaml -f ${EXTRA_COMPOSE_FILE} up -d; \
	else \
		env FPC_VERSION=${FPC_VERSION} ${DOCKER_COMPOSE} up -d; \
	fi

.PHONY: ercc-ecc-start
ercc-ecc-stop:
	${DOCKER_COMPOSE} down

clean:
	docker compose down -v
	./fabric-samples/test-network/network.sh down || true
	rm -rf blockchain-explorer fabric-samples .env
	docker system prune --force
	docker network prune --force
	docker volume prune --force

test:
	./test.sh
