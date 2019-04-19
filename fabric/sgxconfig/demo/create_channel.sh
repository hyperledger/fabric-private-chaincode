#!/bin/bash
SCRIPTDIR="$(dirname $(readlink --canonicalize ${BASH_SOURCE}))"
. ${SCRIPTDIR}/common.sh

${FABRIC_BIN_DIR}/configtxgen -channelID mychannel -profile SampleSingleMSPChannel -outputCreateChannelTx mychannel.tx
