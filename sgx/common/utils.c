/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "utils.h"
#include <stdlib.h>
#include <string.h>

int append_string(char* buf, const char* string)
{
    int len = (int)strlen(string);
    if (buf == NULL)
    {
        return len;
    }
    strncpy(buf, string, len + 1);
    return len;
}

void bytes_swap(void* bytes, size_t len)
{
    unsigned char *start, *end;
    for (start = (unsigned char*)bytes, end = start + len - 1; start < end; ++start, --end)
    {
        unsigned char swap = *start;
        *start = *end;
        *end = swap;
    }
}
