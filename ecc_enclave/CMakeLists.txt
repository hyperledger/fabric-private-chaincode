# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

cmake_minimum_required(VERSION 3.5.1)
project(ecclib)

include(../cmake/Init.cmake)
include(../cmake/NanoPB.cmake)

include(${CMAKE_MODULE_PATH}/ConfigSGX.cmake)

set(COMMON_SOURCE_DIR ${PROJECT_SOURCE_DIR}/../common)

add_subdirectory(enclave)
add_subdirectory(sgxcclib)

set(BUILD_DIR ${CMAKE_LIBRARY_OUTPUT_DIRECTORY}/include ${CMAKE_BINARY_DIR}/lib)
set(DEPLOY_DIR ${PROJECT_SOURCE_DIR}/../ecc/enclave)
