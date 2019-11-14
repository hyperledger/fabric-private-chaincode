# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

FROM hyperledger/fabric-private-chaincode-base:latest

ARG GO_VERSION=go1.13
ARG GO_TAR=$GO_VERSION.linux-amd64.tar.gz
ARG OPENSSL=1.1.0j
ARG SGXSSL=v2.4.1
ARG NANOPB_VERSION=0.3.9.2
ARG FABRIC_REPO=https://github.com/hyperledger/fabric.git
ARG FABRIC_VERSION=v1.4.3

ARG APT_ADD_PKGS=


WORKDIR /tmp

# we need docker commands (even though we use docker daemon from "outside")
RUN apt-get update \
 && apt-get install -y \
    docker.io \
    clang-format \
    $APT_ADD_PKGS

# Install go
RUN wget https://dl.google.com/go/$GO_TAR \
 && tar -xf $GO_TAR \
 && mv go /usr/local \
 && rm $GO_TAR \
 && mkdir -p /projects

ENV GOROOT=/usr/local/go
ENV GOPATH=/project
ENV PATH=$GOPATH/bin:$GOROOT/bin:$PATH

# Install SGX SSL
ENV SGX_SSL /opt/intel/sgxssl
RUN wget https://www.openssl.org/source/openssl-$OPENSSL.tar.gz \
 && git clone  --branch $SGXSSL https://github.com/intel/intel-sgx-ssl.git \
 && . $SGX_SDK/environment \
 && (cd intel-sgx-ssl/openssl_source; mv /tmp/openssl-$OPENSSL.tar.gz . ) \
 && (cd intel-sgx-ssl/Linux; make SGX_MODE=SIM DESTDIR=$SGX_SSL all test ) \
 && (cd intel-sgx-ssl/Linux; make install ) \
 && rm -rf /tmp/intel-sgx-ssl

# Install nanopb
ENV NANOPB_PATH=/usr/local/nanopb/
RUN git clone https://github.com/nanopb/nanopb.git $NANOPB_PATH \
 && cd $NANOPB_PATH \
 && git checkout nanopb-$NANOPB_VERSION \
 && cd generator/proto \
 && make

# Get Fabric (after FPC as we need the FPC patches!)
ENV FABRIC_PATH=$GOPATH/src/github.com/hyperledger/fabric
# Note: could add --single-branch to below to speed-up and keep size smaller. But for now for a dev-image better keep complete repo
# Note: the git config shouldn't really be necessary for a git am but that fails without it.
#   To not commit an id, just do it only temporarily and only locally ...
RUN git clone --branch $FABRIC_VERSION $FABRIC_REPO $FABRIC_PATH \
 && cd $FABRIC_PATH \
 && git config user.email "FPC"

# Go repos needed to build plugins
RUN go get github.com/golang/protobuf/proto

RUN go get github.com/pkg/errors

RUN go get golang.org/x/tools/cmd/goimports

RUN go get golang.org/x/sync/semaphore

# Get FPC
ENV FPC_PATH=$GOPATH/src/github.com/hyperledger-labs/fabric-private-chaincode
ARG FPC_REPO_URL=https://github.com/hyperledger-labs/fabric-private-chaincode.git
ARG FPC_REPO_BRANCH=master

RUN mkdir /tmp/build-src
COPY . /tmp/build-src
# Note: could add --single-branch to below to speed-up and keep size smaller. But for now for a dev-image better keep complete repo
# RUN mkdir -p $FPC_PATH
RUN git clone --branch $FPC_REPO_BRANCH $FPC_REPO_URL $FPC_PATH
# COPY ./ /project/src/github.com/hyperledger-labs/fabric-private-chaincode

RUN cd $FABRIC_PATH \
 && git am $FPC_PATH/fabric/*patch \
 && git config --unset user.email \
 && GO_TAGS=pluginsenabled make peer orderer configtxgen

RUN cd $FPC_PATH \
 && make plugins

RUN cd $FPC_PATH/utils/fabric \
 && make
# RUN apt-get install libtool

ENV FABRIC_BIN_DIR=$FABRIC_PATH/.build/bin
ENV FPC_CMDS=/project/src/github.com/hyperledger-labs/fabric-private-chaincode/fabric/bin
CMD [$FPC_CMDS/peer.sh node start]
