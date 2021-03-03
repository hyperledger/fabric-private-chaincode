#!/bin/bash

# Copyright IBM Corp. All Rights Reserved.
# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

packageChaincode() {
  packageDir=$1
  packageName=$2
  cc_id=$3
  cc_version=$4
  cc_type=$5
  endpoint=$6
  peer=$7

  local tmp_dir=$(mktemp -d -t "${cc_id}-packageXXX")
  mkdir -p "${packageDir}"

  local out_dir=${tmp_dir}/connections/${peer}

  mkdir -p $out_dir
  createConnection ${out_dir} ${endpoint}
  [ -z ${DEBUG+x} ] || {
    echo ${out_dir};
    ls ${out_dir};
    cat ${out_dir}/connection.json
  }

  tar -czf "${tmp_dir}/code.tar.gz" -C ${tmp_dir}/connections .

  output="${packageDir}/${packageName}"

  createMetafile ${tmp_dir} ${cc_id} ${cc_version} ${cc_type}

  tar -czf "${output}" -C "${tmp_dir}" code.tar.gz metadata.json

  rm -rf "${tmp_dir}"
}

createMetafile() {
  out_dir=$1
  cc_id=$2
  cc_version=$3
  cc_type=$4

  cat >"${out_dir}/metadata.json" <<EOF
{
  "type": "${cc_type}",
  "label": "${cc_id}_${cc_version}"
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
