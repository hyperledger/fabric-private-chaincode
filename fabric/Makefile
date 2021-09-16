# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ..
include $(TOP)/build.mk

FABRIC_VERSION_TAG =v$(FABRIC_VERSION)
FABRIC_PEER_BIN = $(FABRIC_PATH)/build/bin/peer

FABRIC_PATCHED = $(FABRIC_PATH)/.fpc_patched
FABRIC_CURRENT_DIFF_COMMITS_CMD = git log $(FABRIC_VERSION_TAG)..HEAD --oneline

FABS = _internal
FABS_FETCHED = $(FABS)/.fetched

build: fetch

fetch: $(FABS_FETCHED)

$(FABS_FETCHED):
	mkdir -p $(FABS) && cd $(FABS) \
	&& curl -sSL https://bit.ly/2ysbOFE | bash -s -- $(FABRIC_VERSION) 1.4.9 -s \
	&& touch .fetched


# this target applies all fabric patches in the patches folder;
# these patches are not required for FPC but may be helpful for 
# developers testing (e.g., residing behind a proxy environment)
#
# optional target
patch: $(FABRIC_PATCHED)

$(FABRIC_PATCHED): patches/*.patch $(FABRIC_PATH)/.git
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
	git am $(FPC_PATH)/fabric/patches/*.patch && \
	$(FABRIC_CURRENT_DIFF_COMMITS_CMD) > $@ && \
	$(GO) mod vendor

clean-patch:
	# Note: for below to work with exit 0, this has to be a single shell command!
	@echo "Cleaning Fabric patches ..."; \
	echo "(Warning: as go.{mod,sum} and  vendor/ are auto-updated, we also (blindly) auto-reset them to committed version!)"; \
        (cd $(FABRIC_PATH) && git checkout -- go.mod go.sum && git checkout HEAD -- vendor && git clean -fd vendor) || exit 1; \
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

peer: patch $(FABRIC_PEER_BIN)

$(FABRIC_PEER_BIN): $(FABRIC_PATCHED)
	cd $(FABRIC_PATH) && \
        $(MAKE) peer

# this target builds custom fabric binaries in the provided FABRIC_PATH
# 
# optional target
native: peer
	cd $(FABRIC_PATH) && \
	$(MAKE) -j orderer cryptogen configtxgen

clean:

clobber: clean-fetched

clean-peer:
	rm -rf $(FABRIC_PEER_BIN)

clean-native:
	rm -rf $(FABRIC_PATH)/build/bin/

clean-fetched:
	rm -rf $(FABS)
