# Copyright 2021 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../../../../../..
include $(TOP)/build.mk

all: build

build:
	go build

clean:
	go clean
	rm -rf keystore wallet

run:
	./test
