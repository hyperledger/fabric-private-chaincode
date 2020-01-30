# Copyright 2019 Intel Corporation
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

cmake_minimum_required(VERSION 3.5.1)

set(SOURCE_FILES
    dispatcher.cpp
    entry.cpp
    error-codes.cpp
    spectrum-auction.cpp
    spectrum-auction-message.cpp
    storage.cpp
    principal.cpp
    territory.cpp
    eligibility.cpp
    bidder.cpp
    auction-state.cpp
    utils.cpp
    evaluate.cpp
    bid.cpp
    )

include($ENV{FPC_PATH}/ecc_enclave/enclave/CMakeLists-common-app-enclave.txt)

