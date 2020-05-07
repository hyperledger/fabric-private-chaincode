# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

ARG FPC_VERSION=latest

FROM hyperledger/fabric-private-chaincode-dev:${FPC_VERSION}

ARG SGX_MODE=SIM
ENV SGX_MODE=${SGX_MODE}

RUN cd $FPC_PATH
RUN make fpc-sdk

WORKDIR $FPC_PATH
