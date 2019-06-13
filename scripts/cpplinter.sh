#!/bin/bash

# Copyright Intel Corp. 2019 All Rights Reserved.
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

SCRIPT=`realpath $0`
SCRIPTDIR=`dirname $SCRIPT`

FILES_TO_PARSE=`find $SCRIPTDIR/.. -iname *.cpp -o -iname *.h -o -iname *.c`

RET=0

for FILE in $FILES_TO_PARSE; do

    #skip file checking for certain folders
    echo $FILE | grep 'ecc_enclave/_build'  > /dev/null && continue
    echo $FILE | grep 'ecc_enclave/build'   > /dev/null && continue
    echo $FILE | grep 'tlcc_enclave/nanopb' > /dev/null && continue
    echo $FILE | grep 'tlcc_enclave/_build' > /dev/null && continue
    echo $FILE | grep 'tlcc_enclave/build'  > /dev/null && continue
    echo $FILE | grep 'common/protobuf'     > /dev/null && continue
    echo $FILE | grep 'common/base64'       > /dev/null && continue
    echo $FILE | grep 'common/json'         > /dev/null && continue

    if [[ $1 == 'DO_FORMAT' ]]
    then
        clang-format -i $FILE
    else

        clang-format $FILE -output-replacements-xml | grep "</replacement>" > /dev/null &&\
            echo "ERROR in format: $FILE" && RET=1
    fi
done

if [[ $RET != 0 ]]
then
    echo "Format check failed. Run '$0 DO_FORMAT' to fix the format."
fi

exit $RET
