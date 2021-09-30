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

#RUN=echo   # uncomment (or define when calling script) to dry-run peer call
#DEBUG=true # uncomment (or define when calling script) to show debug output

SCRIPTDIR="$(dirname $(readlink --canonicalize ${BASH_SOURCE}))"
FPC_PATH="${SCRIPTDIR}/../../"
FABRIC_SCRIPTDIR="${FPC_PATH}/fabric/bin/"

METADATA_FILE="metadata.json"

: ${FABRIC_CFG_PATH:=$(pwd)}
: ${SGX_MODE:=SIM}

. ${FABRIC_SCRIPTDIR}/lib/common_utils.sh
. ${FABRIC_SCRIPTDIR}/lib/common_ledger.sh

export LD_LIBRARY_PATH=${LD_LIBRARY_PATH:+"${LD_LIBRARY_PATH}:"}${FPC_PATH}/tlcc/enclave/lib


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
	    # one arg (true as of v2.3.3) and then the remaining one is the
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
	    # one arg (true as of v2.3.3) and then the remaining one is the
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
		CC_ID=$2;
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
    fi

    # - do normal approve (and exit if not successfull)
    try $RUN ${FABRIC_BIN_DIR}/peer "${ARGS_EXEC[@]}"

    # - iff it is fpc pkg
    if [ -f "${FABRIC_STATE_DIR}/is-fpc-c-package.${PKG_ID}" ]; then
	    #  remember mapping
	    try touch "${FABRIC_STATE_DIR}/is-fpc-c-chaincode.${CC_ID}.${CC_VERSION}"
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
        CC_ID=$2;
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
    if [ -f "${FABRIC_STATE_DIR}/is-fpc-c-chaincode.${CC_ID}.${CC_VERSION}" ]; then
        # no endorsement plugin can be specified in FPC
        [ -z "${ENDORSEMENT_PLUGIN_NAME}" ] || die "Endorsement plugins are disabled for FPC chaincodes"
        # no validation plugin can be specified in FPC
        [ -z "${VALIDATION_PLUGIN_NAME}" ] || die "Validation plugins are disabled for FPC chaincodes"

        # all check passed
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
		CC_ID=$2;
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
    if [ -f "${FABRIC_STATE_DIR}/is-fpc-c-chaincode.${CC_ID}.${CC_VERSION}" ]; then
        # no endorsement plugin can be specified in FPC
        [ -z "${ENDORSEMENT_PLUGIN_NAME}" ] || die "Endorsement plugins are disabled for FPC chaincodes"
        # no validation plugin can be specified in FPC
        [ -z "${VALIDATION_PLUGIN_NAME}" ] || die "Validation plugins are disabled for FPC chaincodes"

        # all check passed
    fi

    # - call real peer so chaincode is committed
    try $RUN ${FABRIC_BIN_DIR}/peer "${ARGS_EXEC[@]}"

    exit 0
}

handle_lifecycle_chaincode_initEnclave() {
    # - remember variables we might need later
    while [[ $# > 0 ]]; do
    case "$1" in
	-h|--help)
	    cat <<EOF
Initialize enclave for fpc chaincode.

Usage:
	peer chaincode initEnclave [flags]

Flags:
  -n, --name string                    Name of the chaincode
      --peerAddresses stringArray      The addresses of the peers to host to-be-initialized enclave
  -s|--sgx-mode string                 SGX mode (HW or SIM).
                                       (Can also be defined by 'SGX_MODE' environment variable)
  -S|--sgx-credentials-path string     Path for sgx credentials.
                                       (Can also be defined by 'SGX_CREDENTIALS_PATH' environment variable)
  -C, --channelID string               The channel on which this command should be executed
  -o|--orderer string 		       Ordering service endpoint
  -h, --help                           help for invoke

EOF
	    exit 0
	    ;;
        -n|--name)
            CC_ID=$2;
            shift; shift
            ;;
        -C|--channelID)
            CHAN_ID=$2;
            shift; shift
            ;;
        -o|--orderer)
            ORDERER_ADDR=$2;
            shift; shift
            ;;
        -s|--sgx-mode)
            SGX_MODE=$2
            shift; shift
            ;;
        -S|--sgx-credentials-path)
            SGX_CREDENTIALS_PATH=$2
            shift; shift
            ;;
        --peerAddresses)
            PEER_ADDRESS=$2
            shift; shift
            ;;
        *)
            die "initEnclave: invalid option"
            ;;
        esac
    done

    if [ "${SGX_MODE}" = "SIM" ] ; then
        # set the default attestation params
        ATTESTATION_PARAMS=$(jq -c -n --arg atype "simulated" '{attestation_type: $atype}' | base64 --wrap=0)
    elif [ "${SGX_MODE}" = "HW" ] ; then
        SPID_FILE_PATH="${SGX_CREDENTIALS_PATH}/spid.txt"
        SPID_TYPE_FILE_PATH="${SGX_CREDENTIALS_PATH}/spid_type.txt"
        [ -f "${SPID_FILE_PATH}" ] || die "no spid file ${SPID_FILE_PATH}"
        [ -f "${SPID_TYPE_FILE_PATH}" ] || die "no spid type file ${SPID_TYPE_FILE_PATH}"
        # set hw-mode attestation params
        # it is assumed that sig_rl is empty
        ATTESTATION_PARAMS=$(jq -c -n --arg atype "$(cat ${SPID_TYPE_FILE_PATH})" --arg spid "$(cat ${SPID_FILE_PATH})" --arg sig_rl "" '{attestation_type: $atype, hex_spid: $spid, sig_rl: $sig_rl}' | base64 --wrap=0)
    else
	die "illegal sgx mode '${SGX_MODE}', should be either 'SIM' or 'HW'"
    fi

    # peer address must be specified in initEnclave
    [ -z "${PEER_ADDRESS}" ] && die "No peer address specified in initEnclave"
    # and there must be only one
    [ $(echo "${PEER_ADDRESS}" | awk -F"," '{print NF-1}') == 0 ] || die "one and only one peer address allowed"
    [ $(echo "${PEER_ADDRESS}" | awk -F":" '{print NF-1}') == 1 ] || die "one and only one port allowed"

    # - initEnclave can only be run on FPC chaincodes
    FILES_NUM=$(ls -1 ${FABRIC_STATE_DIR}/is-fpc-c-chaincode.${CC_ID}.* 2> /dev/null | wc -l)
    if [ ${FILES_NUM} -eq 0 ]; then
        die "initEnclave: $CC_ID is not written in language 'fpc-c'"
    fi

    # create host params
    PEER_ENDPOINT="${PEER_ADDRESS}"

    # create init enclave message
    INIT_ENCLAVE_PROTO=$( (echo "peer_endpoint: \"${PEER_ENDPOINT}\""; echo "attestation_params: \"${ATTESTATION_PARAMS}\"") | protoc --encode fpc.InitEnclaveMessage --proto_path=${FPC_PATH}/protos/fpc --proto_path=${FPC_PATH}/protos/fabric ${FPC_PATH}/protos/fpc/fpc.proto | base64 --wrap=0)
    [ -z ${INIT_ENCLAVE_PROTO} ] && die "init enclave proto is empty"

    # trigger initEnclave
    try_out_r $RUN ${FABRIC_BIN_DIR}/peer chaincode query -o ${ORDERER_ADDR} --peerAddresses "${PEER_ADDRESS}" -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["__initEnclave", "'${INIT_ENCLAVE_PROTO}'"]}'
    CC_CREDS_B64=${RESPONSE}
    [ -z ${CC_CREDS_B64} ] && die "initEnclave failed"
    [ -z ${DEBUG+x} ] || say "initEnclave response (b64): ${CC_CREDS_B64}"
    [ -z ${DEBUG+x} ] || say "initEnclave response (decoded): $(echo "${CC_CREDS_B64}" | base64 -d)"

    echo "Convert credentials"
    CC_CREDS_CONV_B64=$(echo "${CC_CREDS_B64}" | ${PEER_ASSIST_CMD} attestation2Evidence) || die "could not convert credentials"
    [ -z ${DEBUG+x} ] && say "initEnclave converted response (b64): ${CC_CREDS_CONV_B64}"

    echo "Registering with Enclave Registry"
    try $RUN ${FABRIC_BIN_DIR}/peer chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${ERCC_ID} -c '{"Args":["RegisterEnclave", "'${CC_CREDS_CONV_B64}'"]}' --waitForEvent

    # NOTE: the chaincode encryption key is retrieved here for testing purposes
    echo "Querying Chaincode Encryption Key"
    try_out_r $RUN ${FABRIC_BIN_DIR}/peer chaincode query -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${ERCC_ID} -c '{"Args":["QueryChaincodeEncryptionKey", "'${CC_ID}'"]}'
    CC_EK_B64="${RESPONSE}"
    CC_EK=$(echo ${CC_EK_B64} | base64 -d)
    echo "Chaincode EK (b64): ${CC_EK_B64}"
    [ -z ${DEBUG+x} ] || say "Chaincode EK: ${CC_EK}"

    # - exit (otherwise main function will invoke operation again!)
    exit 0
}

# Chaincode command wrappers
#--------------------------------------------

handle_chaincode_call() {
    CMD=$1; shift

    OTHER_ARGS=()
    DO_HEX=0
    DO_RAW=0
    DO_WAIT=0
    while [[ $# > 0 ]]; do
        case "$1" in
	    -h|--help)
		return
		;;
	    -C|--channelID)
		CHAN_ID=$2;
		shift; shift
		;;
            --peerAddresses)
		PEER_ADDRESS=$2
		EXPLICIT_PEER=1
		shift; shift
		;;
            -n|--name)
                CC_ID=$2;
                shift; shift
                ;;
            -c|--ctor)
                CC_MSG=$2;
                shift; shift
                ;;
	    -I|--isInit)
                IS_INIT=$2;
                shift; shift
                ;;
	    --transient)
                TRANSIENT=$2;
                shift; shift
                ;;
	    -x|--hex)
                DO_HEX=1
                shift
                ;;
	    -r|--raw)
                DO_RAW=1
                shift
                ;;
	    --waitForEvent)
                DO_WAIT=1
                shift
		;;
	    --waitForEventTimeout)
                DO_WAIT_TIMEOUT=$1
                shift
		;;
            *)
		# these args we pass to both enclave query and validation invoke
		# e.g., --clientauth, --(tls|ca|cert|key)*, --(conn*)*, --order*
                OTHER_ARGS+=( "$1" )
                shift
                ;;
            esac
    done

    # - iff it is not a fpc pkg
    if [ ! -f "${FABRIC_STATE_DIR}/is-fpc-c-chaincode.${CC_ID}"* ]; then
	[ -z ${DEBUG+x} ] || say "non-FPC chaincode"
	return
    else
	[ -z ${DEBUG+x} ] || say "FPC chaincode"
    fi

    # - check for unsupported options $IS_INIT, $TRANSIENT
    [ ${DO_HEX} ] || die "--hex/-x is not supported for for FPC chaincodes"
    [ ${DO_RAW} ] || die "--raw/-r is not supported for for FPC chaincodes"
    [ -z "${IS_INIT}" ] || die "--isInit/-I is not supported for for FPC chaincodes"
    [ -z "${TRANSIENT}" ] || die "--transient is not supported for for FPC chaincodes"

    # - query ercc.queryChaincodeEncryptionKey for chaincode encryption key
    try_out_r $RUN ${FABRIC_BIN_DIR}/peer chaincode query -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${ERCC_ID} -c '{"Args":["QueryChaincodeEncryptionKey", "'${CC_ID}'"]}'
    CC_EK_B64="${RESPONSE}"
    [ -z ${CC_EK_B64} ] && die "looking up chaincode encryption key failed"
    [ -z ${DEBUG+x} ] || say "Chaincode EK (b64): ${CC_EK_B64}"

    # - if no endpoint specified, query ercc.queryChaincodeEndPoints for endpoint
    if [ !${EXPLICIT_PEER} ]; then
	try_out_r $RUN ${FABRIC_BIN_DIR}/peer chaincode query -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${ERCC_ID} -c '{"Args":["QueryChaincodeEndPoints", "'${CC_ID}'"]}'
	PEER_ADDRESS="${RESPONSE}"
	[ -z ${PEER_ADDRESS} ] && die "looking up peer endpoint failed"
	[ -z ${DEBUG+x} ] || say "peer endpoint: ${PEER_ADDRESS}"
    fi

    # - start cli assist with input to fd 3 and output from fd 4, a fifo
    [ -z ${DEBUG+x} ] || say "starting peer assist .."
    result_pipe=$(mktemp -u -t peer_cli_assistXXXX)
    mkfifo $result_pipe || die "can't create fifo '${result_pipe}'"
    exec 3> >(${PEER_ASSIST_CMD} handleRequestAndResponse "${CC_EK_B64}" "${result_pipe}")
    assist_pid=$!
    exec 4<${result_pipe}

    # - send clear-text request to assist
    [ -z ${DEBUG+x} ] || say "sending cleartext request '${CC_MSG}' to peer assist .."
    echo >&3 "${CC_MSG}"

    # - receive encrypted request from assist
    read <&4 encrypted_request
    [ -z ${encrypted_request} ] && die "peer assist failed to produce encrypted request"
    [ -z ${DEBUG+x} ] || say "received encrypted request '${encrypted_request}' from peer assist"

    # - invoke function in enclave (as fabric query)
    try_out_r $RUN ${FABRIC_BIN_DIR}/peer chaincode query --peerAddresses ${PEER_ADDRESS} "${OTHER_ARGS[@]}" -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["__invoke", "'${encrypted_request}'"]}'
    encrypted_response="${RESPONSE}"
    [ -z ${encrypted_response} ] && die "calling enclave failed"

    # - send encrypted response to assist
    [ -z ${DEBUG+x} ] || say "sending encrypted response '${encrypted_response}' to peer assist .."
    echo >&3 "${encrypted_response}"

    # - receive decrypted response from assist
    read <&4 decrypted_response
    [ -z ${decrypted_response} ] && die "peer assist failed to produce decrypted response"
    [ -z ${DEBUG+x} ] || say "received decrypted response '${decrypted_response}' from peer assist"

    # - get assist return code
    wait ${assist_pid}
    rc=$?
    [ ${rc} == 0 ] || die "assist could not properly transform requests/responses (rc=${rc})"

    # iff invoke (rather than query), cause endorsement flows
    if [ ${CMD} == "invoke" ]; then
	# - invoke validation of enclave endorsement
	opt_arg=""
	if [ ${EXPLICIT_PEER} ]; then
	    opt_arg="--peerAddresses ${PEER_ADDRESS} ${opt_arg}"
	fi
	if [ ${DO_WAIT} == 1 ]; then
	    opt_arg="--waitForEvent ${opt_arg}"
	fi
	if [ ! -z ${DO_WAIT_TIMEOUT+x} ]; then
	    opt_arg="--waitForEvent ${opt_arg}"
	fi
	try_r $RUN ${FABRIC_BIN_DIR}/peer chaincode invoke "${OTHER_ARGS[@]}" ${opt_arg} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["__endorse", "'${encrypted_response}'"]}'
	endorse_response=$(parse_invoke_result_from_log "${RESPONSE}")
        [ "${endorse_response}" == "OK" ] || die "endorsement failed: '${endorse_response}'"
        [ -z ${DEBUG+x} ] || say "endorsement returned '${endorse_response}'"
    fi
    # - return decrypted response
    #   TODO: post-processing once we return proper fabric response objects
    #      As part of this, might also consider supporting -r/--raw and/or -x/--hex
    #      (Note that strangely these options exist only for query, not invoke?!)
    echo ${decrypted_response}

    exit 0
}

handle_chaincode_invoke() {
    handle_chaincode_call "invoke" "$@"
    return
}

handle_chaincode_query() {
    handle_chaincode_call "query" "$@"
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
    ERCC_PATH="${FPC_PATH}/ercc"
    ERCC_TYPE="ercc-type"
    echo "Installing ercc on channel '${CHAN_ID}' ..."
    echo "Packaging ${ERCC_ID} ..."
    handle_lifecycle_ercc_package
    para
    sleep 3
    echo "Installing ${ERCC_ID} ..."
    try $RUN ${FABRIC_BIN_DIR}/peer lifecycle chaincode install ${ERCC_PACKAGE}
    para
    sleep 3
    echo "Querying installed chaincodes to find package id.."
    try $RUN ${FABRIC_BIN_DIR}/peer lifecycle chaincode queryinstalled >& ${ERCC_QUERY_INSTALL_LOG}
    para
    ERCC_PACKAGE_ID=$(awk "/Package ID: ${ERCC_LABEL}/{print}" ${ERCC_QUERY_INSTALL_LOG} | sed -n 's/^Package ID: //; s/, Label:.*$//;p')
    [ ! -z "${ERCC_PACKAGE_ID}" ] || die "Could not extract package id"
    echo "Approve for my org"
    try $RUN ${FABRIC_BIN_DIR}/peer lifecycle chaincode approveformyorg -o ${ORDERER_ADDR} --channelID ${CHAN_ID} --name ${ERCC_ID} --version ${ERCC_VERSION} --package-id ${ERCC_PACKAGE_ID} --sequence ${ERCC_SEQUENCE}
    para
    sleep 3
    echo "Checking for commit readiness"
    try $RUN ${FABRIC_BIN_DIR}/peer lifecycle chaincode checkcommitreadiness --channelID ${CHAN_ID} --name ${ERCC_ID} --version ${ERCC_VERSION} --sequence ${ERCC_SEQUENCE} --output json
    para
    sleep 3
    echo "Committing chaincode definition...."
    try $RUN ${FABRIC_BIN_DIR}/peer lifecycle chaincode commit -o ${ORDERER_ADDR} --channelID ${CHAN_ID} --name ${ERCC_ID} --version ${ERCC_VERSION} --sequence ${ERCC_SEQUENCE}
    para
    sleep 3
    # Note: Below is not crucial but they do display potentially useful info and the second also is liveness-test for ercc
    echo "Query commited chaincodes on the channel"
    try $RUN ${FABRIC_BIN_DIR}/peer lifecycle chaincode querycommitted --channelID ${CHAN_ID}
    para
    sleep 3

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
		    initEnclave)
			shift
			handle_lifecycle_chaincode_initEnclave "$@"
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
