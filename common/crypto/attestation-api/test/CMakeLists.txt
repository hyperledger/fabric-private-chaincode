# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

INCLUDE ("$ENV{FPC_PATH}/cmake/ConfigSGX.cmake")
INCLUDE ("$ENV{FPC_PATH}/cmake/Debug.cmake")
INCLUDE ("$ENV{FPC_PATH}/common/crypto/pdo/common/CMakeVariables.txt")

# Put test artifacts under /tests subdirectory
set(CMAKE_LIBRARY_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/tests)
set(CMAKE_RUNTIME_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/tests)

# add logging library directory for linking
LINK_DIRECTORIES("${LOGGING_PATH}/_build")

##################################################################################################
## Verify Evidence test app:
## this application tests evidence from input files
##################################################################################################

SET(VERIFY_EVIDENCE_APP verify_evidence_app)

ADD_EXECUTABLE(${VERIFY_EVIDENCE_APP}
    ${VERIFY_EVIDENCE_APP}/main.cpp
    ${VERIFY_EVIDENCE_APP}/test.cpp
    common/test-utils.cpp
    )

TARGET_INCLUDE_DIRECTORIES(${VERIFY_EVIDENCE_APP} PRIVATE "$ENV{FPC_PATH}/common")
TARGET_INCLUDE_DIRECTORIES(${VERIFY_EVIDENCE_APP} PRIVATE "$ENV{FPC_PATH}/common/base64")
TARGET_INCLUDE_DIRECTORIES(${VERIFY_EVIDENCE_APP} PRIVATE "$ENV{FPC_PATH}/common/crypto/attestation-api/attestation")
TARGET_INCLUDE_DIRECTORIES(${VERIFY_EVIDENCE_APP} PRIVATE "$ENV{FPC_PATH}/common/crypto/attestation-api/evidence")
TARGET_INCLUDE_DIRECTORIES(${VERIFY_EVIDENCE_APP} PRIVATE "$ENV{FPC_PATH}/common/crypto/attestation-api/test/common")
TARGET_INCLUDE_DIRECTORIES(${VERIFY_EVIDENCE_APP} PRIVATE "${LOGGING_PATH}/untrusted")
TARGET_LINK_LIBRARIES(${VERIFY_EVIDENCE_APP} ${U_LOGGING_LIB_NAME})
# Link the untrusted test application against the untrusted library and openssl
TARGET_LINK_LIBRARIES(${VERIFY_EVIDENCE_APP}
    "-L$ENV{FPC_PATH}/common/crypto/_build"
    "-Wl,--start-group" ${OPENSSL_LDFLAGS} "-l${U_CRYPTO_ADAPT_LIB_NAME}" "-l${U_CRYPTO_LIB_NAME}" "-Wl,--end-group"
    )

COND_ENABLE_DEBUG(${VERIFY_EVIDENCE_APP})

COND_ENABLE_COVERAGE(${VERIFY_EVIDENCE_APP})

###################################################################################################
## Attestation app:
## this application performs a sim/hw attestation from input files and writes output on file
###################################################################################################

SET(GET_ATTESTATION_APP "get_attestation_app")

ADD_SUBDIRECTORY(${GET_ATTESTATION_APP}/enclave)

# ENCLAVE_EDL is defined in the enclave subdirectory
SGX_EDGE_UNTRUSTED(${ENCLAVE_EDL} ENCLAVE_EDGE_SOURCES)

ADD_EXECUTABLE(${GET_ATTESTATION_APP}
    ${GET_ATTESTATION_APP}/app/main.cpp
    common/test-utils.cpp
    ${ENCLAVE_EDGE_SOURCES})

# Make sure the enclave builds before the test app that links it
ADD_DEPENDENCIES(${GET_ATTESTATION_APP} test_enclave)

TARGET_LINK_LIBRARIES(${GET_ATTESTATION_APP} -Wl,-L,${SGX_SDK}/lib64)
TARGET_LINK_LIBRARIES(${GET_ATTESTATION_APP} -Wl,-L,${SGX_SSL}/lib64)
TARGET_LINK_LIBRARIES(${GET_ATTESTATION_APP} ${U_CRYPTO_ADAPT_LIB_NAME})
TARGET_LINK_LIBRARIES(${GET_ATTESTATION_APP} ${URTS_LIBRARY_NAME} sgx_usgxssl ${SGX_EPID_LIB})

TARGET_INCLUDE_DIRECTORIES(${GET_ATTESTATION_APP} PRIVATE "$ENV{FPC_PATH}/common")
TARGET_INCLUDE_DIRECTORIES(${GET_ATTESTATION_APP} PRIVATE "${LOGGING_PATH}/untrusted")
TARGET_INCLUDE_DIRECTORIES(${GET_ATTESTATION_APP} PRIVATE "$ENV{FPC_PATH}/common/crypto/attestation-api/test/common")

TARGET_LINK_LIBRARIES(${GET_ATTESTATION_APP} "${U_LOGGING_LIB_NAME}")

COND_ENABLE_DEBUG(${GET_ATTESTATION_APP})

COND_ENABLE_COVERAGE(${GET_ATTESTATION_APP})

SET(ATTESTED_EVIDENCE_TEST_FILE "attested_evidence_test.sh")
ADD_CUSTOM_COMMAND(
    TARGET ${GET_ATTESTATION_APP}
    PRE_BUILD
    COMMAND cp ${ATTESTED_EVIDENCE_TEST_FILE} ${CMAKE_RUNTIME_OUTPUT_DIRECTORY}
    COMMAND cp ${GET_ATTESTATION_APP}/enclave/test_enclave.config.xml ${CMAKE_RUNTIME_OUTPUT_DIRECTORY}
    COMMAND mkdir -p ${CMAKE_RUNTIME_OUTPUT_DIRECTORY}/../conversion && cp -r ../conversion/* ${CMAKE_RUNTIME_OUTPUT_DIRECTORY}/../conversion
    WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
    )

# Register this application as a test
add_test(
    NAME ${ATTESTED_EVIDENCE_TEST_FILE}
    COMMAND bash ./${ATTESTED_EVIDENCE_TEST_FILE}
    WORKING_DIRECTORY ${CMAKE_RUNTIME_OUTPUT_DIRECTORY}
    )

#################################################################################
## Run the tests automatically after building
#################################################################################
IF(DEFINED TEST_TARGET)
    ADD_DEPENDENCIES(${TEST_TARGET} ${VERIFY_EVIDENCE_APP})
    ADD_DEPENDENCIES(${TEST_TARGET} ${GET_ATTESTATION_APP})
    ADD_DEPENDENCIES(${TEST_TARGET} ${ATTESTED_EVIDENCE_TEST_FILE})
endif()
