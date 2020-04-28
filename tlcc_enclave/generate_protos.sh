#!/bin/bash

# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

set -e

if [[ -z "${NANOPB_PATH}"  ]]; then
    echo "Error: NANOPB_PATH not set"
    exit 1
else
    NANOPB_PATH=${NANOPB_PATH}
fi

if [[ -z "${PROTOC_CMD}"  ]]; then
    echo "Error: PROTOC_CMD not set"
    exit 1
fi

FABRIC_PROTOS=protos/fabric
# check that fabric protos are present
if [[ -z $(find ${FABRIC_PROTOS} -name '*.proto') ]]; then \
    echo "No Fabric protos found! Try 'git pull --recurse-submodules'"; exit 1; \
fi

if [ "$1" != "" ]; then
    BUILD_DIR=$1
else
    BUILD_DIR=enclave/protos
fi
rm -rf $BUILD_DIR
mkdir -p $BUILD_DIR

PROTOC_OPTS="--plugin=protoc-gen-nanopb=$NANOPB_PATH/generator/protoc-gen-nanopb"

# compile google protos (timestamp)
$PROTOC_CMD "$PROTOC_OPTS" --proto_path="protos" --nanopb_out=$BUILD_DIR protos/google/protobuf/*.proto

declare -a arr=("common" "ledger" "msp" "peer")

## now loop through the above array
for i in "${arr[@]}"
do
    # compile fabric protos
    for protos in $(find "$FABRIC_PROTOS" -name '*.proto' -path */$i/* -exec dirname {} \; | sort | uniq) ; do
        $PROTOC_CMD "$PROTOC_OPTS" --proto_path=protos/google --proto_path="$BUILD_DIR" --proto_path="$FABRIC_PROTOS" "--nanopb_out=-f  protos/fabric.options:$BUILD_DIR" "$protos"/*.proto
    done
done

# fix enclave/protos/ledger/rwset/rwset.pb.h
sed  -i 's/namespace/ns/g' enclave/protos/ledger/rwset/rwset.pb.h
sed  -i 's/namespace/ns/g' enclave/protos/ledger/rwset/rwset.pb.c
