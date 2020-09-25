#!/bin/bash

# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

set -e

if [[ -z "${FPC_PATH}" ]]; then
    echo "FPC_PATH not set"
    exit -1
fi

CUR_SCRIPT_PATH="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. ${FPC_PATH}/fabric/bin/lib/common_utils.sh
. ${CUR_SCRIPT_PATH}/define_to_variable.sh

###########################################################
# b64quote_to_iasresponse
#   input:  quote as parameter
#   output: IAS_RESPONSE variable
###########################################################
function b64quote_to_iasresponse() {
    #get api key
    API_KEY_FILEPATH="${FPC_PATH}/config/ias/api_key.txt"
    test -f ${API_KEY_FILEPATH} || die "no api key file ${API_KEY_FILEPATH}"
    API_KEY=$(cat $API_KEY_FILEPATH)

    #get verification report
    QUOTE=$1
    # contact IAS to get the verification report
    IAS_RESPONSE=$(curl -s -H "Content-Type: application/json" -H "Ocp-Apim-Subscription-Key:$API_KEY" -X POST -d '{"isvEnclaveQuote":"'$QUOTE'"}' https://api.trustedservices.intel.com/sgx/dev/attestation/v4/report -i)
    # check status (as there may be multiple header, we rather check presence of relevant fields)
    echo "$IAS_RESPONSE" | grep "X-IASReport-Signature"           >/dev/null || die "IAS Response error"
    echo "$IAS_RESPONSE" | grep "X-IASReport-Signing-Certificate" >/dev/null || die "IAS Response error"
}

###########################################################
# iasresponse_to_evidence
#   input:  ias response as parameter
#   output: IAS_EVIDENCE variable
###########################################################
function iasresponse_to_evidence() {
    IAS_RESPONSE="$1"
    #encode relevant info in json format
    IAS_SIGNATURE=$(echo "$IAS_RESPONSE" | grep -Po 'X-IASReport-Signature: \K[^ ]+' | tr -d '\r')
    IAS_CERTIFICATES=$(echo "$IAS_RESPONSE" | grep -Po 'X-IASReport-Signing-Certificate: \K[^ ]+')
    IAS_REPORT=$(echo "$IAS_RESPONSE" | grep -Po '{"id":[^ ]+')
    JSON_IAS_RESPONSE=$(jq -c -n --arg sig "$IAS_SIGNATURE" --arg cer "$IAS_CERTIFICATES" --arg rep "$IAS_REPORT" '{iasSignature: $sig, iasCertificates: $cer, iasReport: $rep}')
    #set output
    IAS_EVIDENCE=$JSON_IAS_RESPONSE
}

###########################################################
# simulated_to_evidence
#   input:  evidence as parameter
#   output: SIMULATED_EVIDENCE variable
###########################################################
function simulated_to_evidence() {
    SIMULATED_EVIDENCE=$1
}

###########################################################
# get_tag_make_variable
#   input:  tag string (e.g, "TAG_X") as parameter
#   output: tag string variable (e.g., TAG_X)
###########################################################
function get_tag_make_variable() {
    TAGS_PATH="${FPC_PATH}/common/crypto/attestation-api/attestation/attestation_tags.h"
    define_to_variable "$TAGS_PATH" "$1"
}

get_tag_make_variable "ATTESTATION_TYPE_TAG"
get_tag_make_variable "ATTESTATION_TAG"
get_tag_make_variable "EVIDENCE_TAG"
get_tag_make_variable "SIMULATED_TYPE_TAG"
get_tag_make_variable "EPID_LINKABLE_TYPE_TAG"
get_tag_make_variable "EPID_UNLINKABLE_TYPE_TAG"

###########################################################
# attestation_to_evidence
#   input:  attestation as parameter
#   output: EVIDENCE variable
###########################################################
function attestation_to_evidence() {
    if [[ -z "$1" ]]; then
        die "no argument provided"
    fi

    say "Input Attestation: $1"

    ATTESTATION_TYPE=$(echo $1 | jq ".$ATTESTATION_TYPE_TAG" -r)
    ATTESTATION=$(echo $1 | jq ".$ATTESTATION_TAG" -r)

    case "$ATTESTATION_TYPE" in
        $SIMULATED_TYPE_TAG)
            simulated_to_evidence "$ATTESTATION"
            EVIDENCE=$SIMULATED_EVIDENCE
            ;;

        $EPID_LINKABLE_TYPE_TAG)
            ;&
        $EPID_UNLINKABLE_TYPE_TAG)
            b64quote_to_iasresponse "$ATTESTATION"
            iasresponse_to_evidence "$IAS_RESPONSE"
            EVIDENCE=$IAS_EVIDENCE
            ;;
        *)
            die "error attestation type $ATTESTATION_TYPE"
            ;;
    esac

    #package evidence
    EVIDENCE=$(jq -c -n --arg attestation_type "$ATTESTATION_TYPE" --arg evidence "$EVIDENCE" '{'$ATTESTATION_TYPE_TAG': $attestation_type, '$EVIDENCE_TAG': $evidence}')

    say "Output Evidence: $EVIDENCE"
}
