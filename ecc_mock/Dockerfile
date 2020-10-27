# Copyright IBM Corp. All Rights Reserved.
# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

FROM hyperledger/fabric-private-chaincode-base-rt:latest

ENV PATH=/opt/ecc:$PATH

WORKDIR /opt/ecc
COPY ecc .

EXPOSE 9999
CMD ["ecc"]
