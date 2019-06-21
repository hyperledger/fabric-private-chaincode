/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
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
