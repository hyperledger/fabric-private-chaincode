# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

# *** README ***
# This script is meant to run as part of the build.
# The script is transferred to the folder where other test binaries will be located,
# and it will orchestrate the test.
# Orchestration involves: preparing input file for init_attestation,
# calling get_attestation, calling attestation_to_evidence, calling verify_evidence.

set -e

. ../conversion/attestation_to_evidence.sh
. ../conversion/define_to_variable.sh
. ../conversion/enclave_to_mrenclave.sh

DEFINES_FILEPATH="${FPC_PATH}/common/crypto/attestation-api/test/common/test-defines.h"
TAGS_FILEPATH="${FPC_PATH}/common/crypto/attestation-api/attestation/attestation_tags.h"

function remove_artifacts()
{
    rm -rf *.txt
}

function orchestrate()
{
    #get attestation
    ./get_attestation_app
    define_to_variable "${DEFINES_FILEPATH}" "GET_ATTESTATION_OUTPUT"
    [ -f ${GET_ATTESTATION_OUTPUT} ] || die "no output from get_attestation"

    #translate attestation (note: attestation_to_evidence defines the EVIDENCE variable)
    ATTESTATION=$(cat ${GET_ATTESTATION_OUTPUT})
    attestation_to_evidence "${ATTESTATION}"

    define_to_variable "${DEFINES_FILEPATH}" "EVIDENCE_FILE"
    echo ${EVIDENCE} > ${EVIDENCE_FILE}

    #verify evidence
    ./verify_evidence_app
}

function orchestrate_with_go_conversion()
{
    #get attestation
    ./get_attestation_app
    define_to_variable "${DEFINES_FILEPATH}" "GET_ATTESTATION_OUTPUT"
    [ -f ${GET_ATTESTATION_OUTPUT} ] || die "no output from get_attestation"

    #translate attestation (note: attestation_to_evidence defines the EVIDENCE variable)
    ATTESTATION=$(cat ${GET_ATTESTATION_OUTPUT})
    GO_CONVERSION_CMD="go run ${FPC_PATH}/common/crypto/attestation-api/test/conversion_app_go/main.go"
    EVIDENCE=$(${GO_CONVERSION_CMD} "${ATTESTATION}")

    define_to_variable "${DEFINES_FILEPATH}" "EVIDENCE_FILE"
    echo ${EVIDENCE} > ${EVIDENCE_FILE}

    #verify evidence
    ./verify_evidence_app
}

function orchestrate_with_go_verification()
{
    #get attestation
    ./get_attestation_app
    define_to_variable "${DEFINES_FILEPATH}" "GET_ATTESTATION_OUTPUT"
    [ -f ${GET_ATTESTATION_OUTPUT} ] || die "no output from get_attestation"

    #translate attestation (note: attestation_to_evidence defines the EVIDENCE variable)
    ATTESTATION=$(cat ${GET_ATTESTATION_OUTPUT})
    attestation_to_evidence "${ATTESTATION}"

    define_to_variable "${DEFINES_FILEPATH}" "EVIDENCE_FILE"
    echo ${EVIDENCE} > ${EVIDENCE_FILE}

    #verify evidence
    go run -tags WITH_PDO_CRYPTO ${FPC_PATH}/common/crypto/attestation-api/test/verify_evidence_app_go/main.go
}

#######################################
# sim mode test
#######################################
if [[ ${SGX_MODE} == "SIM" ]]; then
    say "Testing simulated attestation"

    #prepare input
    remove_artifacts
    define_to_variable "${DEFINES_FILEPATH}" "CODE_ID_FILE"
    define_to_variable "${DEFINES_FILEPATH}" "STATEMENT_FILE"
    define_to_variable "${DEFINES_FILEPATH}" "STATEMENT"
    define_to_variable "${DEFINES_FILEPATH}" "INIT_DATA_INPUT"

    define_to_variable "${TAGS_FILEPATH}" "ATTESTATION_TYPE_TAG"
    define_to_variable "${TAGS_FILEPATH}" "SIMULATED_TYPE_TAG"

    echo -n "this is ignored" > ${CODE_ID_FILE}
    echo -n "also ignored" > ${STATEMENT_FILE}
    echo -n "{\"${ATTESTATION_TYPE_TAG}\": \"${SIMULATED_TYPE_TAG}\"}" > ${INIT_DATA_INPUT}

    #run attestation generation/conversion/verification tests
    orchestrate

    #run attestation generation/conversion/verification tests (same as before, though with Go-based conversion)
    orchestrate_with_go_conversion

    #run attestation generation/conversion/verification tests (same as before, though with Go-based verification)
    orchestrate_with_go_verification

    say "Test simulated attestation success"
else
    say "Skipping actual attestation test"
fi

#######################################
# hw mode test
#######################################
if [[ ${SGX_MODE} == "HW" ]]; then
    say "Testing HW-mode attestation"

    #prepare input
    remove_artifacts
    define_to_variable "${DEFINES_FILEPATH}" "CODE_ID_FILE"
    define_to_variable "${DEFINES_FILEPATH}" "STATEMENT_FILE"
    define_to_variable "${DEFINES_FILEPATH}" "STATEMENT"
    define_to_variable "${DEFINES_FILEPATH}" "INIT_DATA_INPUT"

    define_to_variable "${DEFINES_FILEPATH}" "UNSIGNED_ENCLAVE_FILENAME"
    enclave_to_mrenclave ${UNSIGNED_ENCLAVE_FILENAME} test_enclave.config.xml
    echo -n "$MRENCLAVE" > ${CODE_ID_FILE}
    echo -n ${STATEMENT} > ${STATEMENT_FILE}

    #get spid type
    SPID_TYPE_FILEPATH="${FPC_PATH}/config/ias/spid_type.txt"
    test -f ${SPID_TYPE_FILEPATH} || die "no spid type file ${SPID_TYPE_FILEPATH}"
    SPID_TYPE=$(cat $SPID_TYPE_FILEPATH)

    #get spid
    SPID_FILEPATH="${FPC_PATH}/config/ias/spid.txt"
    test -f ${SPID_FILEPATH} || die "no spid file ${SPID_FILEPATH}"
    SPID=$(cat $SPID_FILEPATH)

    define_to_variable "${TAGS_FILEPATH}" "SPID_TAG"
    define_to_variable "${TAGS_FILEPATH}" "SIG_RL_TAG"
    echo -n "{\"${ATTESTATION_TYPE_TAG}\": \"$SPID_TYPE\", \"${SPID_TAG}\": \"$SPID\", \"${SIG_RL_TAG}\":\"\"}" > ${INIT_DATA_INPUT}

    #run attestation generation/conversion/verification tests
    orchestrate

    #run attestation generation/conversion/verification tests (same as before, though with Go-based conversion)
    orchestrate_with_go_conversion
else
    say "Skipping actual attestation test"
fi

say "Test successful."
exit 0
