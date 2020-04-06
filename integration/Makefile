# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

TOP = ..
include $(TOP)/build.mk

all: test

test: auction_test echo_test deployment_test

auction_test:
	./auction_test.sh

echo_test:
	./echo_test.sh

deployment_test:
	./deployment_test.sh
