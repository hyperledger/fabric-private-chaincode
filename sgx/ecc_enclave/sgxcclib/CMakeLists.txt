# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

INCLUDE(CMakeVariables.txt)

set(SOURCE_FILES
    sgxcclib.c
    enclave_u.c
    ${COMMON_SOURCE_DIR}/sgxcclib/common-sgxcclib.c
    )

set(cleanup_files
    ${CMAKE_CURRENT_SOURCE_DIR}/enclave_u.c
    ${CMAKE_CURRENT_SOURCE_DIR}/enclave_u.h
    ${CMAKE_LIBRARY_OUTPUT_DIRECTORY}/include
    ${CMAKE_LIBRARY_OUTPUT_DIRECTORY}/lib)
set_directory_properties(PROPERTIES ADDITIONAL_MAKE_CLEAN_FILES "${cleanup_files}")

add_custom_command(
    OUTPUT ${CMAKE_CURRENT_SOURCE_DIR}/enclave_u.c
    COMMAND ${SGX_EDGER8R} --untrusted enclave.edl
    --search-path ../enclave/ # enclave.edl
    --search-path ${COMMON_SOURCE_DIR}/enclave
    --search-path ${COMMON_SOURCE_DIR}/crypto
    --search-path ${COMMON_SOURCE_DIR}/logging/ocalls
    --search-path ${SGX_SDK}/include
    COMMENT "Generating enclave_u.{c,h}"
    WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
    )

add_library(libsgxcc SHARED ${SOURCE_FILES})

include_directories(
    ${CMAKE_CURRENT_SOURCE_DIR}
    ${COMMON_SOURCE_DIR}
    ${COMMON_SOURCE_DIR}/base64
    ${COMMON_SOURCE_DIR}/sgxcclib
    ${COMMON_SOURCE_DIR}/logging/untrusted
    ${SGX_SDK}/include
    )

set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} ${SGX_COMMON_CFLAGS} -fPIC -Wall -Wno-attributes -std=c11 ${SGX_ATTESTATION_FLAG}")

set_target_properties(libsgxcc PROPERTIES PREFIX "")

target_link_libraries(libsgxcc
    -L${SGX_LIBRARY_PATH}
    -l${SGX_URTS_LIB}
    -l${SGX_LNCH_LIB}
    -l${SGX_EPID_LIB}
    -L${SGX_SSL_LIBRARY_PATH}
    -lsgx_usgxssl
    # adding crypto adapt because enclave_u uses ocalls which are built into crypto adapt
    -L${COMMON_SOURCE_DIR}/crypto/_build -l${U_CRYPTO_ADAPT_LIB_NAME}
    -L${COMMON_SOURCE_DIR}/logging/_build -l${U_LOGGING_LIB_NAME}
    )

COND_ENABLE_COVERAGE(libsgxcc)

add_custom_command(TARGET libsgxcc
    POST_BUILD
    COMMAND mkdir -p ${CMAKE_BINARY_DIR}/include
    COMMAND cp ${CMAKE_CURRENT_SOURCE_DIR}/sgxcclib.h ${CMAKE_BINARY_DIR}/include
    COMMAND cp ${COMMON_SOURCE_DIR}/fpc-types.h ${CMAKE_BINARY_DIR}/include
    COMMAND mkdir -p ${CMAKE_BINARY_DIR}/lib/
    COMMAND cp ${CMAKE_BINARY_DIR}/libsgxcc.so ${CMAKE_BINARY_DIR}/lib/
    WORKING_DIRECTORY ${CMAKE_BINARY_DIR}
    COMMENT "Add include to build")

