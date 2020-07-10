#!/bin/bash
# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

# NOTES:
# - multi-peer support: In the earlier v1.4 version, this peer wrapper should
#   have worked with multiple peers, with the channel-creator being the one how
#   instantiates ercc and no additional sync needed beyond peers as already
#   necessary in vanilla Fabric (one peer creates channel, then everybody
#   joins). This pattern does not work anymore with the new lifecycle where ercc
#   instantation requires tighter synchronization due to the new approval process.
#   Hence, as of now, this peer wrapper does _not_ support multi-peer anymore!
# - multi-channel support: Currently FPC supports only a single channel. This
#   script doesn't prevent you, though, configuring ercc on multiple-channels,
#   so make sure externally than 'channel join' is called only for a single channel.

#RUN=echo # uncomment to dry-run peer call

SCRIPTDIR="$(dirname $(readlink --canonicalize ${BASH_SOURCE}))"
FPC_TOP_DIR="${SCRIPTDIR}/../../"
FABRIC_SCRIPTDIR="${FPC_TOP_DIR}/fabric/bin/"

METADATA_FILE="metadata.json"

: ${FABRIC_CFG_PATH:=$(pwd)}
: ${SGX_MODE:=SIM}

. ${FABRIC_SCRIPTDIR}/lib/common_utils.sh
. ${FABRIC_SCRIPTDIR}/lib/common_ledger.sh

FPC_DOCKER_NAME_CMD="${FPC_TOP_DIR}/utils/fabric/get-fabric-container-name"

export LD_LIBRARY_PATH=${LD_LIBRARY_PATH:+"${LD_LIBRARY_PATH}:"}${FPC_TOP_DIR}/tlcc/enclave/lib


# Lifecycle Chaincode command wrappers
#--------------------------------------------

handle_lifecycle_ercc_package() {
    # check required parameters
    [ ! -z "${ERCC_PACKAGE}" ]      || die "undefined ercc package"
    [ ! -z "${ERCC_TYPE}" ]         || die "undefined ercc type"
    [ ! -z "${ERCC_LABEL}" ]        || die "undefined ercc label'"
    [ -d "${ERCC_PATH}" ]           || die "undefined or non-existing ercc path"
    # Note: normal fabric package format & layout:
    # Overall the package is a gzipped tar-file containing files
    # - '${METADATA_FILE}', a json object with 'path', 'type' and 'label' string fields
    # - 'code.tar.gz' a gzipped tar-fil containing files
    #    - 'src/...'

    FPC_PKG_SANDBOX="$(mktemp -d -t  fpc-pkg-sandbox.XXX)" || die "mktemp failed"

    # - create code.tar.gz
    try cd "${ERCC_PATH}"
    [ -f "ercc" ]   || die "no binary file 'ercc' in '${ERCC_PATH}'"
    try tar -zcf "${FPC_PKG_SANDBOX}/code.tar.gz" "ercc"

    # - create ${METADATA_FILE}
    cat <<EOF >"${FPC_PKG_SANDBOX}/${METADATA_FILE}"
{
            "path":"${ERCC_PATH}",
            "type":"${ERCC_TYPE}",
            "label":"${ERCC_LABEL}"
}
EOF
    cat ${FPC_PKG_SANDBOX}/${METADATA_FILE}

    # - tar it
    try cd "${FPC_PKG_SANDBOX}"
    try tar -zcf "${ERCC_PACKAGE}" *

    # - cleanup
    try rm -rf "${FPC_PKG_SANDBOX}"
}

handle_lifecycle_chaincode_package() {
    OTHER_ARGS=()
    while [[ $# > 0 ]]; do
	case "$1" in
	    --label)
		CC_LABEL="$2"
		shift; shift
		;;
	    -p|--path)
		CC_ENCLAVESOPATH="$2"
		shift; shift
		;;
	    -l|--lang)
		CC_LANG="$2"
		shift;shift
		;;
	    -s|--sgx-mode)
		# Note: this is a new parameter not existing in the 'vanilla' peer.
		# If the SGX_MODE environment variable exists, it will also be used
		# (unless overriden by this flag)
		SGX_MODE="$2"
		shift;shift
		;;
	    # Above is the flags we really care, but we need also the outputfile
	    # which doesn't have a flag. So let's enumerate the known no-arg
	    # flags (i.e., --tls -h/--help), assume all other flags have exactly
	    # one arg (true as of v2.1.1) and then the remaining one is the
	    # output file ...
	    -h|--help)
		# with help, no point to continue but run it right here ..
		try $RUN ${FABRIC_BIN_DIR}/peer "${ARGS_EXEC[@]}"
		# .. as well as augment it with additiona -s/--sgx-mode arg
		echo ""
		echo "Flags for fpc-c chaincode:"
		echo "  -s, --sgx-mode string                SGX-mode to run with"
		exit 0 
		;;

	    --clientauth)
		OTHER_ARGS+=( "$1" )
		shift
		;;

	    --tls)
		OTHER_ARGS+=( "$1" )
		shift
		;;

	    -*)
		OTHER_ARGS+=( "$1" "$2" )
		shift; shift
		;;
	    *)
		# Note: we require it later to be an absolute path!!
		OUTPUTFILE=$(readlink -f "$1")
		shift
		;;
	    esac
    done

    if [ ! "${CC_LANG}" = "fpc-c" ]; then
	    # Nothing special to do for non-fpc chaincode, just forward to real peer
	    return
    fi

    # check required parameters
    [ ! -z "${OUTPUTFILE}" ]     || die "no or ill-defined outputfile provided"
    [ ! -z "${CC_LABEL}" ]       || die "undefined parameter '--label'"
    [ -d "${CC_ENCLAVESOPATH}" ] || die "undefined or non-existing '-p'/'--path' parameter '${CC_ENCLAVESOPATH}'"

    # Note: normal fabric package format & layout:
    # Overall the package is a gzipped tar-file containing files
    # - '${METADATA_FILE}', a json object with 'path', 'type' and 'label' string fields
    # - 'code.tar.gz' a gzipped tar-fil containing files
    #    - 'src/...'
    # as for fpc for now we will package already built artifacts 'enclave.signed.so' and
    # 'mrenclave', we will skip 'src' and directly place the built artifacts into
    # the root of 'code.tar.gz'. (Eventually we might add reproducible build to the
    # external builder, in which case we would stuff the related source into 'src/...'
    # for ${METADATA_FILE} use the params passed to us, i.e., in particular type 'fpc-c'.

    FPC_PKG_SANDBOX="$(mktemp -d -t  fpc-pkg-sandbox.XXX)" || die "mktemp failed"

    # - create code.tar.gz
    ENCLAVE_FILE="enclave.signed.so"
    MRENCLAVE_FILE="mrenclave"
    try cd "${CC_ENCLAVESOPATH}"
    [ -f "${ENCLAVE_FILE}" ]   || die "no enclave file '${ENCLAVE_FILE}' in '${CC_ENCLAVESOPATH}'"
    [ -f "${MRENCLAVE_FILE}" ] || die "no enclave file '${MRENCLAVE_FILE}' in '${CC_ENCLAVESOPATH}'"
    try tar -zcf "${FPC_PKG_SANDBOX}/code.tar.gz" \
	"${ENCLAVE_FILE}" \
	"${MRENCLAVE_FILE}"

    # - create ${METADATA_FILE}
    [ ! -z "${SGX_MODE}" ] || die "SGX_MODE not correctly specified either via environment variable or -s/--sgx-mode argument"
    cat <<EOF >"${FPC_PKG_SANDBOX}/${METADATA_FILE}"
{
  "path":"${CC_ENCLAVESOPATH}",
  "type":"${CC_LANG}",
  "label":"${CC_LABEL}",
  "sgx_mode":"${SGX_MODE}"
}
EOF
    # note:
    # - in addition to standard fields we also add the SGX_MODE to be used
    # - for golang path is a normalized go package. In our case we do need
    #   path but just pass it along as it might be useful in debugging

    # - tar it
    try cd "${FPC_PKG_SANDBOX}"
    try tar -zcf "${OUTPUTFILE}" *
    # Note: the
    # - for bizare reason, fabric peer refuses to accept the tar if you tar
    #   as . which also creates a ./ directory entry?!!
    # - file is absolute, so the various cd's do not hurt ..

    # - cleanup
    try rm -rf "${FPC_PKG_SANDBOX}"

    exit 0
}


handle_lifecycle_chaincode_install() {
    # to allow non-fpc CC, we will have to keep track of installed pkg-ids
    # corresponding to fpc-c chaincode

    # parse args to find package name
    while [[ $# > 0 ]]; do
	case "$1" in
	    # we care only about package file name but this is not prefixed
	    # with a flag.  So let's enumerate the known no-arg flags (i.e.,
	    # --tls -h/--help), assume all other flags have exactly
	    # one arg (true as of v2.1.1) and then the remaining one is the
	    # output file ...
	    -h|--help)
		# with help, no point to continue but run it right here ..
		try $RUN ${FABRIC_BIN_DIR}/peer "${ARGS_EXEC[@]}"
		exit 0 
		;;

	    --clientauth)
		OTHER_ARGS+=( "$1" )
		shift
		;;

	    --tls)
		OTHER_ARGS+=( "$1" )
		shift
		;;

	    -*)
		OTHER_ARGS+=( "$1" "$2" )
		shift; shift
		;;
	    *)
		PKG_FILE="$1"
		shift
		;;
	    esac
    done

    # - do normal install (and exit if not successfull)
    try $RUN ${FABRIC_BIN_DIR}/peer "${ARGS_EXEC[@]}"

    # - inspect ${METADATA_FILE} from package tar & get type
    CC_LANG=$(tar -zvxf "${PKG_FILE}" --to-stdout ${METADATA_FILE} | jq .type | tr -d '"' | tr '[:upper:]' '[:lower:]') || die "could not extract cc language type from package file '${PKG_FILE}'"

    # - iff type is fpc-c
    if [ "${CC_LANG}" = "fpc-c" ]; then
	#   - get label from ${METADATA_FILE}
	CC_LABEL=$(tar -zvxf "${PKG_FILE}" --to-stdout ${METADATA_FILE} | jq .label | tr -d '"') || die "could not extract label from package file '${PKG_FILE}'"
	#   - extract package id PKG_ID via queryinstalled
	PKG_ID=$(${FABRIC_BIN_DIR}/peer lifecycle chaincode queryinstalled | awk "/Package ID: ${CC_LABEL}:/{print}" | sed -n 's/^Package ID: //; s/, Label:.*$//;p')
	[ ! -z "${PKG_ID}" ] || die "Could not extract package id"
	#   - remember this 
	try touch "${FABRIC_STATE_DIR}/is-fpc-c-package.${PKG_ID}"
    fi

    # - exit (instead of below return, which would re-execute install)
    exit 0
}

handle_lifecycle_chaincode_approveformyorg() {
    # to allow non-fpc CC, we will have to keep track here of pkg to name.version
    # mapping for fpc-c-code

    # - extract package-id PKG_ID, name CC_ID and version CC_VERSION from args
    while [[ $# > 0 ]]; do
	case "$1" in
	    --package-id)
		PKG_ID=$2;
		shift; shift
		;;
	    -n|--name)
		CC_NAME=$2;
		shift; shift
		;;
	    -v|--version)
		CC_VERSION=$2;
		shift; shift
		;;
	    -C|--channelID)
		CHAN_ID=$2;
		shift; shift
		;;
        -E|--endorsement-plugin)
        ENDORSEMENT_PLUGIN_NAME=$2;
        shift; shift
        ;;
        -V|--validation-plugin)
        VALIDATION_PLUGIN_NAME=$2;
        shift; shift
        ;;
	    *)
		shift
		;;
	    esac
    done

    # - iff it is fpc pkg
    if [ -f "${FABRIC_STATE_DIR}/is-fpc-c-package.${PKG_ID}" ]; then
        # no endorsement plugin can be specified in FPC
        [ -z "${ENDORSEMENT_PLUGIN_NAME}" ] || die "Endorsement plugins are disabled for FPC chaincodes"
        # no validation plugin can be specified in FPC
        [ -z "${VALIDATION_PLUGIN_NAME}" ] || die "Validation plugins are disabled for FPC chaincodes"

        # all check passed

        # add FPC default validation plugin to argument list
        ARGS_EXEC=( "${ARGS_EXEC[@]}" "-V" "fpc-vscc" )
    fi

    # - do normal approve (and exit if not successfull)
    try $RUN ${FABRIC_BIN_DIR}/peer "${ARGS_EXEC[@]}"

    # - iff it is fpc pkg
    if [ -f "${FABRIC_STATE_DIR}/is-fpc-c-package.${PKG_ID}" ]; then
	    #  remember mapping
	    try touch "${FABRIC_STATE_DIR}/is-fpc-c-chaincode.${CC_NAME}.${CC_VERSION}"
    fi
    # - exit (instead of below return, which would re-execute install)
    exit 0
}

handle_lifecycle_chaincode_checkcommitreadiness() {
    # NOTE: this command is intercepted only to hide FPC validation plugin in peer CLI

    # - remember variables we might need later
    while [[ $# > 0 ]]; do
    case "$1" in
        -n|--name)
        CC_NAME=$2;
        shift; shift
        ;;
        -v|--version)
        CC_VERSION=$2;
        shift; shift
        ;;
        -C|--channelID)
        CHAN_ID=$2;
        shift; shift
        ;;
        -E|--endorsement-plugin)
        ENDORSEMENT_PLUGIN_NAME=$2;
        shift; shift
        ;;
        -V|--validation-plugin)
        VALIDATION_PLUGIN_NAME=$2;
        shift; shift
        ;;
        *)
        shift
        ;;
        esac
    done

    # - iff it is fpc pkg
    if [ -f "${FABRIC_STATE_DIR}/is-fpc-c-chaincode.${CC_NAME}.${CC_VERSION}" ]; then
        # no endorsement plugin can be specified in FPC
        [ -z "${ENDORSEMENT_PLUGIN_NAME}" ] || die "Endorsement plugins are disabled for FPC chaincodes"
        # no validation plugin can be specified in FPC
        [ -z "${VALIDATION_PLUGIN_NAME}" ] || die "Validation plugins are disabled for FPC chaincodes"

        # all check passed

        # add FPC default validation plugin to argument list
        ARGS_EXEC=( "${ARGS_EXEC[@]}" "-V" "fpc-vscc" )
    fi

    # - call real peer so channel is joined
    try $RUN ${FABRIC_BIN_DIR}/peer "${ARGS_EXEC[@]}"

    exit 0
}

handle_lifecycle_chaincode_commit() {
    # - remember variables we might need later
    while [[ $# > 0 ]]; do
	case "$1" in
	    -n|--name)
		CC_NAME=$2;
		shift; shift
		;;
	    -v|--version)
		CC_VERSION=$2;
		shift; shift
		;;
	    -C|--channelID)
		CHAN_ID=$2;
		shift; shift
		;;
        -E|--endorsement-plugin)
        ENDORSEMENT_PLUGIN_NAME=$2;
        shift; shift
        ;;
        -V|--validation-plugin)
        VALIDATION_PLUGIN_NAME=$2;
        shift; shift
        ;;
	    *)
		shift
		;;
	    esac
    done

    # - iff it is fpc pkg
    if [ -f "${FABRIC_STATE_DIR}/is-fpc-c-chaincode.${CC_NAME}.${CC_VERSION}" ]; then
        # no endorsement plugin can be specified in FPC
        [ -z "${ENDORSEMENT_PLUGIN_NAME}" ] || die "Endorsement plugins are disabled for FPC chaincodes"
        # no validation plugin can be specified in FPC
        [ -z "${VALIDATION_PLUGIN_NAME}" ] || die "Validation plugins are disabled for FPC chaincodes"

        # all check passed

        # add FPC default validation plugin to argument list
        ARGS_EXEC=( "${ARGS_EXEC[@]}" "-V" "fpc-vscc" )
    fi

    # - call real peer so chaincode is committed
    try $RUN ${FABRIC_BIN_DIR}/peer "${ARGS_EXEC[@]}"

    exit 0
}

handle_lifecycle_chaincode_createenclave() {
    # - remember variables we might need later
    while [[ $# > 0 ]]; do
    case "$1" in
        -n|--name)
        CC_NAME=$2;
        shift; shift
        ;;
        *)
        die "createenclave: invalid option"
        ;;
        esac
    done

    # - createenclave can only be run on FPC chaincodes
    FILES_NUM=$(ls -1 ${FABRIC_STATE_DIR}/is-fpc-c-chaincode.${CC_NAME}.* 2> /dev/null | wc -l) 
    if [ ${FILES_NUM} -eq 0 ]; then
        die "createenclave: $CC_NAME is not written in language 'fpc-c'"
    fi

    # - setup internal ecc state, e.g., register chaincode
    try ${FABRIC_BIN_DIR}/peer chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_NAME} -c '{"Args":["__setup", "'${ERCC_ID}'"]}' --waitForEvent

    # - retrieve public-key (just for fun of it ...)
    try $RUN ${FABRIC_BIN_DIR}/peer chaincode query -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_NAME} -c '{"Args":["__getEnclavePk"]}'

    # - exit (otherwise main function will invoke operation again!)
    exit 0
}

# Chaincode command wrappers
#--------------------------------------------

handle_chaincode_invoke() {
    # TODO: eventually add client-side encryption/decryption
    return
}

handle_chaincode_query() {
    # TODO: eventually add client-side encryption/decryption
    return
}


# Channel command wrappers
#--------------------------

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
    ERCC_LABEL="${ERCC_ID}_${ERCC_VERSION}"
    ERCC_PACKAGE=${FABRIC_STATE_DIR}/ercc.tar.gz
    ERCC_QUERY_INSTALL_LOG=${FABRIC_STATE_DIR}/ercc-query-install.$$.log
    ERCC_PATH="${FPC_TOP_DIR}/ercc"
    ERCC_TYPE="ercc-type"
    say "Installing ercc on channel '${CHAN_ID}' ..."
    say "Packaging ${ERCC_ID} ..."
    handle_lifecycle_ercc_package
    para
    sleep 3
    say "Installing ${ERCC_ID} ..."
    try $RUN ${FABRIC_BIN_DIR}/peer lifecycle chaincode install ${ERCC_PACKAGE}
    para
    sleep 3
    say "Querying installed chaincodes to find package id.."
    try $RUN ${FABRIC_BIN_DIR}/peer lifecycle chaincode queryinstalled >& ${ERCC_QUERY_INSTALL_LOG}
    para
    ERCC_PACKAGE_ID=$(awk "/Package ID: ${ERCC_LABEL}/{print}" ${ERCC_QUERY_INSTALL_LOG} | sed -n 's/^Package ID: //; s/, Label:.*$//;p')
    [ ! -z "${ERCC_PACKAGE_ID}" ] || die "Could not extract package id"
    say "Approve for my org"
    try $RUN ${FABRIC_BIN_DIR}/peer lifecycle chaincode approveformyorg -o ${ORDERER_ADDR} --channelID ${CHAN_ID} --name ${ERCC_ID} --version ${ERCC_VERSION} --package-id ${ERCC_PACKAGE_ID} --sequence ${ERCC_SEQUENCE} -V ercc-vscc
    para
    sleep 3
    say "Checking for commit readiness"
    try $RUN ${FABRIC_BIN_DIR}/peer lifecycle chaincode checkcommitreadiness --channelID ${CHAN_ID} --name ${ERCC_ID} --version ${ERCC_VERSION} --sequence ${ERCC_SEQUENCE} --output json -V ercc-vscc
    para
    sleep 3
    say "Committing chaincode definition...."
    try $RUN ${FABRIC_BIN_DIR}/peer lifecycle chaincode commit -o ${ORDERER_ADDR} --channelID ${CHAN_ID} --name ${ERCC_ID} --version ${ERCC_VERSION} --sequence ${ERCC_SEQUENCE} -V ercc-vscc
    para
    sleep 3
    # Note: Below is not crucial but they do display potentially useful info and the second also is liveness-test for ercc
    say "Query commited chaincodes on the channel"
    try $RUN ${FABRIC_BIN_DIR}/peer lifecycle chaincode querycommitted --channelID ${CHAN_ID}
    para
    sleep 3
    say "call chaincode query ..."
    try $RUN ${FABRIC_BIN_DIR}/peer chaincode query -n ${ERCC_ID} -c '{"function":"getSPID","args":[]}' -C ${CHAN_ID}
    sleep 3


    # - ask tlcc to join channel
    #   IMPORTANT: right now a join is _not_ persistant, so on restart of peer,
    #   it will re-join old channels but tlcc will not!
    say "Attaching TLCC to channel '${CHAN_ID}' ..."
    try $RUN ${FABRIC_BIN_DIR}/peer chaincode query -n tlcc -c '{"Args": ["JOIN_CHANNEL"]}' -C ${CHAN_ID}

    # - exit (otherwise main function will invoke operation again!)
    exit 0
}


# - check whether it is a command which we have to intercept
#   - channel join
#   - lifecycle chaincode package
#   - lifecycle chaincode commit
#   - chaincode invoke
#   - chaincode query
ARGS_EXEC=( "$@" ) # params to eventually pass to real peer /default: just pass all original args ..
case "$1" in
    lifecycle)
	shift
	case "$1" in
	    chaincode)
		shift
		case "$1" in
		    package)
			shift
			handle_lifecycle_chaincode_package "$@"
			;;
		    install)
			shift
			handle_lifecycle_chaincode_install "$@"
			;;
		    approveformyorg)
			shift
			handle_lifecycle_chaincode_approveformyorg "$@"
			;;
            checkcommitreadiness)
            shift
            handle_lifecycle_chaincode_checkcommitreadiness "$@"
            ;;
		    commit)
			shift
			handle_lifecycle_chaincode_commit "$@"
			;;
		    createenclave)
			shift
			handle_lifecycle_chaincode_createenclave "$@"
			;;
		    *)
			# fall through, nothing to do
			;;
		esac
		;;

	    *)
		# fall through, nothing to do
		;;
	esac
	;;

    chaincode)
	shift
	case "$1" in
	    invoke)
		shift
		handle_chaincode_invoke "$@"
		;;
	    query)
		shift
		handle_chaincode_query "$@"
		;;
	    *)
		# fall through, nothing to do
		# Note: old lifecycle commands (e.g.,install, instantiate, upgrade, list)
		# are not supported anymore in v2 channel! So no need to wrap
	esac
	;;

    channel)
	shift
	case "$1" in
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
