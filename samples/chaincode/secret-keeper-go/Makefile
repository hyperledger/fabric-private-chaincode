# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../../..
include $(TOP)/ecc_go/build.mk

CC_NAME ?= fpc-secret-keeper-go

# Define paths for cmd subdirectories
NAIVE_PATH = cmd/naive/main.go
SKVS_PATH = cmd/skvs/main.go

# Target to build naive version
naive:
	$(MAKE) build ECC_MAIN_FILES=$(NAIVE_PATH)

# Target to build skvs version
skvs:
	$(MAKE) build ECC_MAIN_FILES=$(SKVS_PATH)
