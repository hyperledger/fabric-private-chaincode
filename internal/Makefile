# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ..
include $(TOP)/build.mk

# this is needed to run the crypto compatibility test for the go-crypto impl
GOTAGS += -tags WITH_PDO_CRYPTO

test:
	$(GO) test $(GOTAGS) $(GOTESTFLAGS) ./...
