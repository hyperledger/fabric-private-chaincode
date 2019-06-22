#!/bin/bash
# Copyright Intel Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

#RUN=echo # uncomment to dry-run peer call

SCRIPTDIR="$(dirname $(readlink --canonicalize ${BASH_SOURCE}))"
. ${SCRIPTDIR}/lib/common_utils.sh

FPC_TOP_DIR="${SCRIPTDIR}/../../"
FPC_DOCKER_NAME_CMD="${FPC_TOP_DIR}/utils/fabric/get-fabric-container-name"

: ${FABRIC_PATH:="${FPC_TOP_DIR}/../../hyperledger/fabric/"}
: ${FABRIC_BIN_DIR:="${FABRIC_PATH}/.build/bin"}

handle_chaincode_install() {
    OTHER_ARGS=()
    while [[ $# > 0 ]]; do
	case "$1" in
	    -n|--name)
		CC_NAME=$2;
		shift; shift
		;;
	    -v|--version)
		CC_VERSION=$2
		shift; shift
		;;
	    -p|--path)
		CC_PATH=$2
		shift; shift
		;;
	    -l|--lang)
		CC_LANG=$2
		shift;shift
		;;
	    *)
		OTHER_ARGS+=( "$1" )
		shift
		;;
	    esac
    done
    if [ "${CC_LANG}" = "fpc-c" ]; then
	if [ -z ${CC_NAME+x} ] || [ -z ${CC_VERSION+x} ] || [ -z ${CC_PATH+x} ]; then
	    # missing params, don't do anything and let real peer report errrors
	    return
	fi
	yell "Found valid FPC Chaincode Install!"
	# get net-id and peer-id from core.yaml, for now just use defaults ...
	parse_fabric_config .

	# get docker name for this chaincode
	DOCKER_IMAGE_NAME=$(${FPC_DOCKER_NAME_CMD} --cc-name ${CC_NAME} --cc-version ${CC_VERSION} --net-id ${NET_ID} --peer-id ${PEER_ID}) || die "could not get docker image name"

	# install docker
	# TODO: eventually use path to select actual chaincode once Bruno's SDKization is in ...
	try make DOCKER_IMAGE=${DOCKER_IMAGE_NAME} -C ${FPC_TOP_DIR}/ecc docker

	# eplace path and lang arg with dummy go chaincode
	ARGS_EXEC=( 'chaincode' 'install' '-n' "${CC_NAME}" '-v' "${CC_VERSION}" '-p' 'github.com/hyperledger/fabric/examples/chaincode/go/example02/cmd' "${OTHER_ARGS[@]}" )	
    fi
    return
}

handle_channel_create() {
    # TODO: remember that we are "channel creation" peer
    # for now just fall through ...
    return
}

handle_channel_join() {
    # TODO: install and (if "channel creation" peer) instantiate ercc
    # for now just fall through ...
    return
}

# - check whether it is a command which we have to intercept
#   - chaincode install
#   - channel create
#   - channel join
ARGS_EXEC=( "$@" ) # params to eventually pass to real peer /default: just pass all original args ..
case "$1" in
    chaincode)
	shift
	case "$1" in
	    install)
		shift
		handle_chaincode_install "$@"
		;;
	    *)
		# fall through, nothing to do
	esac

	;;
    *) 
	# fall through, nothing to do
	;;
esac

# Call real peer
export LD_LIBRARY_PATH=${LD_LIBRARY_PATH:+"${LD_LIBRARY_PATH}:"}${GOPATH}/src/github.com/hyperledger-labs/fabric-private-chaincode/tlcc/enclave/lib 
try $RUN exec ${FABRIC_BIN_DIR}/peer "${ARGS_EXEC[@]}"
