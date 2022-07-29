# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../..
include $(TOP)/build.mk

HW_EXTENSION=$(shell if [ "${SGX_MODE}" = "HW" ]; then echo "-hw"; fi)


# Names and namespaces
# ------------------
DOCKER_REGISTRY ?= ghcr.io

FPC_DOCKER_NAMESPACE := hyperledger/fabric-private-chaincode
FPC_DOCKER_CC_BUILDER_NAME = $(FPC_DOCKER_NAMESPACE)-cc-builder$(HW_EXTENSION)
FPC_DOCKER_DEV_NAME = $(FPC_DOCKER_NAMESPACE)-dev
FPC_DOCKER_CCENV_NAME = $(FPC_DOCKER_NAMESPACE)-ccenv
FPC_DOCKER_BASE_RT_NAME = $(FPC_DOCKER_NAMESPACE)-base-rt
FPC_DOCKER_BASE_DEV_NAME = $(FPC_DOCKER_NAMESPACE)-base-dev

FPC_DOCKER_PEER_NAMESPACE := hyperledger/fabric-peer-fpc
FPC_DOCKER_PEER_NAME = $(FPC_DOCKER_PEER_NAMESPACE)$(HW_EXTENSION)

DOCKER_DEV_CONTAINER_NAME = fpc-development-${FPC_VERSION}


# Docker & sgx configs
# ------------------------
DOCKER_DAEMON_SOCKET ?= /var/run/docker.sock

SGX_DEVICE_PATH ?= $(shell if [ -e "/dev/isgx" ]; then echo "/dev/isgx"; elif [ -e "/dev/sgx/enclave" ]; then echo "/dev/sgx/enclave"; fi)
SGX_PSW_SOCKET ?= /var/run/aesmd

DOCKER_GOPATH=/project
DOCKER_FPC_PATH=$(DOCKER_GOPATH)/src/github.com/hyperledger/fabric-private-chaincode



# Docker build options
# ------------------
# - docker images are not necessarily rebuild if they exist but are outdated.
#   To force rebuild you have two options
#   - do a 'make clobber' first. This ensures you will have the uptodate images
#     but is a broad and slow brush
#   - to just fore rebuilding an image, call `make` with DOCKER_FORCE_REBUILD defined
#   - to keep docker build quiet unless there is an error, call `make` with DOCKER_QUIET_BUILD defined
DOCKER_BUILD_OPTS ?=
ifdef DOCKER_QUIET_BUILD
	DOCKER_BUILD_OPTS += --quiet
endif
ifdef DOCKER_FORCE_REBUILD
	DOCKER_BUILD_OPTS += --no-cache
endif
DOCKER_BUILD_OPTS += --build-arg FPC_VERSION=$(FPC_VERSION)
DOCKER_BASE_RT_BUILD_OPTS ?=
# - allow additional custom apt pkgs for any container derived from base-rt based on DOCKER_BASE_RT_IMAGE_APT_ADD_PKGS
ifdef DOCKER_BASE_RT_IMAGE_APT_ADD_PKGS
	DOCKER_BASE_RT_BUILD_OPTS += --build-arg APT_ADD_PKGS=$(DOCKER_BASE_RT_IMAGE_APT_ADD_PKGS)
endif
DOCKER_BASE_DEV_BUILD_OPTS ?=
# - allow additional custom apt pkgs for any container derived from base-dev based on DOCKER_BASE_DEV_IMAGE_APT_ADD_PKGS
ifdef DOCKER_BASE_DEV_IMAGE_APT_ADD_PKGS
	DOCKER_BASE_DEV_BUILD_OPTS += --build-arg APT_ADD_PKGS=$(DOCKER_BASE_DEV_IMAGE_APT_ADD_PKGS)
endif
DOCKER_DEV_BUILD_OPTS ?=
# - for all images working on git, e.g., applying of fabric patches inside of docker, we need proper git user config

DOCKER_DEV_BUILD_OPTS += $(shell git config user.name >/dev/null && echo --build-arg GIT_USER_NAME=\'$$(git config user.name)\')
DOCKER_DEV_BUILD_OPTS += $(shell git config user.email >/dev/null && echo --build-arg GIT_USER_EMAIL=\'$$(git config user.email)\')
# - allow additional custom apt pkgs for 'dev' container based on DOCKER_DEV_IMAGE_APT_ADD_PKGS
ifdef DOCKER_DEV_IMAGE_APT_ADD_PKGS
	DOCKER_DEV_BUILD_OPTS += --build-arg APT_ADD_PKGS=$(DOCKER_DEV_IMAGE_APT_ADD_PKGS)
endif



# Docker run options
# ------------------
DOCKER_DEV_RUN_OPTS ?=
# - import docker daemon socket (so dev container can run docker)
DOCKER_DEV_RUN_OPTS += -v $(DOCKER_DAEMON_SOCKET):$(DOCKER_DAEMON_SOCKET)
# - if ~/.docker/ exists, pass it also along to container so we have consistent config
ifneq ($(wildcard ${HOME}/.docker),)
DOCKER_DEV_RUN_OPTS += -v "${HOME}/.docker":"/root/.docker"
endif
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
#   (see above for definitions)
ifneq ($(SGX_DEVICE_PATH),)
	DOCKER_DEV_RUN_OPTS += -v "$(DOCKERD_FPC_PATH)/config/ias/":$(DOCKER_FPC_PATH)/config/ias/ -v $(SGX_PSW_SOCKET):$(SGX_PSW_SOCKET) --device $(SGX_DEVICE_PATH)
endif
# - pass host sgx mode
DOCKER_DEV_RUN_OPTS += --env SGX_MODE=$(SGX_MODE)

ifeq (${DOCKER_DEV_CI_MODE},)
DOCKER_DEV_RUN_OPTS += -i
DOCKER_DEV_RUN_OPTS += -e CI=true
endif

# In CI build, pass the required coverage variable to the docker container.
# These variables will be necessary after the tests to report the coverage "from inside the container".
ifeq (${IS_CI_RUNNING}, true)
ifeq (${CODE_COVERAGE_ENABLED}, true)
DOCKER_DEV_RUN_OPTS += `curl -s https://codecov.io/env | bash`
endif
endif

# Run a specific command (rather than bash) with 'make run-dev' by defining
# the variable DOCKER_DEV_OPTIONAL_CMD
DOCKER_DEV_OPTIONAL_CMD=



# - overall make targets
# ------------------------------
.PHONY: base-rt base-dev ccenv dev peer cc-builder

build: ccenv
# Note 1: we do _not_ include dev here as this is not needed in integration tests and
# a rebuild could cause trouble for a dev container user.
# Note 2: we removed "cc-builder peer" as dependencies since they are currenty unused (but might be helpful).

build-dev: base-dev

run-dev: dev
	# Cleanup existing but non-running (note absence of --force in rm!) old dev containers
	dev_container_id=$$(docker ps -a | grep ${DOCKER_DEV_CONTAINER_NAME} | awk '{ print $$1 }'); \
	   [ -z  "$${dev_container_id}" ] || ${DOCKER} rm "$${dev_container_id}"
	# Now run a new instance
	$(DOCKER) run $(DOCKER_DEV_RUN_OPTS) --name ${DOCKER_DEV_CONTAINER_NAME} -t $(FPC_DOCKER_DEV_NAME):${FPC_VERSION} ${DOCKER_DEV_OPTIONAL_CMD}

clobber:
	# first clean-up dangling images as that might prevent some of the later cleans
	docker system prune --force
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
	; do \
	    for img in \
	        ${FPC_DOCKER_PEER_NAME} \
		$(FPC_DOCKER_CC_BUILDER_NAME) \
		$(FPC_DOCKER_DEV_NAME) \
		$(FPC_DOCKER_CCENV_NAME) \
		$(FPC_DOCKER_BASE_RT_NAME) \
		$(FPC_DOCKER_BASE_DEV_NAME) \
		$(DOCKER_REGISTRY)/$(FPC_DOCKER_CCENV_NAME) \
		$(DOCKER_REGISTRY)/$(FPC_DOCKER_BASE_DEV_NAME) \
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
base-rt:
	$(DOCKER) build $(DOCKER_BUILD_OPTS) $(DOCKER_BASE_RT_BUILD_OPTS) -t $(FPC_DOCKER_BASE_RT_NAME):$(FPC_VERSION) base-rt

base-dev: base-rt
	$(DOCKER) build $(DOCKER_BUILD_OPTS) $(DOCKER_BASE_DEV_BUILD_OPTS) -t $(FPC_DOCKER_BASE_DEV_NAME):$(FPC_VERSION) base-dev

ccenv: base-rt
	$(DOCKER) build $(DOCKER_BUILD_OPTS) -t $(FPC_DOCKER_CCENV_NAME):$(FPC_VERSION) ccenv

# Note: For all docker images, dependencis will cause docker rebuilds.
# Docker caching though makes this though fast if nothing has changed.
# However, for below for overall consistency reason, we use $FPC_TOP/.git 
# for the source to build from. This means that the context is $FPC_TOP
# (restricted to the bare minimually necessary, in particulart, .git,
# via ../../.dockerignore) and implies that any change in the context
# will also cause a rebuild. Given the filtering this mostly applies
# only to changes in git state (e.g., git pull, git add, git commit ...).
# The images should also be in a way which minimizes what will be rebuilt
# in that context. Additionally, the building should happen only as last
# step before the demo, so any code-induced build failure should happen
# before any docker rebuild happens.
peer: base-dev
	(cd ${TOP} &&\
         $(DOCKER) build $(DOCKER_BUILD_OPTS) -t $(FPC_DOCKER_PEER_NAME):$(FPC_VERSION)\
         $(DOCKER_DEV_BUILD_OPTS)\
	 --target peer\
         -f ./utils/docker/dev_peer_cc-builder/Dockerfile\
         --build-arg FPC_REPO_URL=file:///tmp/cloned-local-fpc-git-repo\
         --build-arg FPC_REPO_BRANCH_TAG_OR_COMMIT=$$(git rev-parse HEAD)\
         --build-arg SGX_MODE=${SGX_MODE}\
         . )

dev:
	(cd ${TOP} &&\
         $(DOCKER) build $(DOCKER_BUILD_OPTS) -t $(FPC_DOCKER_DEV_NAME):$(FPC_VERSION)\
         $(DOCKER_DEV_BUILD_OPTS)\
         -f ./utils/docker/dev/Dockerfile\
         . )

cc-builder: base-dev
	(cd ${TOP} &&\
         $(DOCKER) build $(DOCKER_BUILD_OPTS) -t $(FPC_DOCKER_CC_BUILDER_NAME):$(FPC_VERSION)\
         $(DOCKER_DEV_BUILD_OPTS)\
         -f ./utils/docker/dev_peer_cc-builder/Dockerfile\
	 --target cc-builder \
         --build-arg FPC_REPO_URL=file:///tmp/cloned-local-fpc-git-repo\
         --build-arg FPC_REPO_BRANCH_TAG_OR_COMMIT=$$(git rev-parse HEAD)\
         --build-arg SGX_MODE=${SGX_MODE}\
         . )

publish:
	for img in \
		$(FPC_DOCKER_CCENV_NAME) \
		$(FPC_DOCKER_BASE_DEV_NAME) \
	; do \
	  	echo $${img}; \
		$(DOCKER) tag $${img}:$(FPC_VERSION) $(DOCKER_REGISTRY)/$${img}:$(FPC_VERSION); \
		$(DOCKER) push $(DOCKER_REGISTRY)/$${img}:$(FPC_VERSION); \
	done;

pull:
	# ccenv
	$(DOCKER) pull $(DOCKER_REGISTRY)/$(FPC_DOCKER_CCENV_NAME):$(FPC_VERSION)
	$(DOCKER) image tag $(DOCKER_REGISTRY)/$(FPC_DOCKER_CCENV_NAME):$(FPC_VERSION) $(FPC_DOCKER_CCENV_NAME):$(FPC_VERSION)

pull-dev:
	# base-dev
	$(DOCKER) pull $(DOCKER_REGISTRY)/$(FPC_DOCKER_BASE_DEV_NAME):$(FPC_VERSION)
	$(DOCKER) image tag $(DOCKER_REGISTRY)/$(FPC_DOCKER_BASE_DEV_NAME):$(FPC_VERSION) $(FPC_DOCKER_BASE_DEV_NAME):$(FPC_VERSION)
