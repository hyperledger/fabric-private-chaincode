/*
 * Copyright 2020 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "logging.h"
#include "error.h"

int ocall_log(const char* str)
{
    COND2ERR(str == NULL);
    return loggingf(str);

err:
    return 0;
}
