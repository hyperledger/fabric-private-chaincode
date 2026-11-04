#!/usr/bin/env bash

# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

# abort on error
set -e

build_dir=$1
enclave_dir=$2

hex_out_name=enclave_hash.hex
hex_string_out_name=mrenclave
tmp_name=tmp_enc_hash
go_name=mrenclave.go

cd $build_dir
sgx_sign gendata -enclave enclave.so -config $enclave_dir/enclave.config.xml -out $tmp_name
dd if=$tmp_name bs=1 skip=188 of=$hex_out_name count=32
hex --upper-case $hex_out_name > $hex_string_out_name
rm $hex_out_name
rm $tmp_name
echo "Enclave hash extracted."
cat $hex_string_out_name

echo "Create go file"
cat > $go_name << EOF
/*
 * Copyright 2019 Intel Corporation
 * Copyright IBM Corp. All Rights Reserved.

 */

package enclave

// MrEnclave contains hash of enclave code
const MrEnclave = "$(cat $hex_string_out_name)"
EOF

