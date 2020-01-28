#!/bin/bash

#
# Copyright IBM Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#

set -ev

export WAIT_TIME=15
export version=1.0
export FABRIC_BIN_DIR=/project/src/github.com/hyperledger/fabric/.build/bin
export PEER_CMD=/project/src/github.com/hyperledger-labs/fabric-private-chaincode/fabric/bin/peer.sh

docker exec peer0.org1.example.com ${PEER_CMD} chaincode install -n mockcc -v ${version} --path github.com/hyperledger-labs/fabric-private-chaincode/demo/chaincode/golang/cmd -l golang
sleep ${WAIT_TIME}

docker exec peer0.org1.example.com ${PEER_CMD} chaincode instantiate -n mockcc -v ${version} --channelID mychannel -c '{"Args":[]}'
sleep ${WAIT_TIME}

docker exec peer0.org1.example.com ${PEER_CMD} chaincode install -n auctioncc -v ${version} --path demo/chaincode/fpc/_build/lib -l fpc-c
sleep ${WAIT_TIME}

docker exec peer0.org1.example.com ${PEER_CMD} chaincode instantiate -n auctioncc -v ${version} --channelID mychannel -c '{"Args":[]}'
sleep ${WAIT_TIME}
