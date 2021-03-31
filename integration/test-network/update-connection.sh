#!/bin/bash

# Copyright IBM Corp. All Rights Reserved.
# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

if [[ -z "${FPC_PATH}" ]]; then
  echo "Error: FPC_PATH not set"
  exit 1
fi

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
#orgs=("org1")
user="Admin"

shopt -s nullglob

for org in "${orgs[@]}"; do

  ORG_PATH=${FPC_PATH}/integration/test-network/fabric-samples/test-network/organizations/peerOrganizations/${org}.example.com
  CONNECTIONS_PATH=${ORG_PATH}/connection-${org}.yaml
  backup "${CONNECTIONS_PATH}"

  CERTS=("${ORG_PATH}/users/${user}@${org}.example.com/msp/signcerts"/*.pem)
  KEYS=("${ORG_PATH}/users/${user}@${org}.example.com/msp/keystore"/*)

  # add cryptopath and admin cert / key
  yq w -i ${CONNECTIONS_PATH} organizations.${org^}.cryptoPath ${ORG_PATH}/msp
  yq w -i ${CONNECTIONS_PATH} organizations.${org^}.users.${user}.cert.path "${CERTS[0]}"
  yq w -i ${CONNECTIONS_PATH} organizations.${org^}.users.${user}.key.path "${KEYS[0]}"

  # add channels and entity matcher
  yq m -i ${CONNECTIONS_PATH} - <<EOF
channels:
  _default:
    peers:
      peer0.${org}.example.com:
        endorsingPeer: true
        chaincodeQuery: true
        ledgerQuery: true
        eventSource: true
entityMatchers:
  peer:
    - pattern: ([^:]+):(\d+)
      urlSubstitutionExp: localhost:\${2}
      sslTargetOverrideUrlSubstitutionExp: \${1}
      mappedHost: \${1}
  orderer:
    - pattern: ([^:]+):(\d+)
      urlSubstitutionExp: localhost:\${2}
      sslTargetOverrideUrlSubstitutionExp: \${1}
      mappedHost: \${1}
EOF

done



