# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

include $(TOP)/config.mk

# optionlly allow local overriding defaults
-include $(TOP)/config.override.mk

.PHONY: all
all: build test checks # keep checks last as license test is brittle ...

.PHONY: test
test: build

.PHONY: build
.PHONY: checks
.PHONY: clean

.PHONY: clobber
clobber: clean

.PHONY: docker
.PHONY: docker-run
.PHONY: docker-stop
.PHONY: docker-clean
