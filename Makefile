# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = .
include $(TOP)/build.mk

SUB_DIRS = utils/fabric-ccenv-sgx ercc ecc_enclave ecc tlcc_enclave tlcc integration

all build test clean :
	$(foreach DIR, $(SUB_DIRS), $(MAKE) -C $(DIR) $@ || exit;)

checks: license linter

license:
	@echo "License: Running licence checks.."
	@${GOPATH}/src/github.com/hyperledger/fabric/scripts/check_license.sh

linter: gotools build
	@echo "LINT: Running code checks.."
	@cd $$(/bin/pwd) && ./scripts/golinter.sh
	@cd $$(/bin/pwd) && ./scripts/cpplinter.sh

gotools:
	# install goimports if not present
	$(GO) get golang.org/x/tools/cmd/goimports
