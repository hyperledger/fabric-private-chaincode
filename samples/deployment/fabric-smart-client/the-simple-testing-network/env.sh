#!/usr/bin/env bash

# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

if [[ -z "${FPC_PATH}" ]]; then
  echo "Error: FPC_PATH not set"
  exit 1
fi

if [[ "$#" -ne 1 ]]; then
  echo "ERROR: incorrect number of parameters" >&2
  echo "Use: ./env.sh <org>" >&2
  exit 1
fi

ORG=$1

CONF_PATH=$FPC_PATH/samples/deployment/fabric-smart-client/the-simple-testing-network/testdata/fabric.default/crypto
GATEWAY_CONFIG="$CONF_PATH/peerOrganizations/${ORG,,}.example.com/connections.yaml"

if [[ ! -f "$GATEWAY_CONFIG" ]]; then
    echo "ERROR: no connections.yaml found for ${ORG} at ${GATEWAY_CONFIG}" >&2
    exit 1
fi

MSP_ID=$(yq r $GATEWAY_CONFIG organizations.${ORG^}.mspid)
PEER_ID=$(yq r $GATEWAY_CONFIG organizations.${ORG^}.peers[0])
ADDR=$(yq r $GATEWAY_CONFIG peers.\"${PEER_ID}\".url | sed 's/grpcs:\/\///')

echo "export CONF_PATH=$FPC_PATH/samples/deployment/fabric-smart-client/the-simple-testing-network/testdata/fabric.default/crypto"
echo "export GATEWAY_CONFIG=\$CONF_PATH/peerOrganizations/${ORG,,}.example.com/connections.yaml"
echo "export ORG_PATH=\$CONF_PATH/peerOrganizations/org1.example.com"
echo "export ORDERER_PATH=\$CONF_PATH/ordererOrganizations/example.com"
echo "export CC_NAME=echo"
echo "export CHANNEL_NAME=testchannel"
echo "export CORE_PEER_ADDRESS=${ADDR}"
echo "export CORE_PEER_ID=${PEER_ID}"
echo "export CORE_PEER_LOCALMSPID=${MSP_ID}"
echo "export CORE_PEER_MSPCONFIGPATH=\$ORG_PATH/users/Admin@${ORG,,}.example.com/msp"
echo "export CORE_PEER_TLS_CERT_FILE=\$ORG_PATH/peers/${PEER_ID}/tls/server.crt"
echo "export CORE_PEER_TLS_ENABLED=\"true\""
echo "export CORE_PEER_TLS_KEY_FILE=\$ORG_PATH/peers/${PEER_ID}/tls/server.key"
echo "export CORE_PEER_TLS_ROOTCERT_FILE=\$ORG_PATH/peers/${PEER_ID}/tls/ca.crt"
echo "export ORDERER_CA=\$ORDERER_PATH/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem"
