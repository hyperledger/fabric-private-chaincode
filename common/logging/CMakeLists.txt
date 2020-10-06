# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

CMAKE_MINIMUM_REQUIRED(VERSION 3.2 FATAL_ERROR)

INCLUDE(../../cmake/ConfigSGX.cmake)

INCLUDE(CMakeVariables.txt)

FILE(GLOB PROJECT_OCALLS
    "ocalls/*.c"
    )

FILE(GLOB PROJECT_TRUSTED_SOURCES
    "trusted/*.c"
    )

FILE(GLOB PROJECT_UNTRUSTED_SOURCES
    "untrusted/*.c"
    )

###################################################################################################
# Untrusted logging library
###################################################################################################
PROJECT(${U_LOGGING_LIB_NAME} CXX C)
ADD_LIBRARY(${U_LOGGING_LIB_NAME} STATIC ${PROJECT_UNTRUSTED_SOURCES} ${PROJECT_OCALLS})

TARGET_INCLUDE_DIRECTORIES(${U_LOGGING_LIB_NAME} PRIVATE "untrusted/")
TARGET_INCLUDE_DIRECTORIES(${U_LOGGING_LIB_NAME} PRIVATE "../") # path for error.h

###################################################################################################
# Trusted logging library
###################################################################################################
PROJECT(${T_LOGGING_LIB_NAME} CXX C)
ADD_LIBRARY(${T_LOGGING_LIB_NAME} STATIC ${PROJECT_TRUSTED_SOURCES})

TARGET_INCLUDE_DIRECTORIES(${T_LOGGING_LIB_NAME} PRIVATE "trusted/")
TARGET_INCLUDE_DIRECTORIES(${T_LOGGING_LIB_NAME} PRIVATE "../") # for error.h
TARGET_INCLUDE_DIRECTORIES(${T_LOGGING_LIB_NAME} PUBLIC "${SGX_SDK}/include/tlibc")

###################################################################################################
# Test
######i#############################################################################################
enable_testing()
ADD_SUBDIRECTORY(test)
