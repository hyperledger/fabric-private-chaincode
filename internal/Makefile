# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ..
include $(TOP)/build.mk

# with FPC_CGO=1 (fpc_cgo), the tests also cover the PDO crypto and its
# compatibility with the go crypto implementation
test:
	$(GO) test $(GOTAGS) $(GOTESTFLAGS) ./...
