# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../..
include $(TOP)/build.mk

HW_EXTENSION=$(shell if [ "${SGX_MODE}" = "HW" ]; then echo "-hw"; fi)

FPC_DOCKER_NAMESPACE := hyperledger/fabric-private-chaincode
FPC_DOCKER_CC_BUILDER_NAME = $(FPC_DOCKER_NAMESPACE)-cc-builder$(HW_EXTENSION)
FPC_DOCKER_DEV_NAME = $(FPC_DOCKER_NAMESPACE)-dev
FPC_DOCKER_CCENV_NAME = $(FPC_DOCKER_NAMESPACE)-ccenv
FPC_DOCKER_BASE_NAME = $(FPC_DOCKER_NAMESPACE)-base

FPC_DOCKER_PEER_NAMESPACE := hyperledger/fabric-peer-fpc
FPC_DOCKER_PEER_NAME = $(FPC_DOCKER_PEER_NAMESPACE)$(HW_EXTENSION)

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

.PHONY: base ccenv dev peer cc-builder

# - overall targets
# ------------------
build: base ccenv cc-builder

run: dev
	$(DOCKER) run $(DOCKER_DEV_RUN_OPTS) -it $(FPC_DOCKER_NAMESPACE)-dev

clobber:
	# first clean-up dangling images as that might prevent some of the later cleans
	docker system prune --force
	# then clean-up docker-compose network and demo related artifacts
	(cd ../docker-compose; ${MAKE} clobber)
	# delete locally created docker images and left-over peer artifacts
	for imgs in \
		dev-* \
		dev_test-* \
	        ${FPC_DOCKER_PEER_NAME} \
		$(FPC_DOCKER_CC_BUILDER_NAME) \
		$(FPC_DOCKER_DEV_NAME) \
		$(FPC_DOCKER_CCENV_NAME) \
		$(FPC_DOCKER_BASE_NAME) \
	; do \
		IMAGES=$$(${DOCKER} images $${imgs} -q); \
		if [ ! -z "$${IMAGES}" ]; then ${DOCKER} rmi -f $${IMAGES} || exit 1; fi \
	done


# - building individual docker images
# ------------------------------------------------------
# Note:
# - docker images are not necessarily rebuild if they exist but are outdated.
#   To force rebuild you have two options
#   - do a 'make clobber' first. This ensures you will have the uptodate images
#     but is a broad and slow brush
#   - to just fore rebuilding an image, call `make` with FORCE_REBUILD defined
DOCKER_BUILD_OPTS ?=
ifdef FORCE_REBUILD
	DOCKER_BUILD_OPTS += --no-cache
endif


base:
	$(DOCKER) build $(DOCKER_BUILD_OPTS) -t $(FPC_DOCKER_BASE_NAME) base

ccenv: base
	$(DOCKER) build $(DOCKER_BUILD_OPTS) -t $(FPC_DOCKER_CCENV_NAME) ccenv

# Note: for overall consistency reasons we want the FPC code from the current repo
# as this will rebuild each time and take a while, we don't add a dependency
# to the build target above ...
peer: base
	(cd ${TOP}; \
         $(DOCKER) build $(DOCKER_BUILD_OPTS) -t ${FPC_DOCKER_PEER_NAME}\
         -f utils/docker/peer/Dockerfile\
         --build-arg FPC_REPO_URL=file:///tmp/build-src/.git\
         --build-arg FPC_REPO_BRANCH=$$(git rev-parse --abbrev-ref HEAD)\
         --build-arg SGX_MODE=${SGX_MODE}\
         . )

dev: base
	$(DOCKER) build $(DOCKER_BUILD_OPTS) -t $(FPC_DOCKER_DEV_NAME) $(DOCKER_DEV_BUILD_OPTS) dev

cc-builder: dev
	$(DOCKER) build $(DOCKER_BUILD_OPTS) \
		-t $(FPC_DOCKER_CC_BUILDER_NAME)\
		--build-arg SGX_MODE=${SGX_MODE}\
		cc-builder

