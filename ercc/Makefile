# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ..
include $(TOP)/build.mk

build: ercc

ercc:
	# ERCC's binary is created here.
	# The binary is then referenced by the fpc-peer cli, inserted in the ERCC package,
	# and eventually run by the external launcher.
	$(GO) build $(GOTAGS) -o ercc main.go

test: build
	$(GO) test $(GOTAGS) -v ./...

clean:
	$(GO) clean

package:
	tar cfz code.tar.gz connection.json
	tar cfz ercc.tgz metadata.json code.tar.gz
	rm code.tar.gz

docker:
	docker build --no-cache -t fpc/ercc .

run:
	docker run -it --rm --name ercc.org1.example.com --hostname ercc.org1.example.com --env-file chaincode.env --network=net_test fpc/ercc
