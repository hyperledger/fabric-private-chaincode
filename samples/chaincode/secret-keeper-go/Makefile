# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../../..

CC_NAME ?= fpc-secret-keeper-go

# Define paths for cmd subdirectories
DEFAULT= cmd/naive/main.go
SKVS_PATH = cmd/skvs/main.go

ECC_MAIN_FILES ?=$(DEFAULT)

include $(TOP)/ecc_go/build.mk
