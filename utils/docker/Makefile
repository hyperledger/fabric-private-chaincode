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

DOCKER_DEV_CONTAINER_NAME = fpc-development-${FPC_VERSION}
DOCKER_DEV_BUILD_OPTS ?=
ifdef DOCKER_DEV_IMAGE_APT_ADD__PKGS
	DOCKER_DEV_BUILD_OPTS += --build-arg APT_ADD_PKGS=$(DOCKER_DEV_IMAGE_APT_ADD__PKGS)
endif

DOCKER_DEV_OPTIONAL_CMD= 

DOCKER_GOPATH=/project
DOCKER_FPC_PATH=$(DOCKER_GOPATH)/src/github.com/hyperledger-labs/fabric-private-chaincode

DOCKER_DEV_RUN_OPTS ?=
# - import docker daemon socket (so dev container can run docker)
DOCKER_DEV_RUN_OPTS += -v $(DOCKER_DAEMON_SOCKET):$(DOCKER_DAEMON_SOCKET)
# - mount local fpc repo into the dev container so development inside container is
#   persistant (and also can be done from outside with whatever favorite IDE ...)
DOCKER_DEV_RUN_OPTS += -v "$(abspath ${TOP})":$(DOCKER_FPC_PATH)
# - inject into dev containers environment the path of the hosts FPC_PATH to enable
#   volume mounts inside the container
DOCKER_DEV_RUN_OPTS += --env DOCKERD_FPC_PATH=$(FPC_PATH)/
# - to make it possible to easily access docker-compose exposed ports accessible as localhost
#   inside dev container as you would outside, we map the host network stack into the container
DOCKER_DEV_RUN_OPTS += --net=host
# - if sgx exists, pass also the corresponding device and aesmd socket to dev container
SGX_DEVICE_PATH ?= $(shell if [ -e "/dev/isgx" ]; then echo "/dev/isgx"; elif [ -e "/dev/sgx" ]; then echo "/dev/sgx"; fi)
SGX_PSW_SOCKET ?= /var/run/aesmd
ifneq ($(SGX_DEVICE_PATH),)
	DOCKER_DEV_RUN_OPTS += -v "$(DOCKERD_FPC_PATH)/config/ias/":$(DOCKER_FPC_PATH)/config/ias/ -v $(SGX_PSW_SOCKET):$(SGX_PSW_SOCKET) --device $(SGX_DEVICE_PATH)
endif

.PHONY: base ccenv dev peer cc-builder

# - overall targets
# ------------------
build: base ccenv cc-builder

run: dev
	# Cleanup existing but non-running (note absence of --force in rm!) old dev containers
	dev_container_id=$$(docker ps -a | grep ${DOCKER_DEV_CONTAINER_NAME} | awk '{ print $$1 }'); \
	   [ -z  "$${dev_container_id}" ] || ${DOCKER} rm "$${dev_container_id}"
	# Now run a new instance
	$(DOCKER) run $(DOCKER_DEV_RUN_OPTS) --name ${DOCKER_DEV_CONTAINER_NAME} -it $(FPC_DOCKER_NAMESPACE)-dev:${FPC_VERSION} ${DOCKER_DEV_OPTIONAL_CMD}

clobber:
	# first clean-up dangling images as that might prevent some of the later cleans
	docker system prune --force
	# then clean-up docker-compose network and demo related artifacts
	(cd ../docker-compose; ${MAKE} clobber)
	# delete locally created docker images and left-over peer artifacts
	for img in \
		dev-* \
		dev_test-* \
	; do \
		IMAGES=$$(${DOCKER} images $${img} -q); \
		if [ ! -z "$${IMAGES}" ]; then ${DOCKER} rmi -f $${IMAGES} || exit 1; fi \
	done; \
	for tag in \
	    ${FPC_VERSION} \
	    latest \
	; do \
	    for img in \
	        ${FPC_DOCKER_PEER_NAME} \
		$(FPC_DOCKER_CC_BUILDER_NAME) \
		$(FPC_DOCKER_DEV_NAME) \
		$(FPC_DOCKER_CCENV_NAME) \
		$(FPC_DOCKER_BASE_NAME) \
	    ; do \
		if [ ! -z "$$(docker images -q $${img}:$${tag})" ]; then \
		    ${DOCKER} rmi $${img}:$${tag}; \
		    if [ $$? != 0 ]; then \
			if [ "$${img}" = "$(FPC_DOCKER_DEV_NAME)" ]; then \
			    echo 1>&2 "Could not remove dev container '$(FPC_DOCKER_DEV_NAME)', ignoring error"; \
			else \
			    exit 1; \
			fi \
		    fi \
		fi \
	    done \
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

cc-builder: dev
	$(DOCKER) build $(DOCKER_BUILD_OPTS) \
		-t $(FPC_DOCKER_CC_BUILDER_NAME)\
		--build-arg SGX_MODE=${SGX_MODE}\
		cc-builder
	$(DOCKER) tag $(FPC_DOCKER_CC_BUILDER_NAME) $(FPC_DOCKER_CC_BUILDER_NAME):${FPC_VERSION}

