# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

cmake_minimum_required(VERSION 3.5.1)
project(tllib)
enable_testing()

include(../cmake/Init.cmake)

include(${CMAKE_MODULE_PATH}/ConfigSGX.cmake)

set(COMMON_SOURCE_DIR ${PROJECT_SOURCE_DIR}/../common)

add_subdirectory(enclave)
add_subdirectory(trusted_ledger)
add_subdirectory(test)

set(BUILD_DIR ${CMAKE_LIBRARY_OUTPUT_DIRECTORY}/include ${CMAKE_BINARY_DIR}/lib)
