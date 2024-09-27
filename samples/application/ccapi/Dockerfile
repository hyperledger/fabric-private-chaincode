# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

ARG FPC_VERSION=main

FROM hyperledger/fabric-private-chaincode-ccenv:${FPC_VERSION}

COPY fpcclient /usr/local/bin

WORKDIR /opt/gopath/src/github.com/hyperledger/fabric/peer
