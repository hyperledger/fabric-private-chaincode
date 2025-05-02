#!/usr/bin/env bash

# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

set -Eeuo pipefail

# test settings
CC_PATH="$FPC_PATH/samples/chaincode/echo-go"
CC_ID=echo-go
CHANNEL_NAME=mychannel

trap cleanup SIGINT SIGTERM ERR EXIT
function cleanup() {
  # reset traps
  trap - SIGINT SIGTERM ERR EXIT

  echo "########################################"
  echo "Cleanup"
  echo "########################################"

  make -C "$FPC_PATH/samples/deployment/test-network" clean
}

function test_network_setup() {
  echo "########################################"
  echo "Run setup"
  echo "########################################"

  cd "$FPC_PATH/samples/deployment/test-network"
  ./setup.sh

  echo "########################################"
  echo "start test-netowrk"
  echo "########################################"

  cd fabric-samples/test-network
  ./network.sh up createChannel -c $CHANNEL_NAME

  docker ps
}

function test_deploy() {
  echo "########################################"
  echo "build $CC_ID"
  echo "########################################"

  export CC_ID=$CC_ID
  export FPC_CCENV_IMAGE=ubuntu:22.04
  export ERCC_GOTAGS=

  # build ercc
  GOOS=linux make -C "$FPC_PATH/ercc" build docker

  # build fpc chaincode
  GOOS=linux CC_NAME=$CC_ID make -C "$CC_PATH" with_go docker

  local ccver
  ccver=$(cat "$CC_PATH/mrenclave")

  export CC_VER=$ccver

  echo "########################################"
  echo "install FPC chaincode"
  echo "########################################"

  cd "$FPC_PATH/samples/deployment/test-network"
  ./installFPC.sh

  echo "########################################"
  echo "Run chaincodes ..."
  echo "########################################"

  cd "$FPC_PATH/samples/deployment/test-network"
  make ercc-ecc-start

  docker ps

  echo "########################################"
  echo "Update connection details ..."
  echo "########################################"

  cd "$FPC_PATH/samples/deployment/test-network"
  ./update-connection.sh
}

function test_simple_go() {
  echo "########################################"
  echo "run simple_go"
  echo "########################################"

  cd "$FPC_PATH/samples/application/simple-go"
  CC_ID=$CC_ID ORG_NAME=Org1 go run . -withLifecycleInitEnclave

  CC_ID=$CC_ID ORG_NAME=Org1 go run .
  CC_ID=$CC_ID ORG_NAME=Org2 go run .
}

function test_simple_cli() {
  echo "########################################"
  echo "run simple_cli"
  echo "########################################"

  cd "$FPC_PATH/samples/application/simple-cli-go"
  make

  export CC_ID=$CC_ID
  export CHANNEL_NAME=$CHANNEL_NAME
  export CORE_PEER_ADDRESS=localhost:7051
  export CORE_PEER_ID=peer0.org1.example.com
  export CORE_PEER_ORG_NAME=org1
  export CORE_PEER_LOCALMSPID=Org1MSP
  export CORE_PEER_MSPCONFIGPATH="$FPC_PATH/samples/deployment/test-network/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp"
  export CORE_PEER_TLS_CERT_FILE="$FPC_PATH/samples/deployment/test-network/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/server.crt"
  export CORE_PEER_TLS_ENABLED="true"
  export CORE_PEER_TLS_KEY_FILE="$FPC_PATH/samples/deployment/test-network/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/server.key"
  export CORE_PEER_TLS_ROOTCERT_FILE="$FPC_PATH/samples/deployment/test-network/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"
  export ORDERER_CA="$FPC_PATH/samples/deployment/test-network/fabric-samples/test-network/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem"
  export GATEWAY_CONFIG="$FPC_PATH/samples/deployment/test-network/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/connection-org1.yaml"
  export SGX_CREDENTIALS_PATH="$FPC_PATH/config/ias"

  # note that we skip the init here as this is done in test_simple_go already
  #./fpcclient init $CORE_PEER_ID

  ./fpcclient invoke foo
  ./fpcclient query foo
}

function test_blockchain_explorer() {
   echo "########################################"
   echo "start blockchain explorer"
   echo "########################################"

   cd "$FPC_PATH/samples/deployment/test-network/blockchain-explorer"
   docker compose up -d

   #curl http://localhost:8080/

   docker compose down -v
}

function test_shutdown() {
  make -C "$FPC_PATH/samples/deployment/test-network" ercc-ecc-stop
  cd "$FPC_PATH/samples/deployment/test-network/fabric-samples/test-network"
  ./network.sh down
}

test_network_setup
test_deploy
test_simple_go
test_simple_cli
test_blockchain_explorer
test_shutdown
