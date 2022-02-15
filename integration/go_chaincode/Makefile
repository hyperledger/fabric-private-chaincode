# SPDX-License-Identifier: Apache-2.0

TOP = ../..
include $(TOP)/build.mk

GO_TEST_DIRS=auction kv_test

deps: ercc images

ercc:
	$(MAKE) -C $(TOP)/ercc docker

images:
	$(DOCKER) pull hyperledger/fabric-baseos:2.3
	$(DOCKER) image tag hyperledger/fabric-baseos:2.3 hyperledger/fabric-baseos:latest
	$(DOCKER) pull hyperledger/fabric-ccenv:2.3
	$(DOCKER) image tag hyperledger/fabric-ccenv:2.3 hyperledger/fabric-ccenv:latest

test: deps
	$(foreach DIR, $(GO_TEST_DIRS), $(MAKE) -C $(DIR) || exit ;)

