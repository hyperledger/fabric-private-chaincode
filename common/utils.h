/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#ifndef utils_h
#define utils_h

#include <stdint.h>
#include <stdio.h>

#define HASH_SIZE 32

#ifdef __cplusplus
extern "C" {
#endif

int append_string(char* buf, const char* string);
void bytes_swap(void* bytes, size_t len);
char* bytes_to_hexstring(uint8_t* bytes, size_t len);

#ifdef __cplusplus
}
#endif

#endif
