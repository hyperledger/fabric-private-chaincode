# Copyright IBM Corp. All Rights Reserved.
# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

FROM hyperledger/fabric-private-chaincode-base-rt:latest

ARG CC_PATH="/opt/ecc"
ARG CC_LIB_PATH="${CC_PATH}/enclave/lib"
RUN mkdir -p ${CC_LIB_PATH}

ENV LD_LIBRARY_PATH="${LD_LIBRARY_PATH}:${CC_LIB_PATH}"
ENV PATH="${CC_PATH}:${PATH}"

COPY _build/lib/libsgxcc.so ${CC_LIB_PATH}

WORKDIR ${CC_PATH}
