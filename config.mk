# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

GOFLAGS :=
GO := go $(GOFLAGS)

DOCKERFLAGS :=
DOCKER := docker $(DOCKERFLAGS)

PROJECT_NAME=fabric-private-chaincode

export FPC_VERSION := cr2.0.0

export SGX_MODE ?= SIM
export SGX_BUILD ?= PRERELEASE
export SGX_SSL ?= /opt/intel/sgxssl
export SGX_SDK ?= /opt/intel/sgxsdk
export SGX_ARCH ?= x64


export FABRIC_PATH ?= ${GOPATH}/src/github.com/hyperledger/fabric

export FPC_PATH=$(abspath $(TOP))
# to allow volume mounts from within a dev(elopment) container, 
# below variable is used for volume mounts and can hence be
# re-defined to point to the FPC path as seen by the docker daemon
export DOCKERD_FPC_PATH ?= $(FPC_PATH)

# Give the option to overload by custom protoc
# e.g. this is overloaded by travis and docker dev as we use protoc 3.11.4 to build
# protos in tlcc_enclave but use protoc 3.0.x to build SGX SDK and SSL
export PROTOC_CMD ?= protoc

JAVA ?= java
PLANTUML_JAR ?= plantuml.jar
PLANTUML_CMD ?= $(JAVA) -jar $(PLANTUML_JAR)
PLANTUML_IMG_FORMAT ?= png # pdf / png / svg
