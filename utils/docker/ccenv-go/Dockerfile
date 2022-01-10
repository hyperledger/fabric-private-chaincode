# Copyright IBM Corp. All Rights Reserved.
# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

# Description:
#   Sets up the basic environment to run fpc go chaincode in docker
#
#  Configuration (build) paramaters (for defaults, see below section with ARGs)
#  - fpc image version:		FPC_VERSION
#  - ego version:           EGO_VERSION

ARG FPC_VERSION=main

FROM hyperledger/fabric-private-chaincode-base-rt:${FPC_VERSION}

ARG EGO_VERSION=0.4.0

RUN wget https://github.com/edgelesssys/ego/releases/download/v${EGO_VERSION}/ego_${EGO_VERSION}_amd64.deb \
  && apt install -y ./ego_${EGO_VERSION}_amd64.deb \
  && rm ego_${EGO_VERSION}_amd64.deb
