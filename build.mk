# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

include $(TOP)/config.mk

# optionally allow local overriding defaults
-include $(TOP)/config.override.mk

# define composites only here and not in config.mk so we can override parts in config.override.mk
DOCKER := $(DOCKER_CMD) $(DOCKERFLAGS)
DOCKER_COMPOSE := $(DOCKER_CMD) compose

ifeq (${SGX_MODE}, HW)
	GOTAGS = -tags sgx_hw_mode
endif
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
