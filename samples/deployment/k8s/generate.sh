#!/usr/bin/env bash

# Copyright IBM Corp. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

C_RESET='\033[0m'
C_RED='\033[0;31m'
C_GREEN='\033[0;32m'
C_BLUE='\033[0;34m'
C_YELLOW='\033[1;33m'
# successln echos in green color
function successln() {
  println "${C_GREEN}${1}${C_RESET}"
}
# println echos string
function println() {
  echo -e "$1"
}
set -euo pipefail

if [[ -z "${FPC_PATH}" ]]; then
  echo "Error: FPC_PATH not set"
  exit 1
fi

. "${FPC_PATH}"/utils/packaging/utils.sh

channelID="mychannel"

outDir="."
cryptoConfigDir="${outDir}/crypto-config"
channelArtifactsDir="${outDir}/channel-artifacts"
packageDir="${outDir}/packages"

CHAINCODE_SERVER_PORT=9999

FABRIC_BIN_DIR="${FPC_PATH}/fabric/_internal/bin"
if [[ -z "${FABRIC_BIN_DIR}" ]]; then
  echo "Error: FABRIC_BIN_DIR not set"
  echo "Error: FABRIC_BIN_DIR must point to the location of cryptogen and configtxgen"
  exit 1
fi

CRYPPTOGEN_CMD="${FABRIC_BIN_DIR}/cryptogen"
CONFIGTXGEN_CMD="${FABRIC_BIN_DIR}/configtxgen"

echo "Clean existing deployment artifacts"
rm -rf ${cryptoConfigDir}
rm -rf ${channelArtifactsDir}
rm -rf ${packageDir}

echo "Generate crypto material for orgs"
$CRYPPTOGEN_CMD generate --output=${cryptoConfigDir} --config=./crypto-config.yaml

echo "Generate genesis block"
$CONFIGTXGEN_CMD -profile DemoGenesis -channelID testchainid -outputBlock ${channelArtifactsDir}/genesis.block
$CONFIGTXGEN_CMD -profile DemoChannel -outputCreateChannelTx ${channelArtifactsDir}/channel.tx -channelID ${channelID}

echo "Generate client connection profile"
function yaml_connection {
    sed -e "s/{{ORG}}/$1/g" \
        connection-template.yaml | sed -e $'s/\\\\n/\\\n          /g'
}
for org in $(shopt -s globstar; find ${cryptoConfigDir}/**/peerOrganizations/ -mindepth 1 -maxdepth 1 -execdir echo {} ';' | sed 's/^\.\///g');
do
  echo "$(yaml_connection $org)" > ${cryptoConfigDir}/peerOrganizations/${org}/connection.yaml
done

echo "Package ercc and fpccc"
function packageChaincode() {

  address="{{.peername}}_${CC_NAME}_ccaas:${CHAINCODE_SERVER_PORT}"
  prefix=$(basename "$0")
  tempdir=$(mktemp -d -t "$prefix.XXXXXXXX") || error_exit "Error creating temporary directory"
  label=${CC_NAME}_${CC_VERSION}
  mkdir -p "$tempdir/src"

cat > "$tempdir/src/connection.json" <<CONN_EOF
{
  "address": "${address}",
  "dial_timeout": "10s",
  "tls_required": false
}
CONN_EOF

   mkdir -p "$tempdir/pkg"

cat << METADATA-EOF > "$tempdir/pkg/metadata.json"
{
    "type": "ccaas",
    "label": "$label"
}
METADATA-EOF

    tar -C "$tempdir/src" -czf "$tempdir/pkg/code.tar.gz" .
    tar -C "$tempdir/pkg" -czf "$CC_NAME.tar.gz" metadata.json code.tar.gz
    rm -Rf "$tempdir"

#    PACKAGE_ID=$(peer lifecycle chaincode calculatepackageid ${CC_NAME}.tar.gz)
# commenting due to the network not yet being setup to calculate package_id 
  
    successln "Chaincode is packaged  ${address}"
}

CC_TYPE="ccaas"
ERCC_ID="ercc"
ERCC_VER="1.0"
FPCCC_ID="fpccc"
FPCCC_PATH=${TEST_CC_PATH}
if [[ -z "${TEST_CC_PATH}" ]]; then
  echo "Error: TEST_CC_PATH not set"
  echo "Error: TEST_CC_PATH must point to the location FPC Chaincode"
  exit 1
fi

FPC_MRENCLAVE="$(cat "${FPCCC_PATH}"/_build/lib/mrenclave)"

for peer in $(shopt -s globstar; find ${cryptoConfigDir}/**/peers/ -mindepth 1 -maxdepth 1 -execdir echo {} ';' | sed 's/^\.\///g');
do
    # ercc
    CC_NAME=ERCC_ID
    CC_VERSION=ERCC_VER
#    endpoint="${ERCC_ID}-${peer}:${CHAINCODE_SERVER_PORT}"
#    packageName="${ERCC_ID}-${peer}.tgz"
#    packageChaincode "${packageDir}" "${packageName}" "${ERCC_ID}" "${ERCC_VER}" "${CC_TYPE}" "${endpoint}" "${peer}"
    packageChaincode "${peer}"

    # fpc cc
    CC_NAME=FPCCC_ID
#    endpoint="${FPCCC_ID}-${peer}:${CHAINCODE_SERVER_PORT}"
#    packageName="${FPCCC_ID}-${peer}.tgz"
#    packageChaincode "${packageDir}" "${packageName}" "${FPCCC_ID}" "${FPC_MRENCLAVE}" "${CC_TYPE}" "${endpoint}" "${peer}"
    packageChaincode "${peer}"
done

#echo "Store mrenclave for fpccc"
#echo "FPC_MRENCLAVE=${FPC_MRENCLAVE}" >> ${packageDir}/chaincode-config.properties
