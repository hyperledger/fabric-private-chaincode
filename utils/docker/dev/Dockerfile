# Copyright IBM Corp. All Rights Reserved.
# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

# Description:
#   Sets up fabric source environment and builds the following containers:
#   - dev: a fpc-enabled interactive development environment
#
#  Configuration (build) paramaters (for defaults, see below section with ARGs)
#  - fpc image version:         FPC_VERSION
#  - fabric repo:               FABRIC_REPO
#  - fabric branch:             FABRIC_VERSION
#  - git user:                  GIT_USER_NAME
#  - git user's email:          GIT_USER_EMAIL
#  - sgx mode:                  SGX_MODE
#  - additional apt pkgs:       APT_ADD_PKGS

# global config params
ARG FPC_VERSION=main

FROM hyperledger/fabric-private-chaincode-base-dev:${FPC_VERSION}

# config/build params
ARG FABRIC_REPO=https://github.com/hyperledger/fabric.git
ARG FABRIC_VERSION=2.3.3

ARG FABRIC_REL_PATH=src/github.com/hyperledger/fabric
ARG FPC_REL_PATH=src/github.com/hyperledger/fabric-private-chaincode
ARG APT_ADD_PKGS

ENV FPC_PATH=${GOPATH}/${FPC_REL_PATH}
ENV FPC_VERSION=${FPC_VERSION}

# we set default SGX_MODE to simulation
# this can be set via DOCKER_DEV_RUN_OPTS += --env SGX_MODE=$(SGX_MODE)
ENV SGX_MODE=SIM

RUN apt-get update \
 && apt-get install -y -q \
	${APT_ADD_PKGS}

# make sure we have a valid git user (needed for the git am when patching fabric)
ARG GIT_USER_NAME=tester
ARG GIT_USER_EMAIL=tester@fpc
RUN git config --global user.name "${GIT_USER_NAME}" \
 && git config --global user.email "${GIT_USER_EMAIL}"

# Get Fabric
ENV FABRIC_PATH=${GOPATH}/${FABRIC_REL_PATH}
RUN git clone --branch v${FABRIC_VERSION} ${FABRIC_REPO} ${FABRIC_PATH}
# Note: could add --single-branch to below to speed-up and keep size smaller. But for now for a dev-image better keep complete repo

WORKDIR ${FPC_PATH}
