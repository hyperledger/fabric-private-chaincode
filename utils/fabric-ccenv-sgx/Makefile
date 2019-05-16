# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../..
include $(TOP)/build.mk

DOCKER_BASE := docker-sgx
IMAGE_NAME := hyperledger/fabric-ccenv-sgx:latest

all: build

$(DOCKER_BASE): 
	git clone https://github.com/tozd/docker-sgx.git $@
	cd $@ && git checkout 08c2457 && ln -s ubuntu-xenial.dockerfile Dockerfile && git apply ../sgx_2.4.patch

build: $(DOCKER_BASE)
	$(DOCKER) build -t $(IMAGE_NAME) $<

# Building this docker image is an expensive tasks. 
# For this reason, we don't use the clean target to wipe
# the docker image.
docker-clean:
	-rm -rf $(DOCKER_BASE)
	-$(DOCKER) rmi $(IMAGE_NAME)
