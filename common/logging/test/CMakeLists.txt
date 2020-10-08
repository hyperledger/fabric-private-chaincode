# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

INCLUDE ("$ENV{FPC_PATH}/cmake/ConfigSGX.cmake")

##################################################################################################
## LOG test app:
## this app tests the untrusted log library
##################################################################################################
SET(TEST_LOG_APP test_log_app)
PROJECT(${TEST_LOG_APP} CXX)
ADD_EXECUTABLE(${TEST_LOG_APP}
    untrusted/test.cpp
    )

# SGX_COMMON_CFLAGS is set in ConfigSGX.cmake.
# Also, it sets the DO_DEBUG flag when the SGX_BUILD env var is set to DEBUG.
set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} ${SGX_COMMON_CFLAGS}")

TARGET_INCLUDE_DIRECTORIES(${TEST_LOG_APP} PRIVATE "../untrusted") # for logging.h
TARGET_INCLUDE_DIRECTORIES(${TEST_LOG_APP} PRIVATE "../../") # for error.h
TARGET_LINK_LIBRARIES(${TEST_LOG_APP} ${U_LOGGING_LIB_NAME})

# Register this application as a test
add_test(
    NAME ${TEST_LOG_APP}
    COMMAND ./${TEST_LOG_APP}
    )

IF(DEFINED TEST_TARGET)
    ADD_DEPENDENCIES(${TEST_TARGET} ${TEST_LOG_APP})
endif()

