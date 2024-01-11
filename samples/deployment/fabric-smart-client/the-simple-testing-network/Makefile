# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
TOP = ../../../..
include $(TOP)/build.mk

FABRIC_LOGGING_SPEC?="fpc=debug:endorser,grpc,comm.grpc,gossip=error:info"

build:
	go build -o tstn

images: 
	$(DOCKER) pull hyperledger/fabric-ccenv:2.5
	$(DOCKER) image tag hyperledger/fabric-ccenv:2.5 hyperledger/fabric-ccenv:latest
	$(DOCKER) pull hyperledger/explorer:latest
	$(DOCKER) pull hyperledger/explorer-db:latest

clean:
	rm -rf testdata

run: clean
	FABRIC_LOGGING_SPEC=$(FABRIC_LOGGING_SPEC) go run . network start --path ./testdata

