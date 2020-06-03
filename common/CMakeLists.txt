# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

CMAKE_MINIMUM_REQUIRED(VERSION 3.2 FATAL_ERROR)

set(COMMON_SOURCE_DIR ${CMAKE_CURRENT_SOURCE_DIR})

set(SOURCE_FILES
    ${CMAKE_CURRENT_SOURCE_DIR}/crypto/pdo/common/c11_support.cpp)

add_library(pdo-utils STATIC ${SOURCE_FILES})
