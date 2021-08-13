# Copyright IBM Corp. All Rights Reserved.
# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

ARG FPC_VERSION=main

FROM hyperledger/fabric-private-chaincode-ccenv:${FPC_VERSION}

ENV PATH=/opt/ercc:$PATH

WORKDIR /opt/ercc
COPY ercc .

EXPOSE 9999
CMD ["ercc"]
