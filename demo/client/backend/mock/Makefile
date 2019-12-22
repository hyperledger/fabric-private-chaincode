# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../../../..
include $(TOP)/build.mk

CC_NAME = auction_mock_chaincode
DOCKER_IMAGE = auction_mock_backend
APP_SRC = ${PWD}

run-go:
	$(GO) run .

run-fpc:
	$(GO) build -tags fpc
	./mock --debug
