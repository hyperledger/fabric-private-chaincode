# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ..
include $(TOP)/build.mk

BUILD_DIR := _build

$(BUILD_DIR):
	mkdir -p $(BUILD_DIR) && \
		cd $(BUILD_DIR) && \
		cmake ./..

build: $(BUILD_DIR)
	$(MAKE) all --directory=$<

test:
	# make test verbose for now as tlcc doesn't abort on many failures just reports them in the logs ...
	$(MAKE) $@ --directory=$(BUILD_DIR) ARGS="-V"

gen_testcases:
	../integration/deployment_test.sh $(PWD)/test/getLedgerBlocks -o $(PWD)/test/test_blocks

.PHONY: enclave-clean
enclave-clean:
	rm -rf $(BUILD_DIR)
	rm -f enclave/*_t.*
	rm -f trusted_ledger/*_u.*

clean: enclave-clean
