#!/bin/bash

# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

if [ $# -eq 0 ] || [ -z $1 ]; then
    echo "Missing top folder input parameter for $0"
    exit 1
fi

FILES_TO_PARSE=`find $1 -iname *.cpp -o -iname *.h -o -iname *.c`

RET=0

EXCLUDED_PATTERNS="\
    common/protobuf \
    common/base64 \
    common/json \
    enclave_u.c \
    enclave_u.h \
    enclave_t.c \
    enclave_t.h \
    /_build \
    .pb. \
    /node_modules \
    common/crypto/pdo \
    "
for FILE in $FILES_TO_PARSE; do

    #skip file checking for certain folders
    for EP in $EXCLUDED_PATTERNS; do
        echo $FILE | grep -F $EP  > /dev/null && continue 2
    done

    #do format (if specified) or simply check format
    if [[ $2 == 'DO_FORMAT' ]]
    then
        clang-format -i $FILE
    else

        clang-format $FILE -output-replacements-xml | grep "</replacement>" > /dev/null &&\
            echo "ERROR in format: $FILE" && RET=1
    fi
done

#if check fails, provide instructions for fixing the format
if [[ $RET != 0 ]]
then
    echo "Format check failed. Run '$0 <top folder> DO_FORMAT' to fix the format."
fi

exit $RET
