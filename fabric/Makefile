# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ..
include $(TOP)/build.mk

FABRIC_VERSION ?= 2.1.1
FABRIC_VERSION_TAG =v$(FABRIC_VERSION)
FABRIC_PEER_BIN = $(FABRIC_PATH)/build/bin/peer
FABRIC_TLCC = $(FABRIC_PATH)/core/scc/tlcc
FABRIC_VALIDATION = $(FABRIC_PATH)/core/handlers/validation/fpc
TLCC_GO_SRC = ${TOP}/tlcc/*.go
TLCC_ENCLAVE_DIR = ${TOP}/tlcc_enclave
TLCC_ENCLAVE_LIB = ${TLCC_ENCLAVE_DIR}/_build/libtl.so

FABRIC_PATCHED = $(FABRIC_PATH)/.fpc_patched
FABRIC_CURRENT_DIFF_COMMITS_CMD = git log $(FABRIC_VERSION_TAG)..HEAD --oneline

build: native

patch: $(FABRIC_PATCHED)

$(FABRIC_PATCHED): *.patch $(FABRIC_PATH)/.git
	# Note: .git dependency hopefully captures significant changes in fabric dir requiring repatch (and rebuild)
	@echo "Patching Fabric ..."; \
	if [ -f $(FABRIC_PATCHED) ]; then \
		echo "Need to re-apply patch, so try to undo patches first"; \
		$(MAKE) clean-patch; \
	fi; \
	FABRIC_CURRENT_TAG=$$(cd "$(FABRIC_PATH)"; git tag --points-at HEAD) || exit 2; \
	if [ "$${FABRIC_CURRENT_TAG}" != "$(FABRIC_VERSION_TAG)" ]; then \
		echo "Aborting! Tag on current HEAD ($${FABRIC_CURRENT_TAG}) does not match expected tag $(FABRIC_VERSION_TAG)!"; \
		exit 1; \
	fi
	cd $(FABRIC_PATH) && \
	git am $(FPC_PATH)/fabric/*.patch && \
	$(FABRIC_CURRENT_DIFF_COMMITS_CMD) > $@

clean-patch:
	# Note: for below to work with exit 0, this has to be a single shell command!
	@echo "Cleaning Fabric patches ..."; \
	FABRIC_CURRENT_TAG=$$(cd "$(FABRIC_PATH)"; git tag --points-at HEAD) || exit 2; \
	FABRIC_TRACKED_CHANGES=$$(cd "$(FABRIC_PATH)"; git status -s | grep -v '\?'); if [ $$? = 2 ]; then exit 3; fi; \
	if [ "$${FABRIC_CURRENT_TAG}" = "$(FABRIC_VERSION_TAG)" ]; then \
		echo "Nothing to do here ... it seems you haven't applied the patches yet"; \
		exit 0; \
	fi; \
	if [ ! -z "$${FABRIC_TRACKED_CHANGES}" ]; then \
		echo "Aborting! It seems you have some changes in your repo:"; \
		echo \\t$${FABRIC_TRACKED_CHANGES}; \
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

$(FABRIC_VALIDATION):
	# TODO: this will go away once upgraded to Fabric v2.2 has completed
	ln -sfn $$(/bin/pwd)/../validation/fpc $(FABRIC_VALIDATION)
	ln -sfn $$(/bin/pwd)/../validation/ercc $(FABRIC_VALIDATION)/ercc

$(TLCC_ENCLAVE_LIB):
	${MAKE} -C ${TLCC_ENCLAVE_DIR} build

peer: patch $(FABRIC_PEER_BIN)

$(FABRIC_PEER_BIN): $(FABRIC_TLCC) $(FABRIC_VALIDATION) $(TLCC_GO_SRC) $(TLCC_ENCLAVE_LIB) $(FABRIC_PATCHED)
	# Note: above does not have enough dependencies to trigger a rebuild if (recursively)
	# any TLCC and dependent code changes. however, it should at least make sure it will
	# build properly if tlcc enclave is not yet built ..
	cd $(FABRIC_PATH) && \
	GO_TAGS=pluginsenabled,FPC $(MAKE) peer

native: peer
	cd $(FABRIC_PATH) && \
	$(MAKE) -j orderer cryptogen configtxgen

clean: clean-native clean-patch
	rm -f $(FABRIC_TLCC)
	rm -f $(FABRIC_VALIDATION)

clean-peer:
	rm -rf $(FABRIC_PEER_BIN)

clean-native:
	rm -rf $(FABRIC_PATH)/build/bin/
