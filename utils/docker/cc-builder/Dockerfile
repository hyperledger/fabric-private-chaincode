# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

FROM hyperledger/fabric-private-chaincode-dev:latest

RUN cd $FPC_PATH
RUN make fpc-sdk

WORKDIR $FPC_PATH
