# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../..
include $(TOP)/build.mk

BUILD_DIR := _build

PDO_ENCLAVE_CODE_SIGN_PEM := "pdo_enclave_code_sign.pem"

$(BUILD_DIR):
	@if [ ! -d $(BUILD_DIR) ]; then \
		mkdir -p $(BUILD_DIR) && \
        cd $(BUILD_DIR) && \
		openssl genrsa -3 -out $(PDO_ENCLAVE_CODE_SIGN_PEM) 3072 && \
        export PDO_ENCLAVE_CODE_SIGN_PEM=`readlink -f $(PDO_ENCLAVE_CODE_SIGN_PEM)` && \
        cmake ./.. ; \
    fi

build: $(BUILD_DIR)
	$(MAKE) --directory=$<

test: build
	$(MAKE) -C $(BUILD_DIR) test

clean:
	rm -rf $(BUILD_DIR)
