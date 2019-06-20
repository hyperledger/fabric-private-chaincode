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

#ifndef ias_h
#define ias_h

#include <stddef.h>
#include "sgx_quote.h"

#define IAS_SUCCESS 0
#define IAS_ERROR -1

#ifdef __cplusplus
extern "C" {
#endif

typedef struct mrenclave
{
    uint8_t m[32];
} mrenclave_t;

typedef struct attestation_report
{
    uint8_t* enclave_pk;
    char* report_signature;
    char* signing_cert;
    uint8_t* report_body;
} attestation_report_t;

int verify_attestation_report(uint8_t* json_data, size_t json_len, mrenclave_t* mrenclave);

#ifdef __cplusplus
}
#endif

#endif
