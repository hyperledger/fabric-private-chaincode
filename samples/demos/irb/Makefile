# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../../..
include $(TOP)/build.mk

COMPONENTS = protos $(FPC_PATH)/sgx/samples/demos/irb/chaincode experimenter

# redis for the experimenter test; no Fabric ccenv/baseos images, as in
# integration/go_chaincode/Makefile
pull-images:
	$(DOCKER) pull redis:latest

ercc:
	make -C ${FPC_PATH}/ercc all docker
	$(DOCKER) image tag fpc/ercc:$(FPC_VERSION) fpc/ercc:latest

build: pull-images ercc
	$(foreach DIR, $(COMPONENTS), $(MAKE) -C $(DIR) $@ || exit;)

test:
	FABRIC_LOGGING_SPEC=fpc=debug:grpc=error:comm.grpc=error:gossip=warning:info $(GO) test -v -failfast ./...

run:
	$(GO) test -v .

clean:
	$(GO) clean
	rm -rf cmd users
	$(foreach DIR, $(COMPONENTS), $(MAKE) -C $(DIR) $@ || exit;)
