# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ..
include $(TOP)/build.mk

LD_LIBRARY_PATH := LD_LIBRARY_PATH=${LD_LIBRARY_PATH}:$$(pwd)/enclave/lib
LOGGING_SPEC := FABRIC_LOGGING_SPEC=tlcc=debug
TEST_ENV := ${LD_LIBRARY_PATH} ${LOGGING_SPEC}
FABRIC_TLCC_PATH := ${FABRIC_PATH}/core/scc/tlcc

test:
	@if [ ! -d "${FABRIC_TLCC_PATH}" ]; then \
  		echo "Aborting! TLCC is not yet linked into Fabric"; \
  		echo "try to run \"cd \$$FPC_PATH/fabric && make\""; \
  		exit 1; \
  	fi
	cd ${FABRIC_TLCC_PATH} && $(TEST_ENV) $(GO) $@ -v ./...

clean:
	$(GO) clean
	rm -rf ledgersData
