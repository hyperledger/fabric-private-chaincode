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
  yq ".organizations.${org^}.cryptoPath = \"${ORG_PATH}/msp\"" -i "${CONNECTIONS_PATH}"
  yq ".organizations.${org^}.users.${user}.cert.path = \"${CERTS[0]}\"" -i "${CONNECTIONS_PATH}"
  yq ".organizations.${org^}.users.${user}.key.path = \"${KEYS[0]}\"" -i "${CONNECTIONS_PATH}"

  # Create temporary file with channels and entity matchers
  temp_yaml=$(mktemp -t temp-XXXXXXXXXX.yaml --tmpdir="${tmp_dir}")
  cat > "${temp_yaml}" <<EOF
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

  # add channels and entity matcher
  yq eval-all 'select(fileIndex == 0) * select(fileIndex == 1)' -i "${CONNECTIONS_PATH}" "${temp_yaml}"

  # fetch all peers from connections
  yq ".peers" "${CONNECTIONS_PATH}" >> "${tmp_dir}/peers-${org}.yaml"
done

# consolidate all collected peers in a single peers.yaml
yq eval-all '. as $item ireduce ({}; . * $item )' ${tmp_dir}/peers-*.yaml >> "${tmp_dir}/peers.yaml"
yq 'true' "${tmp_dir}/peers.yaml" > /dev/null

# merge peers.yaml into all connection files
for org in "${orgs[@]}"; do
  ORG_PATH="${FPC_PATH}/samples/deployment/test-network/fabric-samples/test-network/organizations/peerOrganizations/${org}.example.com"
  CONNECTIONS_PATH="${ORG_PATH}/connection-${org}.yaml"
  yq eval-all 'select(fileIndex == 0) * select(fileIndex == 1)' -i "${CONNECTIONS_PATH}" "${tmp_dir}/peers.yaml"
  yq 'true' "${CONNECTIONS_PATH}" > /dev/null
done

echo "Updated!"
