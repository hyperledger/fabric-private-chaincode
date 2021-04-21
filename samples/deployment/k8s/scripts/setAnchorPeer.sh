#!/usr/bin/env bash

# Copyright IBM Corp. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

# update anchor peer
# for more details on this see https://hyperledger-fabric.readthedocs.io/en/latest/config_update.html

peer channel fetch config config_block.pb -o orderer0:7050 -c mychannel --tls --cafile $ORDERER_CA
configtxlator proto_decode --input config_block.pb --type common.Block --output config_block.json
jq .data.data[0].payload.data.config config_block.json > config.json

HOST=$(echo $CORE_PEER_ADDRESS | awk -F ':' '{print $1}')
PORT=$(echo $CORE_PEER_ADDRESS | awk -F ':' '{print $2}')

jq '.channel_group.groups.Application.groups.'${CORE_PEER_LOCALMSPID}'.values += {"AnchorPeers":{"mod_policy": "Admins","value":{"anchor_peers": [{"host": "'$HOST'","port": '$PORT'}]},"version": "0"}}' config.json > modified_config.json

configtxlator proto_encode --input config.json --type common.Config --output config.pb
configtxlator proto_encode --input modified_config.json --type common.Config --output modified_config.pb
configtxlator compute_update --channel_id mychannel --original config.pb --updated modified_config.pb --output config_update.pb

configtxlator proto_decode --input config_update.pb --type common.ConfigUpdate --output config_update.json
echo '{"payload":{"header":{"channel_header":{"channel_id":"mychannel", "type":2}},"data":{"config_update":'$(cat config_update.json)'}}}' | jq . > config_update_in_envelope.json
configtxlator proto_encode --input config_update_in_envelope.json --type common.Envelope --output config_update_in_envelope.pb

peer channel update -f config_update_in_envelope.pb -c mychannel -o orderer0:7050 --tls --cafile $ORDERER_CA
