#!/bin/bash

# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

set -e

if [[ -z "${FPC_PATH}"  ]]; then
    echo "Error: FPC_PATH not set"
    exit 1
else
    FPC_PATH=${FPC_PATH}
fi

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

PROTOS_DIR=${FPC_PATH}/protos

FABRIC_PROTOS_DIR=${PROTOS_DIR}/fabric
# check that fabric protos are present
if [[ -z $(find ${FABRIC_PROTOS_DIR} -name '*.proto') ]]; then \
    echo "No Fabric protos found! Try 'git pull --recurse-submodules'"; exit 1; \
fi

if [ "$1" != "" ]; then
    BUILD_DIR=$1
else
    BUILD_DIR=${FPC_PATH}/common/protos
fi
rm -rf $BUILD_DIR
mkdir -p $BUILD_DIR

FABRIC_BUILD_DIR=${BUILD_DIR}/fabric
mkdir -p $FABRIC_BUILD_DIR

# create go proto folder and exclude it from staticcheck
GO_BUILD_DIR=${FPC_PATH}/internal/protos
mkdir -p $GO_BUILD_DIR
cat <<EOT > "${GO_BUILD_DIR}/staticcheck.conf"
# ignore all checks in this package
checks = ["-all"]
EOT

PROTOC_OPTS="--plugin=protoc-gen-nanopb=$NANOPB_PATH/generator/protoc-gen-nanopb-py2"

# compile google protos
$PROTOC_CMD "$PROTOC_OPTS" --proto_path=${PROTOS_DIR} --nanopb_out=$BUILD_DIR ${PROTOS_DIR}/google/protobuf/*.proto

# compile fabric protos
declare -a arr=("common" "ledger" "msp" "peer")

for i in "${arr[@]}"
do
    # filter fabric protos
    for protos in $(find "$FABRIC_PROTOS_DIR" -name '*.proto' -path "*/$i/*" -exec dirname {} \; | sort | uniq) ; do
        $PROTOC_CMD "$PROTOC_OPTS" --proto_path="${PROTOS_DIR}/protos/google" --proto_path="$BUILD_DIR" --proto_path="$FABRIC_PROTOS_DIR" "--nanopb_out=-f  ${PROTOS_DIR}/fabric.options:$FABRIC_BUILD_DIR" "$protos"/*.proto
    done
done

# fix enclave/protos/ledger/rwset/rwset.pb.h
sed  -i 's/namespace/ns/g' ${FABRIC_BUILD_DIR}/ledger/rwset/rwset.pb.h
sed  -i 's/namespace/ns/g' ${FABRIC_BUILD_DIR}/ledger/rwset/rwset.pb.c

# compile fpc protos
$PROTOC_CMD "$PROTOC_OPTS" --proto_path=${PROTOS_DIR} --proto_path=${FABRIC_PROTOS_DIR} "--nanopb_out=-f ${PROTOS_DIR}/fpc.options:${BUILD_DIR}" ${PROTOS_DIR}/fpc/fpc.proto 
$PROTOC_CMD "$PROTOC_OPTS" --proto_path=${PROTOS_DIR} --proto_path=${FABRIC_PROTOS_DIR} --go_out=${GOPATH}/src ${PROTOS_DIR}/fpc/*.proto
