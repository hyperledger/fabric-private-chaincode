#!/bin/bash

# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

# set -eux

if [[ -z "${NANOPB_PATH}"  ]]; then
    echo "Error: NANOPB_PATH not set"
    exit 1
else
    NANOPB_PATH=${NANOPB_PATH}
fi

PROTOC_OPTS="--plugin=protoc-gen-nanopb=$NANOPB_PATH/generator/protoc-gen-nanopb"

FABRIC_PROTOS=protos/fabric

if [ "$1" != "" ]; then
    BUILD_DIR=$1
else
    BUILD_DIR=enclave/protos
fi
rm -rf $BUILD_DIR
mkdir -p $BUILD_DIR

# compile google protos (timestamp)
protoc "$PROTOC_OPTS" --proto_path="protos" --nanopb_out=$BUILD_DIR protos/google/protobuf/*.proto

declare -a arr=("common" "ledger" "msp" "peer")

## now loop through the above array
for i in "${arr[@]}"
do
    # compile fabric protos
    for protos in $(find "$FABRIC_PROTOS" -name '*.proto' -path */$i/* -exec dirname {} \; | sort | uniq) ; do
        protoc "$PROTOC_OPTS" --proto_path=protos/google --proto_path="$BUILD_DIR" --proto_path="$FABRIC_PROTOS" "--nanopb_out=-f  protos/fabric.options:$BUILD_DIR" "$protos"/*.proto
    done
done

# fix enclave/protos/ledger/rwset/rwset.pb.h
sed  -i 's/namespace/ns/g' enclave/protos/ledger/rwset/rwset.pb.h
sed  -i 's/namespace/ns/g' enclave/protos/ledger/rwset/rwset.pb.c
