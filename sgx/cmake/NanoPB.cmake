# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

unset(NANOPB_PATH CACHE)

set(NANOPB_PATH "$ENV{NANOPB_PATH}")
if("${NANOPB_PATH} " STREQUAL " ")
    message(FATAL_ERROR "$NANOPB_PATH not set")
endif()

message(STATUS "NANOPB_PATH: ${NANOPB_PATH}")

