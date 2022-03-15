# Copyright 2022 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

FUNCTION(COND_ENABLE_DEBUG TARGET_LIB)
    IF($ENV{FPC_GDB_DEBUG_ENABLED})
        MESSAGE("FPC Debug enabled")
        TARGET_COMPILE_OPTIONS(${TARGET_LIB} PUBLIC -g)
    ELSE()
        MESSAGE("FPC Debug disabled")
    ENDIF()
ENDFUNCTION()
