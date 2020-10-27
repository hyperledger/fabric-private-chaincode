# Copyright IBM Corp. All Rights Reserved.
# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

FROM hyperledger/fabric-private-chaincode-base-rt:latest

ENV PATH=/opt/ercc:$PATH

WORKDIR /opt/ercc
COPY ercc .

EXPOSE 9999
CMD ["ercc"]
