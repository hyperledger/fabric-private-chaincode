# Copyright IBM Corp. All Rights Reserved.
# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

# Description:
#   Builds the environment with all prerequistes needed to _build_ SGX-enabled apps as needed in FPC
#
#  Configuration (build) paramaters (for defaults, see below section with ARGs)
#  - fpc image version:         FPC_VERSION
#  - go version:                GO_VERSION
#  - nanopb version:            NANOPB_VERSION
#  - openssl version:           OPENSSL
#  - sgxssl version:            SGXSSL
#  - additional apt pkgs:       APT_ADD_PKGS


ARG FPC_VERSION=main

FROM hyperledger/fabric-private-chaincode-base-rt:${FPC_VERSION} as common
LABEL org.opencontainers.image.source https://github.com/hyperledger/fabric-private-chaincode

# config/build params
ARG GO_VERSION=1.17.5
ARG NANOPB_VERSION=0.4.3
ARG OPENSSL=1.1.1g
ARG SGXSSL=2.10_1.1.1g
ARG APT_ADD_PKGS=
ARG EGO_VERSION=1.0.0

# for convenience remember all versions as env variables ..
ENV GO_VERSION=${GO_VERSION}
ENV NANOPB_VERSION=${NANOPB_VERSION}
ENV OPENSSL_VERSION=${OPENSSL_VERSION}
ENV SGXSSL_VERSION=${SGXSSL_VERSION}


WORKDIR /tmp

RUN apt-get update -q \
  && env DEBIAN_FRONTEND="noninteractive" TZ="UTC" \
  # above makes sure any install of 'tzdata' or alike (as e.g., pulled in via ubuntu 20.04) does not hang ...
    apt-get install -y -q \
    # build tools
    build-essential \
    clang-format \
    cmake \
    git \
    libcurl4-openssl-dev \
    libprotobuf-dev \
    libssl-dev \
    libtool \
    pkg-config \
    python \ 
    protobuf-compiler \
    python-protobuf \
    # docker commands (need as we use docker daemon from "outside")
    docker.io \
    docker-compose \
    psmisc \
    bc \
    ${APT_ADD_PKGS}


# Install go
ENV GOROOT=/usr/local/go
RUN GO_TAR=go${GO_VERSION}.linux-amd64.tar.gz \
  && wget -q https://dl.google.com/go/${GO_TAR} \
  && tar -xf ${GO_TAR} \
  && mv go /usr/local \
  && rm ${GO_TAR} \
  && mkdir -p /project
ENV PATH=${GOPATH}/bin:${GOROOT}/bin:${PATH}

#  Go tools we need
RUN  go get golang.org/x/tools/cmd/goimports \
  && go get google.golang.org/protobuf/cmd/protoc-gen-go \
  && GO111MODULE=off go get github.com/maxbrunsfeld/counterfeiter \
  && go get honnef.co/go/tools/cmd/staticcheck@v0.3.3 \
  && go get github.com/client9/misspell/cmd/misspell \
  && GO111MODULE=on go get github.com/mikefarah/yq/v3

# Install SGX SSL
ENV SGX_SSL /opt/intel/sgxssl
RUN git clone 'https://github.com/intel/intel-sgx-ssl.git' \
  && cd intel-sgx-ssl \
  && . /opt/intel/sgxsdk/environment \
  && git checkout lin_${SGXSSL} \
  && cd openssl_source \
  && wget -q https://www.openssl.org/source/openssl-${OPENSSL}.tar.gz \
  && cd ../Linux \
  && make SGX_MODE=SIM DESTDIR=${SGX_SSL} all test \
  # Note: we need explicitly set to SIM as default is HW yet during docker
  # build you never have access to HW, regardless of platform. Note, though,
  # that libraries built work for both modes, on tests are executed for SIM only.
  && make install

# Install nanopb
ENV NANOPB_PATH=/usr/local/nanopb/
RUN git clone https://github.com/nanopb/nanopb.git ${NANOPB_PATH} \
  && cd ${NANOPB_PATH} \
  && git checkout nanopb-${NANOPB_VERSION} \
  && cd generator/proto \
  && make

# Install ego
RUN wget https://github.com/edgelesssys/ego/releases/download/v${EGO_VERSION}/ego_${EGO_VERSION}_amd64.deb \
  && apt install -y -q ./ego_${EGO_VERSION}_amd64.deb \
  && rm ego_${EGO_VERSION}_amd64.deb
