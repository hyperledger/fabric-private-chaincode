# Copyright IBM Corp. All Rights Reserved.
# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

ARG FPC_CCENV_IMAGE=hyperledger/fabric-private-chaincode-ccenv:main

FROM ${FPC_CCENV_IMAGE}

ENV PATH=/opt/ercc:$PATH

WORKDIR /opt/ercc
COPY ercc .

EXPOSE 9999
CMD ["ercc"]
