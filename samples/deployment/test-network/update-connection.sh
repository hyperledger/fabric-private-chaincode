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

script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd -P)
tmp_dir=$(mktemp -d -t tmp-XXXXXXXXXX --tmpdir="${script_dir}")

trap cleanup SIGINT SIGTERM ERR EXIT
cleanup() {
  trap - SIGINT SIGTERM ERR EXIT
  rm -rf "${tmp_dir}"
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
  CONNECTIONS_PATH=${ORG_PATH}/connection-${org}.yaml
  backup "${CONNECTIONS_PATH}"

  CERTS=("${ORG_PATH}/users/${user}@${org}.example.com/msp/signcerts"/*.pem)
  KEYS=("${ORG_PATH}/users/${user}@${org}.example.com/msp/keystore"/*)

  # add cryptopath and admin cert / key
  yq eval ".organizations.${org^}.cryptoPath = \"${ORG_PATH}/msp\"" ${CONNECTIONS_PATH} -i
  yq eval ".organizations.${org^}.users.${user}.cert.path = \"${CERTS[0]}\"" ${CONNECTIONS_PATH} -i
  yq eval ".organizations.${org^}.users.${user}.key.path = \"${KEYS[0]}\"" ${CONNECTIONS_PATH} -i

  # add channels and entity matcher
  yq eval-all 'select(fileIndex == 0) * select(fileIndex == 1)' ${CONNECTIONS_PATH} -i <<EOF

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

  # fetch all peers from connections
  yq e ".peers" "${CONNECTIONS_PATH}" >> "${tmp_dir}/peers-${org}.yaml"
done

# consolidate all collected peers in a single peers.yaml
yq eval-all 'select(fileIndex == 0) * select(fileIndex == 1)' "${tmp_dir}"/peers-*.yaml >> "${tmp_dir}/peers.yaml"
yq eval-all 'select(fileIndex == 0) * select(fileIndex == 1)' "${CONNECTIONS_PATH}" "${tmp_dir}/peers.yaml" -i


# merge peers.yaml into all connection files
for org in "${orgs[@]}"; do
  ORG_PATH="${FPC_PATH}/samples/deployment/test-network/fabric-samples/test-network/organizations/peerOrganizations/${org}.example.com"
  CONNECTIONS_PATH="${ORG_PATH}/connection-${org}.yaml"
  yq eval-all 'select(fileIndex == 0) * select(fileIndex == 1)' "${CONNECTIONS_PATH}" "${tmp_dir}/peers.yaml" -i
  yq e "." "${CONNECTIONS_PATH}"
done

echo "Updated!"
