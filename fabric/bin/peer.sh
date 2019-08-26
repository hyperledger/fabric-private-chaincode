#!/bin/bash
# Copyright Intel Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

#RUN=echo # uncomment to dry-run peer call

SCRIPTDIR="$(dirname $(readlink --canonicalize ${BASH_SOURCE}))"
FPC_TOP_DIR="${SCRIPTDIR}/../../"
FABRIC_SCRIPTDIR="${FPC_TOP_DIR}/fabric/bin/"

: ${FABRIC_CFG_PATH:=$(pwd)}

. ${FABRIC_SCRIPTDIR}/lib/common_utils.sh
. ${FABRIC_SCRIPTDIR}/lib/common_ledger.sh

FPC_DOCKER_NAME_CMD="${FPC_TOP_DIR}/utils/fabric/get-fabric-container-name"

export LD_LIBRARY_PATH=${LD_LIBRARY_PATH:+"${LD_LIBRARY_PATH}:"}${GOPATH}/src/github.com/hyperledger-labs/fabric-private-chaincode/tlcc/enclave/lib 


get_fpc_name() {
    echo "ecc_${1}"
}


# Chaincode command wrappers
#----------------------------

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
		CC_ENCLAVESOPATH=$2
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
	if [ -z ${CC_NAME+x} ] || [ -z ${CC_VERSION+x} ] || [ -z ${CC_ENCLAVESOPATH+x} ] ; then
	    # missing params, don't do anything and let real peer report errors
	    return
	fi
	yell "Found valid FPC Chaincode Install!"
	# get net-id and peer-id from core.yaml, for now just use defaults ...
	parse_fabric_config ${FABRIC_CFG_PATH}

	# add private chaincode name prefix
	FPC_NAME=$(get_fpc_name ${CC_NAME})

	# get docker name for this chaincode
	DOCKER_IMAGE_NAME=$(${FPC_DOCKER_NAME_CMD} --cc-name ${FPC_NAME} --cc-version ${CC_VERSION} --net-id ${NET_ID} --peer-id ${PEER_ID}) || die "could not get docker image name"

	# install docker
	try make ENCLAVE_SO_PATH=${CC_ENCLAVESOPATH} DOCKER_IMAGE=${DOCKER_IMAGE_NAME} -C ${FPC_TOP_DIR}/ecc docker-fpc-app

	# eplace path and lang arg with dummy go chaincode
	ARGS_EXEC=( 'chaincode' 'install' '-n' "${FPC_NAME}" '-v' "${CC_VERSION}" '-p' 'github.com/hyperledger/fabric/examples/chaincode/go/example02/cmd' "${OTHER_ARGS[@]}" )
    fi
    return
}

handle_chaincode_upgrade() {
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
                CC_ENCLAVESOPATH=$2
                shift; shift
                ;;
            -l|--lang)                                                                                                                            CC_LANG=$2
                shift;shift
                ;;
            *)
                OTHER_ARGS+=( "$1" )
                shift
                ;;
            esac
    done
    # accept only if original install is of same "type" (fpc vs vanilla)
    $(is_fpc_cc ${CC_NAME} ${CC_VERSION})
    FPC_INSTALLED=$?
    [ "${CC_LANG}" = "fpc-c" ];
    FPC_UPGRADE=$?
    if [ ${FPC_INSTALLED} != ${FPC_UPGRADE} ]; then
	die "cannot change frpm FPC to non-FPC or vice-versa in chaincode upgrade"
    fi

    # if fpc, do same switcheroo and docker-build as done by install
    if [ "${CC_LANG}" = "fpc-c" ]; then                                                                                                   if [ -z ${CC_NAME+x} ] || [ -z ${CC_VERSION+x} ] || [ -z ${CC_ENCLAVESOPATH+x} ] ; then
            # missing params, don't do anything and let real peer report errors
            return
        fi
        yell "Found valid FPC Chaincode Install!"
        # get net-id and peer-id from core.yaml, for now just use defaults ...
        parse_fabric_config .

        # add private chaincode name prefix
        FPC_NAME=$(get_fpc_name ${CC_NAME})

        # get docker name for this chaincode
        DOCKER_IMAGE_NAME=$(${FPC_DOCKER_NAME_CMD} --cc-name ${FPC_NAME} --cc-version ${CC_VERSION} --net-id ${NET_ID} --peer-id ${PEER_ID}) || die "could not get docker image name"

        # install docker
        try make ENCLAVE_SO_PATH=${CC_ENCLAVESOPATH} DOCKER_IMAGE=${DOCKER_IMAGE_NAME} -C ${FPC_TOP_DIR}/ecc docker-fpc-app

        # replace path and lang arg with dummy go chaincode
        ARGS_EXEC=( 'chaincode' 'upgrade' '-n' "${FPC_NAME}" '-v' "${CC_VERSION}" '-p' 'github.com/hyperledger/fabric/examples/cha
incode/go/example02/cmd' "${OTHER_ARGS[@]}" )
    fi
    return
}


handle_chaincode_instantiate() {
    handle_chaincode_call_with_mapped_name "instantiate" "$@"
    # Note: we rely on above to analyze args and assign (& translate) CC_NAME for -n and -c init-args CC_MSG and IS_FPC_CC

    # - instantiate chaincode
    try $RUN ${FABRIC_BIN_DIR}/peer "${ARGS_EXEC[@]}"
    # Note: the -c init-args which ultimately have to go to the enclave are also passed here: ecc just ignores them, so no point to filter them. Besides we need it if CC is _not_ FPC ..
    sleep 3 # unfortunately, no --waitForEvent equivalent, so do heuristic sleep ...

    ondemand_setup

    # - exit
    exit 0
}

handle_chaincode_invoke() {
    handle_chaincode_call_with_mapped_name "invoke" "$@"
    ondemand_setup
    return
}

handle_chaincode_query() {
    handle_chaincode_call_with_mapped_name "query" "$@"
    ondemand_setup
    return
}

handle_chaincode_list() {
    OUT=$(${FABRIC_BIN_DIR}/peer "${ARGS_EXEC[@]}")
    rc=$?
    if [ $rc = 0 ]; then
	echo "${OUT}" | sed 's/Name: ecc_/Name: /g';
    fi
    exit $rc
}


# chaincode cmd utility functions

ondemand_setup() {
    if [ ${IS_FPC_CC} -eq 1 ]; then
	if [ -z ${CC_VERSION} ]; then
	    # an invoke or query: we don't get version on command-line and have to first query for it ...
	    CC_VERSION=$(${FABRIC_BIN_DIR}/peer chaincode list -C ${CHAN_ID} --instantiated | awk "/Name: ${CC_NAME}/ "'{ print $4 }' | sed 's/,$//')
	fi

	setup_file="${FABRIC_STATE_DIR}/${CHAN_ID}.${CC_NAME}.${CC_VERSION}.setup"

	# - do not do anything iff we are already setup for this chaincode on this peer
	if [ -e "${setup_file}" ]; then
	    return
	fi

        # - setup internal ecc state, e.g., register chaincode
        try $RUN ${FABRIC_BIN_DIR}/peer chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_NAME} -c '{"Args":["__setup", "'${ERCC_ID}'"]}' --waitForEvent

	# - remember we setup this chaincode on this peer ..
	try touch ${setup_file}

        # - retrieve public-key (just for fun of it ...)
        try $RUN ${FABRIC_BIN_DIR}/peer chaincode query -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_NAME} -c '{"Args":["__getEnclavePk"]}'
    fi
}

handle_chaincode_call_with_mapped_name() {
    CMD=$1; shift

    OTHER_ARGS=()
    while [[ $# > 0 ]]; do
	case "$1" in
	    -n|--name)
		CC_NAME=$2;
		shift; shift
		;;
	    -v|--version)
		# Note: this exists only for instantiate, not invoke or query;
		# will handle the case of undefined CC_VERSION inside
		# translate_fpc_name, though  ..
		CC_VERSION=$2
		OTHER_ARGS+=( "$1" "$2" )
		shift; shift
		;;
	    -c|--ctor)
		CC_MSG=$2;
		# Note: we need this only for special case of instantiate
		# by doing it here we have to parse args only once ...
		OTHER_ARGS+=( "$1" "$2" )
		shift; shift
		;;
	    *)
		OTHER_ARGS+=( "$1" )
		shift
		;;
	    esac
    done

    OLD_CC_NAME=${CC_NAME}
    CC_NAME=$(translate_fpc_name ${CC_NAME}  ${CC_VERSION})
    if [ "${OLD_CC_NAME}" == "${CC_NAME}" ]; then
	IS_FPC_CC=0
    else
	IS_FPC_CC=1
    fi
    ARGS_EXEC=( 'chaincode' "${CMD}" '-n' "${CC_NAME}" "${OTHER_ARGS[@]}" )
}

translate_fpc_name() {
    CC_NAME=$1
    CC_VERSION=$2

    if $(is_fpc_cc ${CC_NAME} ${CC_VERSION}); then
        get_fpc_name ${CC_NAME}
    else
        echo "${CC_NAME}"
    fi
}

is_fpc_cc() {
    CC_NAME=$1
    CC_VERSION=$2
    FPC_NAME=$(get_fpc_name ${CC_NAME})

    # if we don't have a version in our invocation ...
    if [ -z ${CC_VERSION} ]; then
	# it must be an instantiated version, so let's
	# check whether a FPC-named one exists
	FPC_VERSION=$(${FABRIC_BIN_DIR}/peer chaincode list -C ${CHAN_ID} --instantiated | awk "/Name: ${FPC_NAME}/ "'{ print $4 }' | sed 's/,$//')
	[ ! -z "${FPC_VERSION}" ];
	return $?
    else
	# otherwise, there should be at least a corresponding docker image
	DOCKER_IMAGE_NAME=$(${FPC_DOCKER_NAME_CMD} --cc-name ${FPC_NAME} --cc-version ${CC_VERSION} --net-id ${NET_ID} --peer-id ${PEER_ID}) || die "could not get docker image name"

	[ ! -z "$(docker images | grep "${DOCKER_IMAGE_NAME}")" ];
	return $?
    fi
}

# Channel command wrappers
#--------------------------

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
    try touch "${FABRIC_STATE_DIR}/${CHAN_ID}.creator"
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
    if [ -e "${FABRIC_STATE_DIR}/${CHAN_ID}.creator" ]; then
	try $RUN ${FABRIC_BIN_DIR}/peer chaincode instantiate -n ${ERCC_ID} -v ${ERCC_VERSION} -c '{"args":["init"]}' -C ${CHAN_ID} -V ercc-vscc
	sleep 4
	try rm "${FABRIC_STATE_DIR}/${CHAN_ID}.creator"
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
	    upgrade)
		shift
		handle_chaincode_upgrade "$@"
		;;
	    instantiate)
		shift
		handle_chaincode_instantiate "$@"
		;;
	    invoke)
		shift
		handle_chaincode_invoke "$@"
		;;
	    query)
		shift
		handle_chaincode_query "$@"
		;;
	    list)
		shift
		handle_chaincode_list "$@"
		;;
	    *)
		# fall through, nothing to do
		# TODO: is this really safe? should we do anything with 'package' and/or 'signpackage'?
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
