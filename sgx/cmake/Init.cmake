# Copyright IBM Corp. All Rights Reserved.
# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

set(CMAKE_MODULE_PATH ${CMAKE_MODULE_PATH} "${PROJECT_SOURCE_DIR}/../cmake/")
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)
set(CMAKE_VERBOSE_MAKEFILE ON)
set(CMAKE_RUNTIME_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR})
set(CMAKE_LIBRARY_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR})


set(FPC_PATH "$ENV{FPC_PATH}")
if("${FPC_PATH} " STREQUAL " ")
    message(FATAL_ERROR "FPC_PATH: undefined environment variable")
endif()
