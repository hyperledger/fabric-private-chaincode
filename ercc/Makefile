# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ..
include $(TOP)/build.mk

VSCC_SRC = vscc/ercc_validation_plugin.go vscc/ercc_validation_logic.go
VSCC_OUT = ercc-vscc.so
DECO_SRC = attestation/ias_credentials/decoration.go
DECO_OUT = ercc-decorator.so

# TODO: re-enable/port plugins
build: ias-keys provider ercc # vscc-plugin decorator-plugin

.PHONY: .sgx_mode

IAS_GO_DEFINES=

ifeq (${SGX_MODE}, HW)
IAS_GO_DEFINES+=-DUSE_SGX_HARDWARE_MODE
endif

TEMPLATE_SUFFIX=.template
GO_TARGETS=provider.go
GO_TEMPLATES=$(addsuffix $(TEMPLATE_SUFFIX), $(GO_TARGETS))

# put a dependency on the SGX_MODE environment variable
ifneq "$(SGX_MODE)" "$(shell cat .sgx_mode 2>/dev/null)"
SGX_MODE_DEPENDENCY=.sgx_mode
else
SGX_MODE_DEPENDENCY=
endif

ias-keys:
ifeq (${SGX_MODE}, SIM)
	if [ ! -f ../config/ias/api_key.txt ]; then\
            touch ../config/ias/api_key.txt;\
	fi;
	if [ ! -f ../config/ias/spid.txt ]; then\
            echo "DEADBEEF00000000DEADBEEF00000000" > ../config/ias/spid.txt;\
	fi;
endif

provider: $(SGX_MODE_DEPENDENCY) $(GO_TARGETS)

$(GO_TARGETS): $(SGX_MODE_DEPENDENCY) $(GO_TEMPLATES)
	$(CC) -E -P $(IAS_GO_DEFINES) - < $@$(TEMPLATE_SUFFIX) > $@
	gofmt -w $@ && goimports -w $@

ercc:
	$(GO) build
	rm ercc # validating a successful build, ercc not meant to be run directly

vscc-plugin: $(VSCC_SRC)
	$(GO) build -o $(VSCC_OUT) -buildmode=plugin $^

decorator-plugin: $(DECO_SRC)
	$(GO) build -o $(DECO_OUT) -buildmode=plugin $^

test: build
	$(GO) test -v ./...

clean: docker-clean
	$(GO) $@
	rm -rf $(GO_TARGETS) .sgx_mode

.sgx_mode:
	@echo $(SGX_MODE) > $@

ERCC_NAME?=ercc
DOCKER_IMAGE?=$$(docker images | grep -- -$(ERCC_NAME)- | awk '{print $$1;}')

docker-clean:
	-if [ ! -z "$(DOCKER_IMAGE)" ]; then docker rmi -f $(DOCKER_IMAGE); fi
	-if [ ! -z "$(DOCKER_IMAGE)" ]; then echo "\033[0;33mWARNING: ercc not deleted (peer still up?)\033[0m"; fi
