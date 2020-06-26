#!/bin/bash
#
# Copyright IBM Corp. All Rights Reserved.
# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0
#

# download the files to a consistent place
export SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd ${SCRIPT_DIR}/..

# Versions for Dependencies
# Notes:
# - as some of script downloads binary which does require full semantic version, we cannot only
#   specify major and minor version and so ask implicitly for latest patch version. 
# - We also allow averride by caller
# * Fabric binaries and images
export FABRIC_VERSION=${FABRIC_VERSION:=2.1.1}
export FABRIC_TAG=${FABRIC_TAG:="$FABRIC_VERSION"}
# * CA
export CA_VERSION=${CA_VERSION:=1.4.7}
export CA_TAG=${CA_TAG:="$CA_VERSION"}
# * thirdparty images (couchdb, kafka and zookeeper) 
export THIRDPARTY_IMAGE_VERSION=${THIRDPARTY_IMAGE_VERSION:=0.4.20}
export THIRDPARTY_TAG=${THIRDPARTY_TAG:="$THIRDPARTY_IMAGE_VERSION"}

export ARCH=$(echo "$(uname -s|tr '[:upper:]' '[:lower:]'|sed 's/mingw64_nt.*/windows/')-$(uname -m | sed 's/x86_64/amd64/g')")
export MARCH=$(uname -m)



printHelp() {
  echo "Usage: bootstrap.sh [<version>] [<ca_version>] [<thirdparty_version>][-d -b]"
  echo
  echo "-d - bypass docker image download"
  echo "-b - bypass download of platform-specific binaries"
  echo
  echo "e.g. bootstrap.sh 2.1 1.4 0.4.15"
  echo "would download docker images and binaries for version 2.1 (fabric) 1.4 (fabric-ca) 0.4.15 (thirdparty)"
}

# dockerFabricPull() pulls docker images from fabric and chaincode repositories
# note, if a docker image doesn't exist for a requested release, it will simply
# be skipped, since this script doesn't terminate upon errors.
dockerFabricPull() {
  local FABRIC_TAG=$1
  for IMAGES in peer orderer ccenv tools; do
      echo "==> FABRIC IMAGE: $IMAGES"
      echo
      docker pull hyperledger/fabric-$IMAGES:$FABRIC_TAG
      docker tag hyperledger/fabric-$IMAGES:$FABRIC_TAG hyperledger/fabric-$IMAGES
  done
}

dockerThirdPartyImagesPull() {
  local THIRDPARTY_TAG=$1
  for IMAGES in couchdb kafka zookeeper; do
      echo "==> THIRDPARTY DOCKER IMAGE: $IMAGES"
      echo
      docker pull hyperledger/fabric-$IMAGES:$THIRDPARTY_TAG
      docker tag hyperledger/fabric-$IMAGES:$THIRDPARTY_TAG hyperledger/fabric-$IMAGES
  done
}

dockerCaPull() {
      local CA_TAG=$1
      echo "==> FABRIC CA IMAGE"
      echo
      docker pull hyperledger/fabric-ca:$CA_TAG
      docker tag hyperledger/fabric-ca:$CA_TAG hyperledger/fabric-ca
}

# Incrementally downloads the .tar.gz file locally first, only decompressing it
# after the download is complete. This is slower than binaryDownload() but
# allows the download to be resumed.
binaryIncrementalDownload() {
      local BINARY_FILE=$1
      local URL=$2
      curl -L -f -s -C - ${URL} -o ${BINARY_FILE} || rc=$?
      # Due to limitations in the current Nexus repo:
      # curl returns 33 when there's a resume attempt with no more bytes to download
      # curl returns 2 after finishing a resumed download
      # with -f curl returns 22 on a 404
      if [ "$rc" = 22 ]; then
	  # looks like the requested file doesn't actually exist so stop here
	  return 22
      fi
      if [ -z "$rc" ] || [ $rc -eq 33 ] || [ $rc -eq 2 ]; then
          # The checksum validates that RC 33 or 2 are not real failures
          echo "==> File downloaded. Verifying the md5sum..."
          localMd5sum=$(md5sum ${BINARY_FILE} | awk '{print $1}')
          remoteMd5sum=$(curl -L -s ${URL}.md5)
          if [ "$localMd5sum" == "$remoteMd5sum" ]; then
              echo "==> Extracting ${BINARY_FILE}..."
              tar xzf ./${BINARY_FILE} --overwrite bin
	      echo "==> Done."
              rm -f ${BINARY_FILE} ${BINARY_FILE}.md5
          else
              echo "Download failed: the local md5sum is different from the remote md5sum. Please try again."
              rm -f ${BINARY_FILE} ${BINARY_FILE}.md5
              exit 1
          fi
      else
          echo "Failure downloading binaries (curl RC=$rc). Please try again and the download will resume from where it stopped."
          exit 1
      fi
}

# This will attempt to download the .tar.gz all at once, but will trigger the
# binaryIncrementalDownload() function upon a failure, allowing for resume
# if there are network failures.
binaryDownload() {
      local BINARY_FILE=$1
      local URL=$2
      echo "===> Downloading: " ${URL}
      # Check if a previous failure occurred and the file was partially downloaded
      if [ -e ${BINARY_FILE} ]; then
          echo "==> Partial binary file found. Resuming download..."
          binaryIncrementalDownload ${BINARY_FILE} ${URL}
      else
          curl -L ${URL} | tar xz bin || rc=$?
          if [ ! -z "$rc" ]; then
              echo "==> There was an error downloading the binary file. Switching to incremental download."
              echo "==> Downloading file..."
              binaryIncrementalDownload ${BINARY_FILE} ${URL}
	  else
	      echo "==> Done."
          fi
      fi
}

binariesInstall() {
   local FABRIC_BINARY_FILE=$1
   local CA_BINARY_FILE=$2
  echo "===> Downloading version ${FABRIC_TAG} platform specific fabric binaries"
  binaryDownload "${FABRIC_BINARY_FILE}" "https://github.com/hyperledger/fabric/releases/download/v${FABRIC_VERSION}/${FABRIC_BINARY_FILE}"
  if [ $? -eq 22 ]; then
     echo
     echo "------> ${FABRIC_TAG} platform specific fabric binary is not available to download <----"
     echo
   fi

  echo "===> Downloading version ${CA_TAG} platform specific fabric-ca-client binary"
  binaryDownload "${CA_BINARY_FILE}" "https://github.com/hyperledger/fabric-ca/releases/download/v${CA_VERSION}/${CA_BINARY_FILE}"
  if [ $? -eq 22 ]; then
     echo
     echo "------> ${CA_TAG} fabric-ca-client binary is not available to download <----"
     echo
   fi
}

dockerInstall() {
  which docker >& /dev/null
  NODOCKER=$?
  if [ "${NODOCKER}" == 0 ]; then
	  echo "===> Pulling fabric Images"
	  dockerFabricPull ${FABRIC_TAG}
	  echo "===> Pulling fabric ca Image"
	  dockerCaPull ${CA_TAG}
	  echo "===> Pulling thirdparty docker images"
	  dockerThirdPartyImagesPull ${THIRDPARTY_TAG}
	  echo
	  echo "===> List out hyperledger docker images"
	  docker images | grep hyperledger*
  else
    echo "========================================================="
    echo "Docker not installed, bypassing download of Fabric images"
    echo "========================================================="
  fi
}



# command-line option defaults
DOCKER=true
SAMPLES=true
BINARIES=true

# then parse opts
while getopts "hdb" opt; do
  case "$opt" in
    h|\?)
      printHelp
      exit 0
    ;;
    d)  DOCKER=false
    ;;
    b)  BINARIES=false
    ;;
  esac
done

if [ "$BINARIES" == "true" ]; then
  echo
  echo "Installing Hyperledger Fabric binaries"
  echo
  binariesInstall "hyperledger-fabric-${ARCH}-${FABRIC_VERSION}.tar.gz" "hyperledger-fabric-ca-${ARCH}-${CA_VERSION}.tar.gz"
fi
if [ "$DOCKER" == "true" ]; then
  echo
  echo "Installing Hyperledger Fabric docker images"
  echo
  dockerInstall
fi
