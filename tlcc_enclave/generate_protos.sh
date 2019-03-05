#!/bin/bash

# set -eux

FABRIC_PATH=${FABRIC_PATH-~/fabric}
NANOPB_PATH=${NANOPB_PATH-~/nanopb}

PROTOC_OPTS="--plugin=protoc-gen-nanopb=$NANOPB_PATH/generator/protoc-gen-nanopb"

FABRIC_PROTOS=$FABRIC_PATH/protos

BUILD_DIR=enclave/protos
rm -rf $BUILD_DIR
mkdir -p $BUILD_DIR

# compile google protos (timestamp)
$(protoc "$PROTOC_OPTS" --proto_path="protos" --nanopb_out=$BUILD_DIR protos/google/protobuf/*.proto)

declare -a arr=("common" "ledger" "msp" "peer" "token")

## now loop through the above array
for i in "${arr[@]}"
do
    # compile fabric protos
    for protos in $(find "$FABRIC_PROTOS" -name '*.proto' -path */$i/* -exec dirname {} \; | sort | uniq) ; do
        echo $protos
        $(protoc "$PROTOC_OPTS" --proto_path=protos --proto_path="$BUILD_DIR" --proto_path="$FABRIC_PROTOS" "--nanopb_out=-f protos/fabric.options:$BUILD_DIR" "$protos"/*.proto)
    done
done

# fix enclave/protos/ledger/rwset/rwset.pb.h
sed  -i 's/namespace/ns/g' enclave/protos/ledger/rwset/rwset.pb.h 
sed  -i 's/namespace/ns/g' enclave/protos/ledger/rwset/rwset.pb.c
