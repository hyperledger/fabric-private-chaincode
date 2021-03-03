#!/usr/bin/env bash

# Copyright IBM Corp. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

# for all peers of orgX we install ercc and fpccc
ERCC=ercc-peer0-$ORG
FPCCC=fpccc-peer0-$ORG

peer lifecycle chaincode install packages/$ERCC.tgz
peer lifecycle chaincode install packages/$FPCCC.tgz
peer lifecycle chaincode queryinstalled

ERCC_PKG_ID=$(peer lifecycle chaincode queryinstalled | grep ercc | awk '{print $3}' | sed 's/.$//')
FPCCC_PKG_ID=$(peer lifecycle chaincode queryinstalled | grep fpccc | awk '{print $3}' | sed 's/.$//')

echo "$ERCC=$ERCC_PKG_ID" >> packages/chaincode-config.properties
echo "$FPCCC=$FPCCC_PKG_ID" >> packages/chaincode-config.properties
