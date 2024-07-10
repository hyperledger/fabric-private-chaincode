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
#  - openssl version:           OPENSSL_VERSION
#  - sgxssl version:            SGXSSL_VERSION
#  - additional apt pkgs:       APT_ADD_PKGS


ARG FPC_VERSION=main

FROM hyperledger/fabric-private-chaincode-base-rt:${FPC_VERSION} as common
LABEL org.opencontainers.image.source https://github.com/hyperledger/fabric-private-chaincode

# config/build params
ARG GO_VERSION=1.21.9
ARG NANOPB_VERSION=0.4.7
ARG OPENSSL_VERSION=3.0.12
ARG SGXSSL_VERSION=3.0_Rev1
ARG APT_ADD_PKGS=
ARG EGO_VERSION=1.5.3

# for convenience remember all versions as env variables ..
ENV GO_VERSION=${GO_VERSION}
ENV NANOPB_VERSION=${NANOPB_VERSION}
ENV OPENSSL_VERSION=${OPENSSL_VERSION}
ENV SGXSSL_VERSION=${SGXSSL_VERSION}

ENV DEBIAN_FRONTEND "noninteractive"

WORKDIR /tmp

RUN apt-get update -q \
    && apt-get install -y -q \
    build-essential \
    clang-format \
    cmake \
    git \
    libcurl4-openssl-dev \
    libprotobuf-dev \
    libssl-dev \
    libtool \
    pkg-config \
    python-is-python3 \
    protobuf-compiler \
    python3-protobuf \
    psmisc \
    bc \
    software-properties-common \
    ${APT_ADD_PKGS}

# docker commands (need as we use docker daemon from "outside")
RUN curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add - \
  && add-apt-repository "deb [arch="$(dpkg --print-architecture)"] https://download.docker.com/linux/ubuntu "$(. /etc/os-release && echo "$VERSION_CODENAME")" stable" \
  && apt-get update -q \
  && apt-get install -y -q \
    docker-ce-cli \
    docker-buildx-plugin \
    docker-compose-plugin

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
RUN go install golang.org/x/tools/cmd/goimports@latest \
  && go install google.golang.org/protobuf/cmd/protoc-gen-go@latest \
  && go install github.com/maxbrunsfeld/counterfeiter/v6@latest \
  && go install honnef.co/go/tools/cmd/staticcheck@2023.1.3 \
  && go install github.com/client9/misspell/cmd/misspell@latest \
  && go install github.com/mikefarah/yq/v4@v4.35.2

# Install SGX SSL
ENV SGX_SSL /opt/intel/sgxssl
RUN . /opt/intel/sgxsdk/environment \
  && git clone --depth 1 --branch ${SGXSSL_VERSION} 'https://github.com/intel/intel-sgx-ssl.git' \
  && wget -q -P /tmp/intel-sgx-ssl/openssl_source https://www.openssl.org/source/openssl-${OPENSSL_VERSION}.tar.gz \
  && cd /tmp/intel-sgx-ssl/Linux \
  && make SGX_MODE=SIM NO_THREADS=1 SKIP_INTELCPU_CHECK=TRUE DESTDIR=${SGX_SSL} all \
  # Note: we need explicitly set to SIM as default is HW yet during docker
  # build you never have access to HW, regardless of platform. Note, though,
  # that libraries built work for both modes, on tests are executed for SIM only.
  && make install \
  && make clean \
  && rm -rf /tmp/intel-sgx-ssl

# Install nanopb
ENV NANOPB_PATH=/usr/local/nanopb/
RUN git clone https://github.com/nanopb/nanopb.git ${NANOPB_PATH} \
  && cd ${NANOPB_PATH} \
  && git checkout nanopb-${NANOPB_VERSION} \
  && cd generator/proto \
  && make

# Install ego
RUN wget https://github.com/edgelesssys/ego/releases/download/v${EGO_VERSION}/ego_${EGO_VERSION}_amd64_ubuntu-22.04.deb \
  && apt install -y -q ./ego_${EGO_VERSION}_amd64_ubuntu-22.04.deb \
  && rm ego_${EGO_VERSION}_amd64_ubuntu-22.04.deb
