# Copyright IBM Corp. All Rights Reserved.
# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

# get arch information
execute_process(COMMAND getconf LONG_BIT OUTPUT_VARIABLE VAR_LONG_BIT)

# Check for the existence of common required SGX variables,
# and fail if these are not defined. Note global build defaults are defined `../config.mk`.
set(SGX_SDK "$ENV{SGX_SDK}")
if("${SGX_SDK} " STREQUAL " ")
    message(FATAL_ERROR "SGX_SDK: undefined environment variable")
endif()

set(SGX_ARCH "$ENV{SGX_ARCH}")
if("${SGX_ARCH} " STREQUAL " ")
    message(FATAL_ERROR "SGX_ARCH: undefined environment variable")
endif()

set(SGX_MODE "$ENV{SGX_MODE}")
if("${SGX_MODE} " STREQUAL " ")
    message(FATAL_ERROR "SGX_MODE: undefined environment variable")
endif()

set(SGX_BUILD "$ENV{SGX_BUILD}")
if("${SGX_BUILD} " STREQUAL " ")
    message(FATAL_ERROR "SGX_BUILD: undefined environment variable")
endif()

set(SGX_SSL "$ENV{SGX_SSL}")
if("${SGX_SSL} " STREQUAL " ")
    message(FATAL_ERROR "SGX_SSL: undefined environment variable")
endif()

set(SGX_COMMON_CFLAGS -m64)
set(SGX_LIBRARY_PATH ${SGX_SDK}/lib64)
set(SGX_ENCLAVE_SIGNER ${SGX_SDK}/bin/x64/sgx_sign)
set(SGX_EDGER8R ${SGX_SDK}/bin/x64/sgx_edger8r)

if (SGX_MODE STREQUAL HW)
    set(SGX_COMMON_CFLAGS "${SGX_COMMON_CFLAGS} -DSGX_HW_MODE")
    set(SGX_URTS_LIB sgx_urts)
    set(SGX_LNCH_LIB sgx_launch)
    set(SGX_EPID_LIB sgx_epid)
    set(SGX_TRTS_LIB sgx_trts)
    set(SGX_TSVC_LIB sgx_tservice)
else ()
    set(SGX_COMMON_CFLAGS "${SGX_COMMON_CFLAGS} -DSGX_SIM_MODE")
    set(SGX_URTS_LIB sgx_urts_sim)
    set(SGX_LNCH_LIB sgx_launch_sim)
    set(SGX_EPID_LIB sgx_epid_sim)
    set(SGX_TRTS_LIB sgx_trts_sim)
    set(SGX_TSVC_LIB sgx_tservice_sim)
endif (SGX_MODE STREQUAL HW)

if (SGX_BUILD STREQUAL "DEBUG")
    set(SGX_COMMON_CFLAGS "${SGX_COMMON_CFLAGS} -O0 -g")
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -DDEBUG -UNDEBUG -UEDEBUG")
elseif(SGX_BUILD STREQUAL "PRERELEASE")
    set(SGX_COMMON_CFLAGS "${SGX_COMMON_CFLAGS} -O2")
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -UDEBUG -DNDEBUG -DEDEBUG")
elseif(SGX_BUILD STREQUAL "RELEASE")
    if(SGX_MODE STREQUAL "HW")
        set(SGX_COMMON_CFLAGS "${SGX_COMMON_CFLAGS} -O2")
        set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -UDEBUG -DNDEBUG -UEDEBUG")
    else()
        message(FATAL_ERROR "HW mode must be set with RELEASE")
    endif()
else()
    message(FATAL_ERROR "Unknown build ${SGX_BUILD}")
endif()

set(SGX_SSL_LIBRARY_PATH ${SGX_SSL}/lib64)

message(STATUS "SGX_COMMON_CFLAGS: ${SGX_COMMON_CFLAGS}")
message(STATUS "SGX_SDK: ${SGX_SDK}")
message(STATUS "SGX_MODE: ${SGX_MODE}")
message(STATUS "SGX_BUILD: ${SGX_BUILD}")
message(STATUS "SGX_LIBRARY_PATH: ${SGX_LIBRARY_PATH}")
message(STATUS "SGX_ENCLAVE_SIGNER: ${SGX_ENCLAVE_SIGNER}")
message(STATUS "SGX_EDGER8R: ${SGX_EDGER8R}")
message(STATUS "SGX_SSL: ${SGX_SSL}")
message(STATUS "SGX_SSL_LIBRARY_PATH: ${SGX_SSL_LIBRARY_PATH}")
