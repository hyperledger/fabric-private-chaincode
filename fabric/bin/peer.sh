#!/bin/bash
# Copyright Intel Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

#RUN=echo # uncomment to dry-run peer call

SCRIPTDIR="$(dirname $(readlink --canonicalize ${BASH_SOURCE}))"
FPC_TOP_DIR="${SCRIPTDIR}/../../"
CONFIG_HOME="$(pwd)"

. ${SCRIPTDIR}/lib/common_ledger.sh

FPC_DOCKER_NAME_CMD="${FPC_TOP_DIR}/utils/fabric/get-fabric-container-name"

export LD_LIBRARY_PATH=${LD_LIBRARY_PATH:+"${LD_LIBRARY_PATH}:"}${GOPATH}/src/github.com/hyperledger-labs/fabric-private-chaincode/tlcc/enclave/lib 


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
        -e|--enclavesopath)
        CC_ENCLAVESOPATH=$2
        shift;shift
        ;;
	    *)
		OTHER_ARGS+=( "$1" )
		shift
		;;
	    esac
    done
    if [ "${CC_LANG}" = "fpc-c" ]; then
	if [ -z ${CC_NAME+x} ] || [ -z ${CC_VERSION+x} ] || [ -z ${CC_PATH+x} ] || [ -z ${CC_ENCLAVESOPATH+x} ] ]; then
	    # missing params, don't do anything and let real peer report errors
	    return
	fi
	yell "Found valid FPC Chaincode Install!"
	# get net-id and peer-id from core.yaml, for now just use defaults ...
	parse_fabric_config .

	# get docker name for this chaincode
	DOCKER_IMAGE_NAME=$(${FPC_DOCKER_NAME_CMD} --cc-name ${CC_NAME} --cc-version ${CC_VERSION} --net-id ${NET_ID} --peer-id ${PEER_ID}) || die "could not get docker image name"

	# install docker
	# TODO: eventually use path to select actual chaincode once Bruno's SDKization is in ...
	try make ENCLAVE_SO_PATH=${CC_ENCLAVESOPATH} DOCKER_IMAGE=${DOCKER_IMAGE_NAME} -C ${FPC_TOP_DIR}/ecc docker-fpc-app

	# eplace path and lang arg with dummy go chaincode
	ARGS_EXEC=( 'chaincode' 'install' '-n' "${CC_NAME}" '-v' "${CC_VERSION}" '-p' 'github.com/hyperledger/fabric/examples/chaincode/go/example02/cmd' "${OTHER_ARGS[@]}" )	
    fi
    return
}

handle_channel_create() {
    # - get channel name
    while [[ $# > 0 ]]; do
	case "$1" in
	    -c|--channelID)
		CHAN_ID=$2;
		shift; shift
		;;
	    *)
		shift
		;;
	    esac
    done
    # - remember that we are "channel creation" peer
    try touch "${CONFIG_HOME}/${CHAN_ID}.creator"
    # fall through to "real" peer ...
    # (Note: we didn't modify args, so we can use original ones already storeed in ARGS_EXEC)
    return
}

handle_channel_join() {
    # - get channel name
    #   (we rely here on convention that block is named ${CHAN_ID}.block
    #   as channel id is not explicitly passed as argument!)
    while [[ $# > 0 ]]; do
	case "$1" in
	    -b|--blockpath)
		CHAN_BLOCK=$2;
		shift; shift
		;;
	    *)
		shift
		;;
	    esac
    done
    CHAN_ID=$(basename -s .block ${CHAN_BLOCK}) || die "Cannot derive channel id from block param '$CHAN_BLOCK}'"
    yell "Deriving channel id '${CHAN_ID}' from channel block file '${CHAN_BLOCK}', relying on naming convention '..../<chan_id>.block' for that file!"

    # - call real peer so channel is joined
    try $RUN ${FABRIC_BIN_DIR}/peer "${ARGS_EXEC[@]}"

    # - handle ercc
    say "Installing & Instantiating ercc on channel '${CHAN_ID}' ..."
    #   - install ercc
    try $RUN ${FABRIC_BIN_DIR}/peer chaincode install -n ${ERCC_ID} -v ${ERCC_VERSION} -p github.com/hyperledger-labs/fabric-private-chaincode/ercc/cmd
    sleep 1
    #   - instantiate ercc iff "channel creation" peer
    if [ -e "${CONFIG_HOME}/${CHAN_ID}.creator" ]; then
	try $RUN ${FABRIC_BIN_DIR}/peer chaincode instantiate -n ${ERCC_ID} -v ${ERCC_VERSION} -c '{"args":["init"]}' -C ${CHAN_ID} -V ercc-vscc
	sleep 3
	try rm "${CONFIG_HOME}/${CHAN_ID}.creator"
    fi
    #   - get SPID (mostly as debug output)
    try $RUN ${FABRIC_BIN_DIR}/peer chaincode query -n ${ERCC_ID} -c '{"args":["getSPID"]}' -C ${CHAN_ID}
    sleep 3

    # - ask tlcc to join channel
    #   IMPORTANT: right now a join is _not_ persistant, so on restart of peer,
    #   it will re-join old channels but tlcc will not!
    say "Attaching TLCC to channel '${CHAN_ID}' ..."
    try $RUN ${FABRIC_BIN_DIR}/peer chaincode query -n tlcc -c '{"Args": ["JOIN_CHANNEL"]}' -C ${CHAN_ID}

    # - exit
    exit 0
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

    channel)
	shift
	case "$1" in
	    create)
		shift
		handle_channel_create "$@"
		;;
	    join)
		shift
		handle_channel_join "$@"
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
try $RUN exec ${FABRIC_BIN_DIR}/peer "${ARGS_EXEC[@]}"
