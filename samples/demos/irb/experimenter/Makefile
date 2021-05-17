# Copyright 2021 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../../../..
include $(TOP)/build.mk

mkfile_path := $(abspath $(firstword $(MAKEFILE_LIST)))
mkfile_dir := $(dir $(mkfile_path))

PDO_CRYPTO_PY_DIR=${FPC_PATH}/common/crypto/pdo/python
PDO_SETUP_PY_PATCH=${mkfile_dir}/pdo_python_setup.py.patch

PDO_CRYPTO_PY_BUILD_DIR=${PDO_CRYPTO_PY_DIR}/build/lib.linux-x86_64-3.8/pdo
PDO_CRYPTO_PY_INSTALL_DIR=${mkfile_dir}/worker/pdo

all: build

${PDO_CRYPTO_PY_BUILD_DIR}:
	cd ${PDO_CRYPTO_PY_DIR} && \
	git apply ${PDO_SETUP_PY_PATCH} && \
	python3.8 setup.py build_ext && \
	PDO_HOME=${./build} python3.8 setup.py bdist_egg

${PDO_CRYPTO_PY_INSTALL_DIR}: ${PDO_CRYPTO_PY_BUILD_DIR}
	cp -r ${PDO_CRYPTO_PY_BUILD_DIR} ${PDO_CRYPTO_PY_INSTALL_DIR}

clean:
	cd ${PDO_CRYPTO_PY_DIR}; python3 setup.py clean --all; git checkout -- setup.py
	rm -rf ${PDO_CRYPTO_PY_INSTALL_DIR}
	make -C test $@

.PHONY: docker
docker: ${PDO_CRYPTO_PY_INSTALL_DIR}
	DOCKER_BUILDKIT=0 docker build -t fpc/experimenter-worker .

build: ${PDO_CRYPTO_PY_INSTALL_DIR} docker
	make -C test $@

