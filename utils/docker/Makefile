# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../..
include $(TOP)/build.mk

FPC_DOCKER_NAMESPACE := hyperledger/fabric-private-chaincode

.PHONY: base ccenv dev

build: base ccenv

base:
	$(DOCKER) build -t $(FPC_DOCKER_NAMESPACE)-base base

ccenv:
	$(DOCKER) build -t $(FPC_DOCKER_NAMESPACE)-ccenv ccenv

dev:
	$(DOCKER) build -t $(FPC_DOCKER_NAMESPACE)-dev dev

run:
	$(DOCKER) run -it $(FPC_DOCKER_NAMESPACE)-dev