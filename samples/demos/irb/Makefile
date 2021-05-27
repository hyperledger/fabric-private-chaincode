# Copyright 2021 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../../..
include $(TOP)/build.mk

SUB_DIRS = protos experimenter experiment-approval-service storage-service data-provider ledger-helper principal-investigator

build clean:
	$(foreach DIR, $(SUB_DIRS), $(MAKE) -C $(DIR) $@ || exit;)

test:
	(cd test && ./orchestrate.sh)
