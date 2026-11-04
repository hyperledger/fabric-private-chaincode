/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#ifndef _FPC_TYPES_H_
#define _FPC_TYPES_H_

#include <stdarg.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>

typedef uint64_t enclave_id_t;
typedef uint8_t* quote_t;
typedef struct spid_t
{
    uint8_t id[16];
} spid_t;

typedef uint8_t report_t[432];
typedef uint8_t target_info_t[512];
typedef uint8_t cmac_t[16];

typedef struct ec256_public_t
{
    uint8_t gx[32];
    uint8_t gy[32];
} ec256_public_t;

typedef struct ec256_signature_t
{
    uint32_t x[8];
    uint32_t y[8];
} ec256_signature_t;

#endif
