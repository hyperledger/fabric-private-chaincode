# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

cmake_minimum_required(VERSION 3.5.1)
project(FPC_DEMO_EAS)

set(SOURCE_FILES
    entry.cpp
    dispatcher.cpp
    chaincode.cpp
    errors.cpp
    messages.cpp
    _protos/irb.pb.c
    storage.cpp
    id.cpp
    experiment.cpp
    study.cpp
    signedapproval.cpp
    evaluationpack.cpp
    )

# NANOPB for protos
include($ENV{FPC_PATH}/cmake/NanoPB.cmake)
include_directories(
    ${CMAKE_CURRENT_SOURCE_DIR}
    ${NANOPB_PATH}
    )

# FPC Common chaincode
include($ENV{FPC_PATH}/ecc_enclave/enclave/CMakeLists-common-app-enclave.txt)
