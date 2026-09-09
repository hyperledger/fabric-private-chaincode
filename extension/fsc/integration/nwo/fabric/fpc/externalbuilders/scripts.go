/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package externalbuilders

// This package provides the external builder scripts for FPC.
// The original scripts can be found at https://github.com/hyperledger/fabric-private-chaincode/tree/main/fabric/externalBuilder/chaincode_server/bin

const (
	Build = `#!/bin/bash

# Copyright 2020 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

SOURCE=$1
OUTPUT=$3

#external chaincodes expect connection.json file in the chaincode package
if [ ! -f "$SOURCE/${CORE_PEER_ID}/connection.json" ]; then
    >&2 echo "$SOURCE/connection.json not found"
    exit 1
fi

#simply copy the endpoint information to specified output location
cp $SOURCE/${CORE_PEER_ID}/connection.json $OUTPUT/connection.json

if [ -d "$SOURCE/metadata" ]; then
    cp -a $SOURCE/metadata $OUTPUT/metadata
fi

exit 0
`

	Detect = `#!/bin/bash

# Copyright 2020 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

METADIR=$2
# check if the "type" field is set to "external"
# crude way without jq which is not in the default fabric peer image
TYPE=$(tr -d '\n' < "$METADIR/metadata.json" | awk -F':' '{ for (i = 1; i < NF; i++){ if ($i~/type/) { print $(i+1); break }}}'| cut -d\" -f2)

if [ "$TYPE" = "external" ]; then
    exit 0
fi

exit 1
`

	Release = `#!/bin/bash

# Copyright 2020 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

BLD="$1"
RELEASE="$2"

if [ -d "$BLD/metadata" ]; then
   cp -a "$BLD/metadata/"* "$RELEASE/"
fi

#external chaincodes expect artifacts to be placed under "$RELEASE"/chaincode/server
if [ -f $BLD/connection.json ]; then
   mkdir -p "$RELEASE"/chaincode/server
   cp $BLD/connection.json "$RELEASE"/chaincode/server

   #if tls_required is true, copy TLS files (using above example, the fully qualified path for these fils would be "$RELEASE"/chaincode/server/tls)

   exit 0
fi

exit 1
`
)
