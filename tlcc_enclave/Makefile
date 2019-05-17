# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ..
include $(TOP)/build.mk

PROTO_BUILD_DIR := enclave/protos
BUILD_DIR := _build

.PHONY: protos
protos: $(PROTO_BUILD_DIR)

$(PROTO_BUILD_DIR):
	./generate_protos.sh $@

.PHONY: protos-clean
protos-clean:
	-rm -rf $(PROTO_BUILD_DIR)

$(BUILD_DIR): protos
	mkdir -p $(BUILD_DIR) && \
		cd $(BUILD_DIR) && \
		cmake ./..

build: $(BUILD_DIR) protos
	$(MAKE) all --directory=$<

test:
	$(MAKE) $@ --directory=$(BUILD_DIR)

.PHONY: enclave-clean
enclave-clean:
	-rm -rf $(BUILD_DIR)
	-rm -f enclave/*_t.*
	-rm -f trusted_ledger/*_u.*

clean: enclave-clean protos-clean
