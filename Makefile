SUB_DIRS = utils/fabric-ccenv-sgx ercc ecc_enclave ecc tlcc_enclave tlcc

all clean test:
	$(foreach DIR, $(SUB_DIRS), $(MAKE) -C $(DIR) $@;)
