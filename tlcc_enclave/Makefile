TOP = ..
include $(TOP)/build.mk

PROTO_BUILD_DIR := enclave/protos
BUILD_DIR := _build

.PHONY: protos
protos: $(PROTO_BUILD_DIR)

$(PROTO_BUILD_DIR):
	./generate_protos.sh $@

.PHONY: protos-clean
protos-clean:
	-rm -rf $(PROTO_BUILD_DIR)

$(BUILD_DIR): protos
	mkdir -p $(BUILD_DIR) && \
		cd $(BUILD_DIR) && \
		cmake ./..

build: $(BUILD_DIR) protos
	$(MAKE) all --directory=$<

.PHONY: enclave-clean
enclave-clean:
	-rm -rf $(BUILD_DIR)
	-rm enclave/*_t.*
	-rm trusted_ledger/*_u.*

clean: enclave-clean protos-clean
