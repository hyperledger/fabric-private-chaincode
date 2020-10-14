/*
 * Copyright 2020 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "logging.h"
#include <stdarg.h>
#include "error.h"

static log_callback_f g_log_callback = puts;

bool logging_set_callback(log_callback_f log_callback)
{
    COND2ERR(log_callback == NULL);

    g_log_callback = log_callback;
    return true;

err:
    return false;
}

int loggingf(const char* fmt, ...)
{
    COND2ERR(g_log_callback == NULL);

    char buf[BUFSIZ] = {'\0'};
    va_list ap;
    int n;

    va_start(ap, fmt);
    n = vsnprintf(buf, BUFSIZ, fmt, ap);
    va_end(ap);

    char* pbuf;
    if (n >= 0 || n < BUFSIZ)
        pbuf = buf;
    else
        pbuf = ERROR_LOG_STRING;

    n = g_log_callback(pbuf);
    COND2ERR(n < 0);

    COND2ERR(pbuf != buf);  // if outputted error log, fail

    return 1;

err:
    return 0;
}
