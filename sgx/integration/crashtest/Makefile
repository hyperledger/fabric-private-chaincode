# SPDX-License-Identifier: Apache-2.0

TOP = ../..
include $(TOP)/build.mk

GO_TEST_DIRS=unmarshal_values

test:
	$(foreach DIR, $(GO_TEST_DIRS), $(MAKE) -C $(DIR) || exit ;)

