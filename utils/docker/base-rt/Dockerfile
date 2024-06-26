# Copyright IBM Corp. All Rights Reserved.
# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

# Description:
#   Builds the environment with all prerequistes needed to _run_ (but not necessarily build) SGX-enabled apps as needed in FPC
#
#  Configuration (build) paramaters (for defaults, see below section with ARGs)
#  - ubuntu version to use:     UBUNTU_VERSION
#  - ubuntu name to use:        UBUNTU_NAME
#  - sgx sdk/psw version:       SGX
#  - protobuf version:          PROTO_VERSION
#  - additional apt pkgs:       APT_ADD_PKGS

# config/build params (part 1)
ARG UBUNTU_VERSION=22.04
ARG UBUNTU_NAME=jammy
# NOTE:
# - unfortunately, we do need both name (for repo) and version (for sgx directories), only docker image supports both ..
#   18.04 <-> bionic, 20.04 <-> focal
# - right now, full sgx support exists only for bionic (18.04) and, since v2.12, focal (20.04); 
#   xenial (16.04) has support only PSW but not SDK; 

FROM ubuntu:${UBUNTU_VERSION}
LABEL org.opencontainers.image.source https://github.com/hyperledger/fabric-private-chaincode

# Dockerfile limitations force a repetition of global args
ARG UBUNTU_VERSION
ARG UBUNTU_NAME

# config/build params (part 2)
ARG SGX=2.22
ARG PROTO_VERSION=22.3
ARG APT_ADD_PKGS=

# We define it here even so the installation path is known to all derivates,
# even when not all of them use go (or have it installed)
ENV GOPATH=/project

ENV DEBIAN_FRONTEND "noninteractive"

# Get all necessary apt packages
RUN apt-get update -q \
  && apt-get install -y -q \
    basez \
    ca-certificates \
    curl \
    gnupg2 \
    unzip \
    wget \
    # jq need for e.g., external-builder
    jq \
    build-essential \
    ${APT_ADD_PKGS} \
  && apt-get clean \
  && rm -rf /var/lib/apt/lists/*

# Install SGX PSW packages
RUN echo "deb [arch=amd64] https://download.01.org/intel-sgx/sgx_repo/ubuntu ${UBUNTU_NAME} main" >> /etc/apt/sources.list \
  && wget -qO - https://download.01.org/intel-sgx/sgx_repo/ubuntu/intel-sgx-deb.key | apt-key add - \
  && apt-get update -q \
  && apt-get install -y -q \
    # We do not need daemons like AESMD as we run them on host (side-steps also
    # issues with config of /etc/aesmd.conf like proxy ..). Without this option
    # aesmd and lots of other plugsin are automatically pulled in. 
    # See SGX Installation notes and, in particular, linux/installer/docker/Dockerfile
    # in linux-sgx git repo of sdk/psw source. 
    --no-install-recommends \
    libsgx-urts \
    libsgx-uae-service \
    #   - EPID-based attestation service \
    libsgx-epid \
    #   - DCAP-based attesation service
    #     libsgx-dcap* ...
  && apt-get clean \
  && rm -rf /var/lib/apt/lists/*

# Install SGX SDK
# Note: not all descendents of this base image, e.g., ccenv, boilerplate and fpc-app, build sgx app.
#   However, as simulation-mode libraries are only in the sdk and not in the psw packages, we need
#   it already here and not only in the 'dev' image. 
WORKDIR /opt/intel
RUN SGX_SDK_BIN_REPO=https://download.01.org/intel-sgx/sgx-linux/${SGX}/distro/ubuntu${UBUNTU_VERSION}-server \
  && SGX_SDK_BIN_FILE=$(wget -P /tmp --delete-after --spider --recursive --level=1 --no-parent ${SGX_SDK_BIN_REPO} 2>&1 | perl  -ne 'if (m|'${SGX_SDK_BIN_REPO}'/(sgx_linux_x64_sdk.*)|) { print "$1\n"; }') \
  && wget -q -P /tmp ${SGX_SDK_BIN_REPO}/${SGX_SDK_BIN_FILE} \
  && chmod +x /tmp/${SGX_SDK_BIN_FILE} \
  && echo -e "no\n/opt/intel" | /tmp/${SGX_SDK_BIN_FILE} \
  && rm /tmp/${SGX_SDK_BIN_FILE}

ENV SGX_SDK=/opt/intel/sgxsdk
ENV PATH=${PATH}:${SGX_SDK}/bin:${SGX_SDK}/bin/x64
ENV PKG_CONFIG_PATH=${PKG_CONFIG_PATH}:${SGX_SDK}/pkgconfig
ENV LD_LIBRARY_PATH=${LD_LIBRARY_PATH}:${SGX_SDK}/sdk_libs

# LVI mitigations, needed to compile sgxssl, requires a
#   recent version of binutils (>= 2.32).
WORKDIR /opt/intel
RUN SGX_SDK_BINUTILS_REPO=https://download.01.org/intel-sgx/sgx-linux/${SGX} \
  && SGX_SDK_BINUTILS_FILE=$(wget -P /tmp --delete-after --spider --recursive --level=1 --no-parent ${SGX_SDK_BINUTILS_REPO} 2>&1 | perl  -ne 'if (m|'${SGX_SDK_BINUTILS_REPO}'/(as.ld.objdump.*)|) { print "$1\n"; }') \
  && wget -q -P /tmp ${SGX_SDK_BINUTILS_REPO}/${SGX_SDK_BINUTILS_FILE} \
  && mkdir sgxsdk.extras \
  && cd sgxsdk.extras \
  && tar -zxf /tmp/${SGX_SDK_BINUTILS_FILE} \
  && rm /tmp/${SGX_SDK_BINUTILS_FILE} \
  && (cd /opt/intel/sgxsdk.extras/external/toolset/ && \
      for f in $(ls | grep -v ${UBUNTU_VERSION}); do rm -rf ${f}; done)
# Note: above install file contains binutitls for _all_ supported distros
#   and are fairly large, so clean out anything we do not need
ENV PATH="/opt/intel/sgxsdk.extras/external/toolset/ubuntu${UBUNTU_VERSION}:${PATH}"


# install custom protoc
ENV PROTO_DIR=/usr/local/proto3
RUN PROTO_ZIP=protoc-${PROTO_VERSION}-linux-x86_64.zip \
  && PROTO_REPO=https://github.com/google/protobuf/releases/download \
  && wget -q ${PROTO_REPO}/v${PROTO_VERSION}/${PROTO_ZIP} \
  && unzip ${PROTO_ZIP} -d ${PROTO_DIR} \
  && rm ${PROTO_ZIP}
ENV PROTOC_CMD=${PROTO_DIR}/bin/protoc
