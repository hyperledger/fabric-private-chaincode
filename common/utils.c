/*
* Copyright IBM Corp. 2018 All Rights Reserved.
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

#include "utils.h"
#include <stdlib.h>
#include <string.h>

int append_string(char *buf, const char *string)
{
    int len = (int)strlen(string);
    if (buf == NULL) {
        return len;
    }
    strncpy(buf, string, len + 1);
    return len;
}

void bytes_swap(void *bytes, size_t len)
{
    unsigned char *start, *end;
    for (start = (unsigned char *)bytes, end = start + len - 1; start < end; ++start, --end) {
        unsigned char swap = *start;
        *start = *end;
        *end = swap;
    }
}

char *bytes_to_hexstring(uint8_t *bytes, size_t len)
{
    const char *hexdigs = "0123456789abcdef";
    size_t k = len * 2 + 1;
    char *out = malloc(k);
    for (int i = 0; i < len; i++) {
        out[i * 2] = hexdigs[bytes[i] >> 4];
        out[i * 2 + 1] = hexdigs[bytes[i] & 0x0f];
    }
    out[k - 1] = '\0';
    return out;
}
