#!/usr/bin/env bash

# Copyright IBM Corp. All Rights Reserved.
# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

if [[ -z "${FPC_PATH}" ]]; then
  echo "Error: FPC_PATH not set"
  exit 1
fi

trap cleanup SIGINT SIGTERM ERR EXIT
cleanup() {
  trap - SIGINT SIGTERM ERR EXIT
}

backup() {
  FILE=$1
  BACKUP="${FILE}.backup"

  if [[ -e "${BACKUP}" ]]; then
    cp "${BACKUP}" "${FILE}"
  else
    cp "${FILE}" "${BACKUP}"
  fi
}

orgs=("org1" "org2")
user="Admin"

shopt -s nullglob

for org in "${orgs[@]}"; do

  ORG_PATH=${FPC_PATH}/samples/deployment/test-network/fabric-samples/test-network/organizations/peerOrganizations/${org}.example.com
  EXTERNAL_CONNECTIONS_PATH=${ORG_PATH}/external-connection-${org}.yaml
  CONNECTIONS_PATH=${ORG_PATH}/connection-${org}.yaml

  # Copy the file from the connection profile
  cp "${CONNECTIONS_PATH}" "${EXTERNAL_CONNECTIONS_PATH}"

  backup "${EXTERNAL_CONNECTIONS_PATH}"


  # This is needed in both files
  yq eval ".\"peer0.org1.example.com\".url = \"grpcs://peer0.org1.example.com:7051\"" -i "$EXTERNAL_CONNECTIONS_PATH"
  yq eval ".\"peer0.org2.example.com\".url = \"grpcs://peer0.org2.example.com:9051\"" -i "$EXTERNAL_CONNECTIONS_PATH"
  # Check if the org is org1
  if [[ "$org" == "org1" ]]; then
    # edit localhost urls to use hostnames for org1
    yq eval ".peers.\"peer0.org1.example.com\".url = \"grpcs://peer0.org1.example.com:7051\"" -i "$EXTERNAL_CONNECTIONS_PATH"
    yq eval ".certificateAuthorities.\"ca.org1.example.com\".url = \"https://ca.org1.example.com:7054\"" -i "$EXTERNAL_CONNECTIONS_PATH"

  # Check if the org is org2
  elif [[ "$org" == "org2" ]]; then
    # edit localhost urls to use hostnames for org2
    yq eval ".peers.\"peer0.org2.example.com\".url = \"grpcs://peer0.org2.example.com:9051\"" -i "$EXTERNAL_CONNECTIONS_PATH"
    yq eval ".certificateAuthorities.\"ca.org2.example.com\".url = \"https://ca.org2.example.com:8054\"" -i "$EXTERNAL_CONNECTIONS_PATH"
  fi
  # remove entity matcher
  yq eval 'del(.entityMatchers)' -i "$EXTERNAL_CONNECTIONS_PATH"

done

echo "Updated!"
