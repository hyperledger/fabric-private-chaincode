# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

CMAKE_MINIMUM_REQUIRED(VERSION 3.2 FATAL_ERROR)

INCLUDE(../../cmake/ConfigSGX.cmake)

SET(U_CRYPTO_ADAPT_LIB_NAME updo-crypto-adapt)
SET(T_CRYPTO_ADAPT_LIB_NAME tpdo-crypto-adapt)

SET(PDO_CRYPTO_DIR "pdo/common/crypto")
SET(PDO_CRYPTO_TEST_DIR "${PDO_CRYPTO_DIR}/../tests/crypto")

FILE(GLOB PROJECT_HEADERS "${PDO_CRYPTO_DIR}/../*.h")
FILE(GLOB PROJECT_SOURCES
    "${PDO_CRYPTO_DIR}/../c11_support.cpp"
    "${PDO_CRYPTO_DIR}/../hex_string.cpp"
    "${PDO_CRYPTO_DIR}/../log.cpp"
    "pdo-types.cpp"
    "$ENV{FPC_PATH}/common/base64/base64.cpp"
    )

###################################################################################################
# Untrusted crypto adapt library
###################################################################################################
PROJECT(${U_CRYPTO_ADAPT_LIB_NAME} CXX)
ADD_LIBRARY(${U_CRYPTO_ADAPT_LIB_NAME} STATIC ${PROJECT_HEADERS} ${PROJECT_SOURCES})
TARGET_INCLUDE_DIRECTORIES(${U_CRYPTO_ADAPT_LIB_NAME} PRIVATE "${PDO_CRYPTO_DIR}/..")
TARGET_INCLUDE_DIRECTORIES(${U_CRYPTO_ADAPT_LIB_NAME} PRIVATE "$ENV{FPC_PATH}/common/base64")
TARGET_INCLUDE_DIRECTORIES(${U_CRYPTO_ADAPT_LIB_NAME} PUBLIC $ENV{SGX_SDK}/include)

###################################################################################################
# Trusted crypto adapt library
###################################################################################################
PROJECT(${T_CRYPTO_ADAPT_LIB_NAME} CXX)
ADD_LIBRARY(${T_CRYPTO_ADAPT_LIB_NAME} STATIC ${PROJECT_HEADERS} ${PROJECT_SOURCES})
TARGET_INCLUDE_DIRECTORIES(${T_CRYPTO_ADAPT_LIB_NAME} PRIVATE "${PDO_CRYPTO_DIR}/..")
TARGET_INCLUDE_DIRECTORIES(${T_CRYPTO_ADAPT_LIB_NAME} PRIVATE "$ENV{FPC_PATH}/common/base64")
TARGET_INCLUDE_DIRECTORIES(${T_CRYPTO_ADAPT_LIB_NAME} PUBLIC $ENV{SGX_SDK}/include)
TARGET_COMPILE_OPTIONS(${T_CRYPTO_ADAPT_LIB_NAME} PRIVATE -nostdinc++)
TARGET_INCLUDE_DIRECTORIES(${T_CRYPTO_ADAPT_LIB_NAME} PUBLIC ${SGX_SDK}/include/libcxx)
TARGET_INCLUDE_DIRECTORIES(${T_CRYPTO_ADAPT_LIB_NAME} PUBLIC ${SGX_SDK}/include/tlibc)

###################################################################################################
# PDO Crypto library
###################################################################################################
set(CMAKE_ARCHIVE_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR})
ADD_SUBDIRECTORY (${PDO_CRYPTO_DIR})

###################################################################################################
# PDO Crypto Tests
###################################################################################################
# cmakevariable needed to load function definition
INCLUDE(${PDO_CRYPTO_DIR}/../CMakeVariables.txt)
# pdo top dir needed for includes
SET(PDO_TOP_DIR "${CMAKE_CURRENT_SOURCE_DIR}/pdo")
# lib names needed at link time
SET(TRUSTED_LIB_NAME ${T_CRYPTO_ADAPT_LIB_NAME})
SET(UNTRUSTED_LIB_NAME ${U_CRYPTO_ADAPT_LIB_NAME})
SET(U_CRYPTO_LIB_NAME updo-crypto)
SET(T_CRYPTO_LIB_NAME tpdo-crypto)
# pkg_check_modules needed to set OPENSSL_LDFLAGS
find_package(PkgConfig REQUIRED)
pkg_check_modules (OPENSSL REQUIRED openssl>=1.1.0g)
# add pdo crypto tests
ADD_SUBDIRECTORY (${PDO_CRYPTO_TEST_DIR})
