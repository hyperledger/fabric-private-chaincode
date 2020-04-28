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
# needed because we're doing the apply of fabric patches inside of docker, where git is unconfigured
DOCKER_DEV_BUILD_OPTS += --build-arg GIT_USER_NAME="$(shell git config user.name)" --build-arg GIT_USER_EMAIL="$(shell git config user.email)"

DOCKER_DEV_RUN_OPTS ?=
DOCKER_DEV_RUN_OPTS += -p $(FABRIC_PEER_DAEMON_CHAINCODE_PORT):$(FABRIC_PEER_DAEMON_CHAINCODE_PORT) -v $(DOCKER_DAEMON_SOCKET):$(DOCKER_DAEMON_SOCKET)
DOCKER_DEV_RUN_OPTS += -v $(PWD)/../../:/project/src/github.com/hyperledger-labs/fabric-private-chaincode

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
	$(DOCKER) run $(DOCKER_DEV_RUN_OPTS) -it $(FPC_DOCKER_NAMESPACE)-dev:${FPC_VERSION}

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
DOCKER_BUILD_OPTS += --build-arg FPC_VERSION=$(FPC_VERSION)


base:
	$(DOCKER) build $(DOCKER_BUILD_OPTS) -t $(FPC_DOCKER_BASE_NAME) base
	$(DOCKER) tag $(FPC_DOCKER_BASE_NAME) $(FPC_DOCKER_BASE_NAME):${FPC_VERSION}

ccenv: base
	$(DOCKER) build $(DOCKER_BUILD_OPTS) -t $(FPC_DOCKER_CCENV_NAME) ccenv
	$(DOCKER) tag $(FPC_DOCKER_CCENV_NAME) $(FPC_DOCKER_CCENV_NAME):${FPC_VERSION}

# Note: for overall consistency reasons we want the FPC code from the current repo
# as this will rebuild each time and take a while, we don't add a dependency
# to the build target above for peer and dev ...
# To enable docker to maximize build sharing between dev and peer
# we add DOCKER_DEV_BUILD_OPTS (e.g., to get additional packages) also to peer
peer: base
	(cd ${TOP}; \
         $(DOCKER) build $(DOCKER_BUILD_OPTS) -t ${FPC_DOCKER_PEER_NAME}\
         $(DOCKER_DEV_BUILD_OPTS)\
         -f utils/docker/peer/Dockerfile\
         --build-arg FPC_REPO_URL=file:///tmp/cloned-local-fpc-git-repo\
         --build-arg FPC_REPO_BRANCH_TAG_OR_COMMIT=$$(git rev-parse HEAD)\
         --build-arg SGX_MODE=${SGX_MODE}\
         . )
	$(DOCKER) tag $(FPC_DOCKER_PEER_NAME) $(FPC_DOCKER_PEER_NAME):${FPC_VERSION}

dev: base
	(cd ${TOP}; \
         $(DOCKER) build $(DOCKER_BUILD_OPTS) -t $(FPC_DOCKER_DEV_NAME)\
         $(DOCKER_DEV_BUILD_OPTS)\
         -f utils/docker/dev/Dockerfile\
         --build-arg FPC_REPO_URL=file:///tmp/cloned-local-fpc-git-repo\
         --build-arg FPC_REPO_BRANCH_TAG_OR_COMMIT=$$(git rev-parse HEAD)\
         --build-arg SGX_MODE=${SGX_MODE}\
         . )
	$(DOCKER) tag $(FPC_DOCKER_DEV_NAME) $(FPC_DOCKER_DEV_NAME):${FPC_VERSION}

# TODO bring cc-builder back
cc-builder: dev
#	$(DOCKER) build $(DOCKER_BUILD_OPTS) \
#		-t $(FPC_DOCKER_CC_BUILDER_NAME)\
#		--build-arg SGX_MODE=${SGX_MODE}\
#		cc-builder
#	$(DOCKER) tag $(FPC_DOCKER_CC_BUILDER_NAME) $(FPC_DOCKER_CC_BUILDER_NAME):${FPC_VERSION}

