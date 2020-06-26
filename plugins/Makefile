# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

#GOFLAGS :=
#GO := go $(GOFLAGS)

TOP = ..
include $(TOP)/build.mk

FABRIC_FPC_PLUGINS = $(FABRIC_PATH)/fpc_plugins

ECC_VSCC_SRC = ecc-vscc/ecc_validation_plugin.go ecc-vscc/ecc_validation_logic.go
ECC_VSCC_OUT = ecc-vscc.so

ERCC_VSCC_SRC = ercc-vscc/ercc_validation_plugin.go ercc-vscc/ercc_validation_logic.go
ERCC_VSCC_OUT = ercc-vscc.so

ERCC_DECO_SRC = ercc-decorator/decoration.go
ERCC_DECO_OUT = ercc-decorator.so

ATTESTATION_DIR = ../ercc/attestation/
ATTESTATION_PROVIDER = $(ATTESTATION_DIR)/provider.go

$(FABRIC_FPC_PLUGINS):
	ln -sfn $$(/bin/pwd) $(FABRIC_FPC_PLUGINS)

$(ATTESTATION_PROVIDER):
	$(MAKE) -C $(ATTESTATION_DIR)

build: $(FABRIC_FPC_PLUGINS)
	cd $(FABRIC_FPC_PLUGINS) && \
	$(MAKE) build-from-fabric;

.PHONY: build-from-fabric
build-from-fabric: godeps ecc-vscc ercc-vscc ercc-decorator

godeps:
	$(GO) get github.com/hyperledger/fabric-chaincode-go/shim
	$(GO) get github.com/hyperledger/fabric-chaincode-go/shimtest

.PHONY: ecc-vscc
ecc-vscc: $(ECC_VSCC_SRC)
	$(GO) build -o $(ECC_VSCC_OUT) -buildmode=plugin $^

.PHONY: ercc-vscc
ercc-vscc: $(ERCC_VSCC_SRC) $(ATTESTATION_PROVIDER)
	$(GO) build -o $(ERCC_VSCC_OUT) -buildmode=plugin $(ERCC_VSCC_SRC)

.PHONY: ercc-decorator
ercc-decorator: $(ERCC_DECO_SRC)
	$(GO) build -o $(ERCC_DECO_OUT) -buildmode=plugin $^

clean:
	rm -rf *.so $(FABRIC_FPC_PLUGINS)
