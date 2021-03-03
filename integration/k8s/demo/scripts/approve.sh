#!/usr/bin/env bash

# Copyright IBM Corp. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

ERCC_PKG_ID=$(peer lifecycle chaincode queryinstalled | grep ercc | awk '{print $3}' | sed 's/.$//')
FPCCC_PKG_ID=$(peer lifecycle chaincode queryinstalled | grep fpccc | awk '{print $3}' | sed 's/.$//')

# approve enclave registry
peer lifecycle chaincode approveformyorg --channelID mychannel --name ercc --version 1.0 --package-id $ERCC_PKG_ID --sequence 1 -o orderer0:7050 --tls --cafile $ORDERER_CA

# approve FPC chaincode
peer lifecycle chaincode approveformyorg --channelID mychannel --name fpccc --version $FPC_MRENCLAVE --package-id $FPCCC_PKG_ID --sequence 1 -o orderer0:7050 --tls --cafile $ORDERER_CA
