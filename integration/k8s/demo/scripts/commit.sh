#!/usr/bin/env bash

# Copyright IBM Corp. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

# commit enclave registry
peer lifecycle chaincode checkcommitreadiness --channelID mychannel --name ercc --version 1.0 --sequence 1 -o orderer0:7050 --tls --cafile $ORDERER_CA
peer lifecycle chaincode commit -o orderer0:7050 --channelID mychannel --name ercc --version 1.0 --sequence 1 --tls true --cafile $ORDERER_CA --peerAddresses peer0-org1:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1/peers/peer0-org1/tls/ca.crt --peerAddresses peer0-org2:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2/peers/peer0-org2/tls/ca.crt

# commit FPC chaincode
peer lifecycle chaincode checkcommitreadiness --channelID mychannel --name fpccc --version  $FPC_MRENCLAVE --sequence 1 -o orderer0:7050 --tls --cafile $ORDERER_CA
peer lifecycle chaincode commit -o orderer0:7050 --channelID mychannel --name fpccc --version  $FPC_MRENCLAVE --sequence 1 --tls true --cafile $ORDERER_CA --peerAddresses peer0-org1:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1/peers/peer0-org1/tls/ca.crt --peerAddresses peer0-org2:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2/peers/peer0-org2/tls/ca.crt
