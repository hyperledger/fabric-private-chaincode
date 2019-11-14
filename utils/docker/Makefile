# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../..
include $(TOP)/build.mk

FPC_DOCKER_NAMESPACE := hyperledger/fabric-private-chaincode
FPC_PEER_DOCKER_NAMESPACE := hyperledger/fabric-peer-fpc

DOCKER_DAEMON_SOCKET ?= /var/run/docker.sock
FABRIC_PEER_DAEMON_CHAINCODE_PORT ?= 7052

DOCKER_DEV_BUILD_OPTS ?=
ifdef DOCKER_DEV_IMAGE_APT_ADD__PKGS
	DOCKER_DEV_BUILD_OPTS += --build-arg APT_ADD_PKGS=$(DOCKER_DEV_IMAGE_APT_ADD__PKGS)
endif


DOCKER_DEV_RUN_OPTS ?=
DOCKER_DEV_RUN_OPTS += -p $(FABRIC_PEER_DAEMON_CHAINCODE_PORT):$(FABRIC_PEER_DAEMON_CHAINCODE_PORT) -v $(DOCKER_DAEMON_SOCKET):$(DOCKER_DAEMON_SOCKET)
DOCKER_GOPATH=/project
DOKER_FPCPATH=$(DOCKER_GOPATH)/src/github.com/hyperledger-labs/fabric-private-chaincode
SGX_DEVICE_PATH ?= $(shell if [ -e "/dev/isgx" ]; then echo "/dev/isgx"; elif [ -e "/dev/sgx" ]; then echo "/dev/sgx"; fi)
SGX_PSW_SOCKET ?= /var/run/aesmd
ifneq ($(SGX_DEVICE_PATH),)
	DOCKER_DEV_RUN_OPTS += -v $(PWD)/../../config/ias/:$(DOKER_FPCPATH)/config/ias/ -v $(SGX_PSW_SOCKET):$(SGX_PSW_SOCKET) --device $(SGX_DEVICE_PATH)
endif

.PHONY: base ccenv dev peer

build: base ccenv

base:
	$(DOCKER) build -t $(FPC_DOCKER_NAMESPACE)-base base

ccenv: base
	$(DOCKER) build -t $(FPC_DOCKER_NAMESPACE)-ccenv ccenv

dev: base
	$(DOCKER) build -t $(FPC_DOCKER_NAMESPACE)-dev $(DOCKER_DEV_BUILD_OPTS) dev

peer: base
	$(DOCKER) build -t ${FABRIC_PEER_DOCKER_NAMESPACE} peer

run: dev
	$(DOCKER) run $(DOCKER_DEV_RUN_OPTS) -it $(FPC_DOCKER_NAMESPACE)-dev
