/*
 * Copyright IBM Corp. All Rights Reserved.
 * Copyright 2020 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#ifndef logging_h
#define logging_h

#include <stdbool.h>
#include "../include/log-defines.h"

#undef TAG
#define TAG "[Enclave] "

#ifdef __cplusplus
extern "C" {
#endif
int printf(const char* format, ...);
#ifdef __cplusplus
}
#endif

#endif
