# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

FABRIC_LOGGING_SPECS="fpc=debug:grpc=error:comm.grpc=error:gossip=warning:info"

build:
	go build -o tstn

clean:
	rm -rf testdata

run: clean
	FABRIC_LOGGING_SPECS=$(FABRIC_LOGGING_SPECS) go run . network start --path ./testdata

