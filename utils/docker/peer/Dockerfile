# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

ARG FPC_VERSION=latest

FROM hyperledger/fabric-private-chaincode-base:${FPC_VERSION}

ARG GO_VERSION=go1.13
ARG GO_TAR=$GO_VERSION.linux-amd64.tar.gz
ARG NANOPB_VERSION=0.3.9.2
ARG FABRIC_REPO=https://github.com/hyperledger/fabric.git
ARG FABRIC_VERSION=1.4.3
ARG FPC_REPO_URL=https://github.com/hyperledger-labs/fabric-private-chaincode.git
ARG FPC_REPO_BRANCH_TAG_OR_COMMIT=master
ARG OPENSSL_VERSION=1.1.0j
ARG SGXSSL_VERSION=v2.4.1
ARG SGX_MODE=SIM

ARG APT_ADD_PKGS=

ENV NANOPB_VERSION=${NANOPB_VERSION}
ENV FABRIC_VERSION=${FABRIC_VERSION}
ENV OPENSSL_VERSION=${OPENSSL_VERSION}
ENV SGXSSL_VERSION=${SGXSSL_VERSION}
ENV SGX_MODE=${SGX_MODE}


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
 && mkdir -p /project

ENV GOROOT=/usr/local/go
ENV GOPATH=/project
ENV PATH=$GOPATH/bin:$GOROOT/bin:$PATH

# Install SGX SSL
ENV SGX_SSL /opt/intel/sgxssl
RUN wget https://www.openssl.org/source/openssl-${OPENSSL_VERSION}.tar.gz \
 && git clone  --branch ${SGXSSL_VERSION} https://github.com/intel/intel-sgx-ssl.git \
 && . $SGX_SDK/environment \
 && (cd intel-sgx-ssl/openssl_source; mv /tmp/openssl-${OPENSSL_VERSION}.tar.gz . ) \
 && (cd intel-sgx-ssl/Linux; make SGX_MODE=SIM DESTDIR=$SGX_SSL all test ) \
 # Note: in docker build we cannot run with SGX_MODE=HW, so run as SIM.
 # However, compiled libraries can still be used in both modes!
 && (cd intel-sgx-ssl/Linux; make install ) \
 && rm -rf /tmp/intel-sgx-ssl

# Install nanopb
ENV NANOPB_PATH=/usr/local/nanopb/
RUN git clone https://github.com/nanopb/nanopb.git $NANOPB_PATH \
 && cd $NANOPB_PATH \
 && git checkout nanopb-$NANOPB_VERSION \
 && cd generator/proto \
 && make

# Go repos we need (and pre-load them here rather than later in make)
RUN go get github.com/golang/protobuf/proto \
 && go get github.com/pkg/errors \
 && go get golang.org/x/tools/cmd/goimports \
 && go get golang.org/x/sync/semaphore \
 && go get github.com/spf13/viper

# Get Fabric
ENV FABRIC_PATH=$GOPATH/src/github.com/hyperledger/fabric
RUN git clone --branch v$FABRIC_VERSION $FABRIC_REPO $FABRIC_PATH
# Note: could add --single-branch to below to speed-up and keep size smaller. But for now for a dev-image better keep complete repo

# Get FPC
ENV FPC_PATH=$GOPATH/src/github.com/hyperledger-labs/fabric-private-chaincode
# We copy context so we can use that to potentially get local .git as repo ...
COPY .git /tmp/cloned-local-fpc-git-repo
RUN git clone $FPC_REPO_URL $FPC_PATH \
  && cd $FPC_PATH \
  && git checkout $FPC_REPO_BRANCH_TAG_OR_COMMIT
# Note: could add --single-branch to below to speed-up and keep size smaller. But for now for a dev-image better keep complete repo

# Build Fabric (after FPC as we need the FPC patches!)
# Note: the git config shouldn't really be necessary for a git am but that fails without it.
#   To not commit an id, just do it only temporarily and only locally ...
RUN cd $FABRIC_PATH \
 && git config user.email "FPC" \
 && git am $FPC_PATH/fabric/*patch \
 && git config --unset user.email \
 && GO_TAGS=pluginsenabled make peer orderer cryptogen configtxgen

# Build FPC plugins for peer
RUN cd $FPC_PATH \
 && make plugins \
 && make -C $FPC_PATH/utils/fabric 

ENV FABRIC_BIN_DIR=$FABRIC_PATH/.build/bin
ENV FPC_CMDS=/project/src/github.com/hyperledger-labs/fabric-private-chaincode/fabric/bin

CMD [$FPC_CMDS/peer.sh node start]
