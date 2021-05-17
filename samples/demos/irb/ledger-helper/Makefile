# Copyright 2021 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

all: build

build:
	go build -o register_chaincode

clean:
	go clean
	rm -rf keystore

run:
	./start-ledger.sh
	./register_chaincode

stop:
	./stop-ledger.sh
	rm -rf keystore
