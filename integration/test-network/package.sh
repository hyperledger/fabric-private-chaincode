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

DEPLOYMENT_PATH=$1

if [[ -z "${DEPLOYMENT_PATH}" ]]; then
  echo "ERROR: No outpath defined"
  echo "Use: ./package.sh _deployment"
  exit 1
fi

packageChaincode() {
  output_dir=$1
  cc_name=$2
  cc_version=$3
  cc_type=$4
  port=$5

  tmp_dir=$(mktemp -d -t "${cc_name}-packageXXX")
  mkdir -p "${output_dir}"

  for p in "${PEERS[@]}"; do
    local out_dir=${tmp_dir}/connections/${p}

    mkdir -p $out_dir
    addr="${cc_name}.${p}:${port}"
    createConnection ${out_dir} ${addr}
    echo ${out_dir}
    ls ${out_dir}
    cat ${out_dir}/connection.json
  done

  tar -czf "${tmp_dir}/code.tar.gz" -C ${tmp_dir}/connections .

  output="${output_dir}/${cc_name}.tgz"

  createMetafile ${tmp_dir} ${cc_name} ${cc_version} ${cc_type}

  tar -czf "${output}" -C "${tmp_dir}" code.tar.gz metadata.json

  rm -rf "${tmp_dir}"
}

createMetafile() {
  out_dir=$1
  cc_name=$2
  cc_version=$3
  cc_type=$4

  cat >"${out_dir}/metadata.json" <<EOF
{
  "type": "${cc_type}",
  "label": "${cc_name}_${cc_version}"
}
EOF
}

createConnection() {
  out_dir=$1
  address=$2
  timeout=10s
  use_tls=false

  cat >"${out_dir}/connection.json" <<EOF
{
  "address": "${address}",
  "dial_timeout": "${timeout}",
  "tls_required": ${use_tls}
}
EOF
}

PEERS=("peer0.org1.example.com" "peer0.org2.example.com")

packageChaincode "${DEPLOYMENT_PATH}" "ercc" "1.0" "external" 9999
packageChaincode "${DEPLOYMENT_PATH}" "ecc" "1.0" "external" 9999
