# Copyright IBM Corp. All Rights Reserved.
# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

include(CMakeVariables.txt)

set(SOURCE_FILES
    crypto.cpp
    enclave.cpp
    enclave_t.c
    shim.cpp
    ${COMMON_SOURCE_DIR}/enclave/common.cpp
    ${COMMON_SOURCE_DIR}/base64/base64.cpp
    ${COMMON_SOURCE_DIR}/utils.c
    ${COMMON_SOURCE_DIR}/json/parson.c
    ${COMMON_SOURCE_DIR}/logging.c
    )

add_definitions(-DENCLAVE_CODE)

set(cleanup_files
    ${CMAKE_CURRENT_SOURCE_DIR}/enclave_t.c
    ${CMAKE_CURRENT_SOURCE_DIR}/enclave_t.h
    )
set_directory_properties(PROPERTIES ADDITIONAL_MAKE_CLEAN_FILES "${cleanup_files}")

add_custom_command(
    OUTPUT ${CMAKE_CURRENT_SOURCE_DIR}/enclave_t.c
    COMMAND ${SGX_EDGER8R} --trusted enclave.edl
    --search-path ${COMMON_SOURCE_DIR}/enclave
    --search-path ${SGX_SDK}/include
    --search-path ${SGX_SSL}/include
    COMMENT "Generating enclave_t.{c,h}"
    WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
    )

add_library(${SHIM_LIB_NAME} STATIC ${SOURCE_FILES})

include_directories(
    ${CMAKE_CURRENT_SOURCE_DIR}
    ${COMMON_SOURCE_DIR}
    ${COMMON_SOURCE_DIR}/base64
    ${COMMON_SOURCE_DIR}/enclave
    ${COMMON_SOURCE_DIR}/json
    ${SGX_SDK}/include
    ${SGX_SDK}/include/tlibc
    ${SGX_SDK}/include/libcxx
    ${SGX_SSL}/include
    ${FPC_PATH}/common/crypto
    ${FPC_PATH}/common/crypto/pdo/common
    )

set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} ${SGX_COMMON_CFLAGS} -nostdinc -fno-builtin -fvisibility=hidden -fpie -fstack-protector -std=c11")
set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} ${SGX_COMMON_CFLAGS} -nostdinc -fno-builtin -fvisibility=hidden -fpie -fstack-protector -std=c++11")

add_custom_command(TARGET ${SHIM_LIB_NAME}
    POST_BUILD
    COMMAND mkdir -p ${CMAKE_BINARY_DIR}/lib
    COMMAND cp lib${SHIM_LIB_NAME}.a ${CMAKE_BINARY_DIR}/lib/
    )
