# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

ARG GO_VER=1.21
ARG ALPINE_VER=3.20
ARG CC_SERVER_PORT=9999

FROM golang:${GO_VER}-alpine${ALPINE_VER}

WORKDIR /go/src/github.com/hyperledger/fabric-private-chaincode/samples/chaincode/cc-tools-demo
COPY . .

RUN go get -d -v .
RUN go build -o cc-tools-demo -v .

EXPOSE ${CC_SERVER_PORT}
CMD ["./cc-tools-demo"]
