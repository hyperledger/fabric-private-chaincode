# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../../..
include $(TOP)/build.mk

COMPONENTS = protos chaincode experimenter

all: ercc build test clean

pull-images:
	$(DOCKER) pull redis:latest
	$(DOCKER) pull hyperledger/fabric-baseos:2.5
	$(DOCKER) image tag hyperledger/fabric-baseos:2.5 hyperledger/fabric-baseos:latest
	$(DOCKER) pull hyperledger/fabric-ccenv:2.5
	$(DOCKER) image tag hyperledger/fabric-ccenv:2.5 hyperledger/fabric-ccenv:latest

ercc:
	make -C ${FPC_PATH}/ercc all docker
	$(DOCKER) image tag fpc/ercc:$(FPC_VERSION) fpc/ercc:latest

build: ercc
	$(foreach DIR, $(COMPONENTS), $(MAKE) -C $(DIR) $@ || exit;)

test: pull-images
	$(GO) test -v -failfast ./...

run:
	$(GO) test -v -failfast .

clean:
	$(GO) clean
	rm -rf cmd users
	$(foreach DIR, $(COMPONENTS), $(MAKE) -C $(DIR) $@ || exit;)
