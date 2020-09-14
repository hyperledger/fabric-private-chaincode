#!/bin/bash

# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

set -e

###########################################################
# enclave_to_mrenclave
#   input:  non-signed enclave file path, enclave configuration file as parameters
#   output: MRENCLAVE variable
###########################################################
function enclave_to_mrenclave() {
    if [[ ! -f $1 ]]; then
        echo "missing enclave file path"
        exit -1
    fi
    if [[ ! -f $2 ]]; then
        echo "missing enclave configuration file path"
        exit -1
    fi

    TMP1=$(mktemp) 
    TMP2=$(mktemp) 
    sgx_sign gendata -enclave $1 -config $2 -out $TMP1
    dd if=$TMP1 bs=1 skip=188 of=$TMP2 count=32
    MRENCLAVE=$(hex -c $TMP2)
}
