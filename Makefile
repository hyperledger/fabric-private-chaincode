# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = .
include $(TOP)/build.mk

SUB_DIRS = utils/fabric-ccenv-sgx ercc ecc_enclave ecc tlcc_enclave tlcc

all build test clean :
	$(foreach DIR, $(SUB_DIRS), $(MAKE) -C $(DIR) $@;)

checks: license linter

license:
	@echo "License: Running licence checks.."
	@${GOPATH}/src/github.com/hyperledger/fabric/scripts/check_license.sh

linter:
	@echo "LINT: Running code checks.."
	@./scripts/golinter.sh
	@./scripts/cpplinter.sh

