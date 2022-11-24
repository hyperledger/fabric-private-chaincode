#!/usr/bin/env bash

# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

if [[ -z "${FPC_PATH}" ]]; then
  echo "Error: FPC_PATH not set"
  exit 1
fi

if [[ -z "${CC_NAME}" ]]; then
  echo "Error: CC_NAME not set"
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

echo "export CONF_PATH=$FPC_PATH/samples/deployment/fabric-smart-client/the-simple-testing-network/testdata/fabric.default/crypto" > ${ORG}.env
echo "export GATEWAY_CONFIG=\$CONF_PATH/peerOrganizations/${ORG,,}.example.com/connections.yaml" >> ${ORG}.env
echo "export ORG_PATH=\$CONF_PATH/peerOrganizations/${ORG,,}.example.com" >> ${ORG}.env
echo "export ORDERER_PATH=\$CONF_PATH/ordererOrganizations/example.com" >> ${ORG}.env
#echo "export CC_NAME=${CC_NAME}" >> ${ORG}.env
echo "export CHANNEL_NAME=testchannel" >> ${ORG}.env
echo "export CORE_PEER_ADDRESS=${ADDR}" >> ${ORG}.env
echo "export CORE_PEER_ID=${PEER_ID}" >> ${ORG}.env
echo "export CORE_PEER_LOCALMSPID=${MSP_ID}" >> ${ORG}.env
echo "export CORE_PEER_MSPCONFIGPATH=\$ORG_PATH/users/Admin@${ORG,,}.example.com/msp" >> ${ORG}.env
echo "export CORE_PEER_TLS_CERT_FILE=\$ORG_PATH/peers/${PEER_ID}/tls/server.crt" >> ${ORG}.env
echo "export CORE_PEER_TLS_ENABLED=\"true\"" >> ${ORG}.env
echo "export CORE_PEER_TLS_KEY_FILE=\$ORG_PATH/peers/${PEER_ID}/tls/server.key" >> ${ORG}.env
echo "export CORE_PEER_TLS_ROOTCERT_FILE=\$ORG_PATH/peers/${PEER_ID}/tls/ca.crt" >> ${ORG}.env
echo "export ORDERER_CA=\$ORDERER_PATH/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" >> ${ORG}.env
echo "export FABRIC_LOGGING_SPEC=error" >> ${ORG}.env

cat ${ORG}.env

echo
echo "To get the FPC cli environment run:"
echo "source ${ORG}.env"
