# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

TOP = ..
include $(TOP)/build.mk

all: test

test: auction_test echo_test kv_test deployment_test client_sdk_test stress_test ecc_go_test

auction_test:
	./auction_test.sh

echo_test:
	./echo_test.sh

kv_test:
	./kv_test.sh

deployment_test:
	./deployment_test.sh

client_sdk_test:
	./client_sdk_test.sh

stress_test:
	./stress_test.sh

ecc_go_test:
	$(MAKE) -C go_chaincode
