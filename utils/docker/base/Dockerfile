# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

FROM ubuntu:bionic

RUN apt-get update \
 && apt-get install -y \
    wget \
    curl \
    git \
    build-essential \
    ca-certificates \
    protobuf-compiler \
    python-protobuf \
    libprotobuf-dev \
    libssl-dev \
    libcurl4-openssl-dev \
    libtool \
    cmake \
    vim \
    unzip \
    basez \
    pkg-config \
 && apt-get -y -q upgrade \
 && apt-get clean \
 && rm -rf /var/lib/apt/lists/*

ARG SGX_REPO=https://download.01.org/intel-sgx/linux-2.6/ubuntu18.04-server
ARG SGX_LIB_COMMON=libsgx-enclave-common_2.6.100.51363-bionic1_amd64.deb
ARG SGX_SDK_BIN=sgx_linux_x64_sdk_2.6.100.51363.bin

ARG PROTO_VERSION=3.11.4
ARG PROTO_ZIP=protoc-${PROTO_VERSION}-linux-x86_64.zip
ARG PROTO_REPO=https://github.com/google/protobuf/releases/download
ARG PROTO_DIR=/usr/local/proto3

# Create a place for our SGX SDK
RUN mkdir -p /opt/intel \
 && mkdir -p /etc/init

WORKDIR /opt/intel

# Install SGX COMMON
RUN wget ${SGX_REPO}/${SGX_LIB_COMMON} \
 && dpkg -i ${SGX_LIB_COMMON}

# Install SGX SDK
RUN wget ${SGX_REPO}/${SGX_SDK_BIN} \
 && chmod +x ${SGX_SDK_BIN} \
 && echo "yes" | ./${SGX_SDK_BIN} \
 && rm ${SGX_SDK_BIN}

ENV SGX_SDK=/opt/intel/sgxsdk
ENV PATH=$PATH:$SGX_SDK/bin:$SGX_SDK/bin/x64
ENV PKG_CONFIG_PATH=$PKG_CONFIG_PATH:$SGX_SDK/pkgconfig
ENV LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$SGX_SDK/sdk_libs

# install custom protoc
RUN wget ${PROTO_REPO}/v${PROTO_VERSION}/${PROTO_ZIP} \
 && unzip ${PROTO_ZIP} -d ${PROTO_DIR}

ENV PROTOC_CMD=${PROTO_DIR}/bin/protoc
