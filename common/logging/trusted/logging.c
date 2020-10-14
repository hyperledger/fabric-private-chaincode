/*
 * Copyright IBM Corp. All Rights Reserved.
 * Copyright 2020 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "logging.h"
#include <stdio.h>  //for va_list
#include <string.h>
#include "error.h"

int ocall_log(int* retval, const char* str);

int loggingf(const char* fmt, ...)
{
    char buf[BUFSIZ] = {'\0'};
    va_list ap;
    va_start(ap, fmt);
    int n = vsnprintf(buf, BUFSIZ, fmt, ap);
    va_end(ap);

    char* pbuf;
    if (n >= 0 || n < BUFSIZ)
        pbuf = buf;
    else
        pbuf = ERROR_LOG_STRING;

    int sgxstatus, ret;
    sgxstatus = ocall_log(&ret, buf);
    COND2ERR(sgxstatus != 0);

    COND2ERR(pbuf != buf);  // if outputted error log, fail

    return ret;

err:
    return 0;
}
