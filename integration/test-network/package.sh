#!/bin/bash

# Copyright IBM Corp. All Rights Reserved.
# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

# Package a fpc chaincode for CaaS mode
# (for normal external-builder, use '$FPC_PATH/fabric/bin/peer.sh lifecycle chaincode package')

set -euo pipefail

#DEBUG=true # uncomment (or define when calling script) to show debug output


if [[ -z "${FPC_PATH}" ]]; then
  echo "Error: FPC_PATH not set"
  exit 1
fi

if [ "$#" -ne 6 ]; then
  echo "ERROR: incorrect number of parameters" >&2
  echo "Use: ./package.sh <_deployment> <ercc-id> <ercc-version> <cc-id> <cc-version> <peer-id>" >&2

  exit 1
fi


DEPLOYMENT_PATH="$1"
ERCC_ID="$2"
ERCC_VER="$3"
CC_ID="$4"
CC_VER="$5"
PEER="$6"

CHAINCODE_SERVER_PORT=9999

packageChaincode() {
  output_dir=$1
  cc_name=$2
  cc_version=$3
  cc_type=$4
  port=$5
  peer=$6

  local tmp_dir=$(mktemp -d -t "${cc_name}-packageXXX")
  mkdir -p "${output_dir}"

  local out_dir=${tmp_dir}/connections/${peer}

  mkdir -p $out_dir
  addr="${cc_name}.${peer}:${port}"
  createConnection ${out_dir} ${addr}
  [ -z ${DEBUG+x} ] || {
    echo ${out_dir};
    ls ${out_dir};
    cat ${out_dir}/connection.json
  }

  tar -czf "${tmp_dir}/code.tar.gz" -C ${tmp_dir}/connections .

  output="${output_dir}/${cc_name}.${peer}.tgz"

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

packageChaincode "${DEPLOYMENT_PATH}" "${ERCC_ID}" "${ERCC_VER}" "external" ${CHAINCODE_SERVER_PORT} "${PEER}"
packageChaincode "${DEPLOYMENT_PATH}" "${CC_ID}" "${CC_VER}" "external" ${CHAINCODE_SERVER_PORT} "${PEER}"
