# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

FABRIC_LOGGING_SPEC?="fpc=debug:endorser,grpc,comm.grpc,gossip=error:info"

build:
	go build -o tstn

clean:
	rm -rf testdata

run: clean
	FABRIC_LOGGING_SPEC=$(FABRIC_LOGGING_SPEC) go run . network start --path ./testdata

