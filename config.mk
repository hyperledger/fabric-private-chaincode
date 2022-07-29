# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

# Note: any of these settings can be overridden in an (optional) file
#   `config.override.mk` in this directory!


# Go related settings
#--------------------------------------------------
GOFLAGS :=
GO_CMD := go


# Docker related settings
#--------------------------------------------------
export DOCKER_BUILDKIT ?= 0 
# Building with build-kit makes multi-stage builds more efficient
# and also provides nicer output. However, as docker from older
# versions of ubuntu 18.04 can hang and travis explicitly rejects
# (rather than ignores) it, we support both and leave the default
# as a more robust 0. If you prefer the benefits of buildkit,
# override default in your `config.override.mk`
DOCKERFLAGS :=
DOCKER_CMD := docker
# Note:
# - to get quiet docker builds, you can define in config.override.mk
#   DOCKER_QUIET_BUILD=1
# - similarly you could also always turn off docker caching and force
#   a complete rebuild by defining DOCKER_FORCE_REBUILD=1 (although
#   this will have drastic performance implication and you might be better
#   off doing that selectively on particular builds and/or use 
#   `make clobber`.
# - also useful docker overrides are following variables which allow you
#   to add additional apt packages to various docker images
#   - DOCKER_BASE_RT_IMAGE_APT_ADD_PKGS (for all infrastructure containers)
#   - DOCKER_BASE_DEV_IMAGE_APT_ADD_PKGS (for all images which build fabric/fpc code)
#   - DOCKER_DEV_IMAGE_APT_ADD_PKGS (for dev image)


# SGX related settings
#--------------------------------------------------
# (Note: vars are exported as env variables as we also need them in various scripts)
# alternatives for SGX_MODE: SIM or HW
export SGX_MODE ?= SIM
export SGX_BUILD ?= PRERELEASE
export SGX_SSL ?= /opt/intel/sgxssl
export SGX_SDK ?= /opt/intel/sgxsdk
export SGX_ARCH ?= x64


# Settings for other apps
#--------------------------------------------------
# Give the option to override by custom protoc
# e.g. this is overloaded by travis and docker dev as we use protoc 3.11.4 to build
# protos in ecc_enclave but use protoc 3.0.x to build SGX SDK and SSL
export PROTOC_CMD ?= protoc

JAVA ?= java
PLANTUML_JAR ?= plantuml.jar
PLANTUML_CMD ?= $(JAVA) -jar $(PLANTUML_JAR)
PLANTUML_IMG_FORMAT ?= png # pdf / png / svg


# Fabric and FPC related defaults
#--------------------------------------------------
PROJECT_NAME=fabric-private-chaincode

export FABRIC_VERSION ?= 2.3.3

export FPC_VERSION := go-support
export FPC_CCENV_IMAGE ?= hyperledger/fabric-private-chaincode-ccenv:$(FPC_VERSION)

export FABRIC_PATH ?= ${GOPATH}/src/github.com/hyperledger/fabric

export FPC_PATH=$(abspath $(TOP))
# to allow volume mounts from within a dev(elopment) container, 
# below variable is used for volume mounts and can hence be
# re-defined to point to the FPC path as seen by the docker daemon
export DOCKERD_FPC_PATH ?= $(FPC_PATH)

# Fabric binaries are needed for testing; you can customize these via the following
# env variable. By default we fetch the binaries into $(FPC_PATH)/fabric/_internal/bin
# In case you want to use your custom fabric bins, for instance: $(FABRIC_PATH)/build/bin
export FABRIC_BIN_DIR ?= $(FPC_PATH)/fabric/_internal/bin

# Additional SGX related settings
#--------------------------------------------------
export SGX_CREDENTIALS_PATH ?= $(FPC_PATH)/config/ias

# Environment settings
# by default, CI is not running
export IS_CI_RUNNING ?= false

# in CI build, enable code coverage; disable it otherwise
ifeq (${IS_CI_RUNNING}, true)
        #this enables coverage in c/cpp code, libraries and linked binaries
        export CODE_COVERAGE_ENABLED ?= true
else
        export CODE_COVERAGE_ENABLED ?= false
endif

