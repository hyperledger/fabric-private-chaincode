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

#ifndef _TYPES_H_
#define _TYPES_H_

#include <stdarg.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>

typedef uint64_t enclave_id_t;
typedef uint8_t* quote_t;
typedef struct spid_t {
    uint8_t id[16];
} spid_t;

typedef uint8_t report_t[432];
typedef uint8_t target_info_t[512];
typedef uint8_t cmac_t[16];

typedef struct ec256_public_t {
    uint8_t gx[32];
    uint8_t gy[32];
} ec256_public_t;

typedef struct ec256_signature_t {
    uint32_t x[8];
    uint32_t y[8];
} ec256_signature_t;

#endif
