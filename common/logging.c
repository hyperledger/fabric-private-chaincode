/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "logging.h"

#include <stdio.h>  //for va_list
#include <string.h>

void ocall_log(const char* str);

int printf(const char* fmt, ...)
{
    char buf[BUFSIZ] = {'\0'};
    va_list ap;
    va_start(ap, fmt);
    vsnprintf(buf, BUFSIZ, fmt, ap);
    va_end(ap);
    ocall_log(buf);
    return strlen(buf);
}
