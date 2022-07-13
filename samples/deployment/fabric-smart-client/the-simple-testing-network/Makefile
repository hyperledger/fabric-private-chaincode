# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

build:
	go build -o tstn

clean:
	rm -rf testdata

fetch-bins:
	curl -sSL https://bit.ly/2ysbOFE | bash -s -- 2.2.7 1.4.9 -s -d

run-start:
	./tstn network start --path ./testdata

