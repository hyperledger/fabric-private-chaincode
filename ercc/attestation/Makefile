# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../..
include $(TOP)/build.mk


IAS_DIR = $(TOP)/config/ias

build: ias-keys

.PHONY: .sgx_mode

IAS_GO_DEFINES=

ifeq (${SGX_MODE}, HW)
IAS_GO_DEFINES+=-DUSE_SGX_HARDWARE_MODE
endif


# put a dependency on the SGX_MODE environment variable
ifneq "$(SGX_MODE)" "$(shell cat .sgx_mode 2>/dev/null)"
SGX_MODE_DEPENDENCY=.sgx_mode
else
SGX_MODE_DEPENDENCY=
endif

ias-keys:
ifeq (${SGX_MODE}, SIM)
	if [ ! -f $(IAS_DIR)/api_key.txt ]; then\
            touch $(IAS_DIR)/api_key.txt;\
	fi;
	if [ ! -f $(IAS_DIR)/spid.txt ]; then\
            echo "DEADBEEF00000000DEADBEEF00000000" > $(IAS_DIR)/spid.txt;\
	fi;
endif

test: build
	$(GO) test $(GOTAGS) -v ./...

clean:
	$(GO) clean
	rm -rf $(GO_TARGETS) .sgx_mode

.sgx_mode:
	@echo $(SGX_MODE) > $@

