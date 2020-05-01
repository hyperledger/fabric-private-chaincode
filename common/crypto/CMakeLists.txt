# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

CMAKE_MINIMUM_REQUIRED(VERSION 3.2 FATAL_ERROR)

INCLUDE(../../cmake/ConfigSGX.cmake)

SET(U_CRYPTO_ADAPT_LIB_NAME updo-crypto-adapt)
SET(T_CRYPTO_ADAPT_LIB_NAME tpdo-crypto-adapt)

SET(PDO_CRYPTO_DIR "pdo/common/crypto")

FILE(GLOB PROJECT_HEADERS "${PDO_CRYPTO_DIR}/../*.h")
FILE(GLOB PROJECT_SOURCES 
    "${PDO_CRYPTO_DIR}/../c11_support.cpp"
    "${PDO_CRYPTO_DIR}/../hex_string.cpp"
    "${PDO_CRYPTO_DIR}/../log.cpp"
    "pdo-types.cpp"
    )

###################################################################################################
# Untrusted crypto adapt library
###################################################################################################
#PROJECT(${U_CRYPTO_ADAPT_LIB_NAME} CXX)
#ADD_LIBRARY(${U_CRYPTO_ADAPT_LIB_NAME} STATIC ${PROJECT_HEADERS} ${PROJECT_SOURCES})
#TARGET_INCLUDE_DIRECTORIES(${U_CRYPTO_ADAPT_LIB_NAME} PRIVATE "${PDO_CRYPTO_DIR}/..")
#TARGET_INCLUDE_DIRECTORIES(${U_CRYPTO_ADAPT_LIB_NAME} PRIVATE "$ENV{FPC_PATH}/common/base64")
#TARGET_INCLUDE_DIRECTORIES(${U_CRYPTO_ADAPT_LIB_NAME} PUBLIC $ENV{SGX_SDK}/include)

###################################################################################################
# Trusted crypto adapt library
###################################################################################################
#PROJECT(${T_CRYPTO_ADAPT_LIB_NAME} CXX)
#ADD_LIBRARY(${T_CRYPTO_ADAPT_LIB_NAME} STATIC ${PROJECT_HEADERS} ${PROJECT_SOURCES})
#TARGET_INCLUDE_DIRECTORIES(${T_CRYPTO_ADAPT_LIB_NAME} PRIVATE "${PDO_CRYPTO_DIR}/..")
#TARGET_INCLUDE_DIRECTORIES(${T_CRYPTO_ADAPT_LIB_NAME} PRIVATE "$ENV{FPC_PATH}/common/base64")
#TARGET_INCLUDE_DIRECTORIES(${T_CRYPTO_ADAPT_LIB_NAME} PUBLIC $ENV{SGX_SDK}/include)

###################################################################################################
# PDO Crypto library
###################################################################################################
set(CMAKE_ARCHIVE_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR})
ADD_SUBDIRECTORY (${PDO_CRYPTO_DIR})
