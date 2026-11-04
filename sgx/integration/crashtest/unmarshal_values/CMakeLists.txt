# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

cmake_minimum_required(VERSION 3.5.1)

set(SOURCE_FILES cc.cpp)

include($ENV{FPC_PATH}/sgx/ecc_enclave/enclave/CMakeLists-common-app-enclave.txt)
