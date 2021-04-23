# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

include $(TOP)/config.mk

# optionlly allow local overriding defaults
-include $(TOP)/config.override.mk

# define composites only here and not in config.mk so we can override parts in config.override.mk
DOCKER := DOCKER_BUILDKIT=$(DOCKER_BUILDKIT) $(DOCKER_CMD) $(DOCKERFLAGS)
ifeq (${SGX_MODE}, HW)
	GOTAGS = -tags sgx_hw_mode
endif
GO := $(GO_CMD) $(GOFLAGS)

.PHONY: all
# if we run CI, we add some additional targets on the default make
ifeq (${IS_CI_RUNNING}, true)
all: build test docker ci_report checks # keep checks last as license test is brittle ...
else
all: build test checks # keep checks last as license test is brittle ...
endif

.PHONY: ci_report

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
