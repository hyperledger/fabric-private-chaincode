# SPDX-License-Identifier: Apache-2.0

TOP = ../..
include $(TOP)/build.mk

GO_TEST_DIRS=echo auction kv_test simulation

# no Fabric ccenv/baseos images: FSC needs them only for chaincode the peers
# build themselves, and FPC chaincode runs as its own container
deps: ercc

ercc:
	$(MAKE) -C $(TOP)/ercc docker

test: deps
	$(foreach DIR, $(GO_TEST_DIRS), $(MAKE) -C $(DIR) || exit ;)

