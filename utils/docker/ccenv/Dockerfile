# Copyright IBM Corp. All Rights Reserved.
# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

# Description:
#   Sets up the basic environment to run fpc cc in docker
#
#  Configuration (build) paramaters (for defaults, see below section with ARGs)
#  - fpc image version:		FPC_VERSION

ARG FPC_VERSION=main

FROM hyperledger/fabric-private-chaincode-base-rt:${FPC_VERSION}
LABEL org.opencontainers.image.source https://github.com/hyperledger/fabric-private-chaincode
