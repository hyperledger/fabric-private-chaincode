# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../../..
include $(TOP)/build.mk

CC_ID ?= echo
GO_BUILD_OPT ?=

build:
	$(GO) build $(GO_BUILD_OPT)

clean:
	$(GO) clean
	rm -rf keystore wallet

run:
	env CC_ID=${CC_ID} $(GO) run .
