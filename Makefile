TOP = .
include $(TOP)/build.mk

SUB_DIRS = utils/fabric-ccenv-sgx ercc ecc_enclave ecc tlcc_enclave tlcc

all build test clean :
	$(foreach DIR, $(SUB_DIRS), $(MAKE) -C $(DIR) $@;)

checks: linter

linter:
	@echo "LINT: Running code checks.."
	@./scripts/golinter.sh
	@./scripts/cpplinter.sh

