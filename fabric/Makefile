# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ..
include $(TOP)/build.mk

FABRIC_VERSION ?= 2.1.0
FABRIC_VERSION_TAG =v$(FABRIC_VERSION)
FABRIC_PEER_BIN = $(FABRIC_PATH)/build/bin/peer
FABRIC_TLCC = $(FABRIC_PATH)/core/scc/tlcc

FABRIC_CURRENT_TAG = $(shell cd $(FABRIC_PATH); git tag --points-at HEAD)
FABRIC_PATCHED = $(FABRIC_PATH)/.fpc_patched
FABRIC_TRACKED_CHANGES = $(shell cd $(FABRIC_PATH); git status -s | grep -v '\?')
FABRIC_CURRENT_DIFF_COMMITS_CMD = git log $(FABRIC_VERSION_TAG)..HEAD --oneline

build: native

patch: $(FABRIC_PATCHED)

$(FABRIC_PATCHED):
	@echo "Patching Fabric ..."; \
	if [ "$(FABRIC_CURRENT_TAG)" != "$(FABRIC_VERSION_TAG)" ]; then \
		echo "Aborting! Tag on current HEAD ($(FABRIC_CURRENT_TAG)) does not match expected tag $(FABRIC_VERSION_TAG)!"; \
		exit 1; \
	fi
 
	cd $(FABRIC_PATH) && \
	git am $(FPC_PATH)/fabric/*.patch && \
	$(FABRIC_CURRENT_DIFF_COMMITS_CMD) > $@

clean-patch:
	# Note: for below to work with exit 0, this has to be a single shell command!
	@echo "Cleaning Fabric patches ..."; \
	if [ "$(FABRIC_CURRENT_TAG)" = "$(FABRIC_VERSION_TAG)" ]; then \
		echo "Nothing to do here ... it seems you haven't applied the patches yet"; \
		exit 0; \
	fi; \
	if [ ! -z "$(FABRIC_TRACKED_CHANGES)" ]; then \
		echo "Aborting! It seems you have some changes in your repo:"; \
		echo \\t$(FABRIC_TRACKED_CHANGES); \
		echo "Please clean/stash your changes and try again."; \
		exit 1; \
	fi; \
	if [ ! -f $(FABRIC_PATCHED) ]; then \
		echo "You didn't apply patches automatically. Please manually reset to ${FABRIC_VERSION_TAG}"; \
		exit 1; \
	fi; \
	cd $(FABRIC_PATH); \
	$(FABRIC_CURRENT_DIFF_COMMITS_CMD) > $(FABRIC_PATCHED).tmp; \
	if [ ! -z "$$(diff $(FABRIC_PATCHED) $(FABRIC_PATCHED).tmp)" ]; then \
  		echo "You have other commits than the patches:"; \
		echo "- expected: "; cat $(FABRIC_PATCHED); \
		echo "- found: "; cat $(FABRIC_PATCHED).tmp; \
   		echo "Please check and resolve manually!"; \
   		exit 1; \
  	fi; \
	\
	echo "Removing the following fpc commits in $(FABRIC_PATH)"; \
	$(FABRIC_CURRENT_DIFF_COMMITS_CMD); \
	set -x; \
	cd $(FABRIC_PATH); \
	git reset --hard $(FABRIC_VERSION_TAG) && \
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
