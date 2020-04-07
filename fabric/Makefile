# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ..
include $(TOP)/build.mk

FABRIC_VERSION ?= v2.0.1
FABRIC_PEER_BIN = $(FABRIC_PATH)/build/bin/peer
FABRIC_TLCC = $(FABRIC_PATH)/core/scc/tlcc

FABRIC_CURRENT_BRANCH = $(shell cd $(FABRIC_PATH); git rev-parse --abbrev-ref HEAD)
FABRIC_CURRENT_TAG = $(shell cd $(FABRIC_PATH); git describe --tags)
FABRIC_PATCHED_BRANCH = $(FABRIC_VERSION)-fpc
FABRIC_PATCHED = $(FABRIC_PATH)/.fpc_patched
FABRIC_TRACKED_CHANGES = $(shell cd $(FABRIC_PATH); git status -s | grep -v '\?')
FABRIC_COMMITS = $(shell cd $(FABRIC_PATH); git log $(FABRIC_VERSION)..HEAD --oneline)
FABRIC_LOG_DIFF = $(shell diff $(FABRIC_PATCHED) <(FABRIC_COMMITS))

build: native

patch: $(FABRIC_PATCHED)

$(FABRIC_PATCHED):
	@echo "Patching Fabric ..."
	@if [ "$(FABRIC_CURRENT_BRANCH)" != "$(FABRIC_PATCHED_BRANCH)" ]; then \
		echo "Aborting! Current branch is $(FABRIC_CURRENT_BRANCH) but want $(FABRIC_PATCHED_BRANCH)"; exit 1; \
	fi

	@if [ "$(FABRIC_CURRENT_TAG)" != "$(FABRIC_VERSION)" ]; then \
		echo "Aborting! Current branch HEAD does not match tag/$(FABRIC_VERSION)! $(FABRIC_CURRENT_TAG) but want $(FABRIC_VERSION)"; exit 1; \
	fi

	cd $(FABRIC_PATH) && \
	git am $(FPC_PATH)/fabric/*.patch && \
	git log $(FABRIC_VERSION)..HEAD --oneline > $@

clean-patch:
	@echo "Cleaning Fabric patches ..."
	@if [ "$(FABRIC_CURRENT_BRANCH)" != "$(FABRIC_PATCHED_BRANCH)" ]; then \
		echo "Aborting! Current branch is $(FABRIC_CURRENT_BRANCH) but want $(FABRIC_PATCHED_BRANCH)"; \
		echo "Please checkout $(FABRIC_PATCHED_BRANCH) and come back."; \
		exit 1; \
	fi

	@if [ "$(FABRIC_CURRENT_TAG)" = "$(FABRIC_VERSION)" ]; then \
		echo "Aborting! Nothing to do here ... it seems you haven't applied the patches yet"; exit 1; \
	fi

	@if [ ! -z "$(FABRIC_TRACKED_CHANGES)" ]; then \
		echo "Aborting! It seems you have some changes in your branch:"; \
		echo \\t$(FABRIC_TRACKED_CHANGES); \
		echo "Please clean/stash your changes and try again."; \
		exit 1; \
	fi

	@cd $(FABRIC_PATH) && git log $(FABRIC_VERSION)..HEAD --oneline > $(FABRIC_PATCHED).tmp
	@if [ "$(diff $(FABRIC_PATCHED) $(FABRIC_PATCHED).tmp)" ]; then \
  		echo "You have other commits than the patches:"; \
   		cd $(FABRIC_PATH) && git log $(FABRIC_VERSION)..HEAD --oneline; \
   		echo "Please check and resolve manually!"; \
   		exit 1; \
  	fi

	@echo "Removing the following commits on branch ${FABRIC_PATCHED_BRANCH} in $(FABRIC_PATH)"
	@cd $(FABRIC_PATH) && git log $(FABRIC_VERSION)..HEAD --oneline

	cd $(FABRIC_PATH) && \
	git reset --hard $(FABRIC_VERSION)
	rm -f $(FABRIC_PATCHED) $(FABRIC_PATCHED).tmp

$(FABRIC_TLCC):
	ln -sfn $$(/bin/pwd)/../tlcc $(FABRIC_TLCC)

peer: patch $(FABRIC_PEER_BIN)

$(FABRIC_PEER_BIN): $(FABRIC_TLCC)
	cd $(FABRIC_PATH) && \
	GO_TAGS=pluginsenabled $(MAKE) peer

native: peer
	cd $(FABRIC_PATH) && \
	$(MAKE) -j orderer cryptogen configtxgen

clean: clean-native clean-patch
	rm -f $(FABRIC_TLCC)

clean-peer:
	rm -rf $(FABRIC_PEER_BIN)

clean-native:
	rm -rf $(FABRIC_PATH)/build/bin/
