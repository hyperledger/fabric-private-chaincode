# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

include $(TOP)/config.mk

# optionally allow local overriding defaults
-include $(TOP)/config.override.mk

# define composites only here and not in config.mk so we can override parts in config.override.mk
DOCKER := $(DOCKER_CMD) $(DOCKERFLAGS)
DOCKER_COMPOSE := $(DOCKER_CMD) compose

# Go build tags. Makefiles add tags with `GO_TAG_LIST += <tag>`; GOTAGS passes
# them as one comma-separated -tags flag, since go honours only the last -tags.
# fpc_cgo: the C side (cgo against FPC's C code) is built, see FPC_CGO.
GO_TAG_LIST += $(if $(filter 1,$(FPC_CGO)),fpc_cgo)
GO_TAG_LIST += $(if $(filter HW,$(SGX_MODE)),sgx_hw_mode)
comma := ,
empty :=
space := $(empty) $(empty)
GOTAGS = $(if $(strip $(GO_TAG_LIST)),-tags $(subst $(space),$(comma),$(strip $(GO_TAG_LIST))))
GO := $(GO_CMD) $(GOFLAGS)

GOTESTFLAGS := -v -race -covermode=atomic -coverprofile=coverage.out

.PHONY: all
all: build test checks # keep checks last as license test is brittle ...

.PHONY: build
.PHONY: test
.PHONY: checks
.PHONY: clean

.PHONY: clobber
clobber: clean

.PHONY: docker
.PHONY: docker-run
.PHONY: docker-stop
.PHONY: docker-clean
