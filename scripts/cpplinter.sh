#!/bin/bash

# Copyright Intel Corp. 2019 All Rights Reserved.
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

SCRIPT=`realpath $0`
SCRIPTDIR=`dirname $SCRIPT`

FILES_TO_PARSE=`find $SCRIPTDIR/.. -iname *.cpp -o -iname *.h -o -iname *.c`

RET=0

EXCLUDED_PATTERNS="\
    ecc_enclave/_build \
    tlcc_enclave/nanopb \
    tlcc_enclave/_build \
    examples/auction/_build \
    examples/echo/_build \
    common/protobuf \
    common/base64 \
    common/json \
    enclave_u.c \
    enclave_u.h \
    enclave_t.c \
    enclave_t.h \
    .pb. \
    "
for FILE in $FILES_TO_PARSE; do

    #skip file checking for certain folders
    for EP in $EXCLUDED_PATTERNS; do
        echo $FILE | grep -F $EP  > /dev/null && continue 2
    done

    #do format (if specified) or simply check format
    if [[ $1 == 'DO_FORMAT' ]]
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
    echo "Format check failed. Run '$0 DO_FORMAT' to fix the format."
fi

exit $RET
