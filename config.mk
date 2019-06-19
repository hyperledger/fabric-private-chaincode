# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

GOFLAGS :=
GO := go $(GOFLAGS)

DOCKERFLAGS :=
DOCKER := docker $(DOCKERFLAGS)

export SGX_MODE ?= HW
export SGX_BUILD ?= PRERELEASE

export FABRIC_PATH ?= ${GOPATH}/src/github.com/hyperledger/fabric
