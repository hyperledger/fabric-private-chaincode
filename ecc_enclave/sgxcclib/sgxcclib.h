/*
 * Copyright 2019 Intel Corporation
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#ifndef _SGXCCLIB_H_
#define _SGXCCLIB_H_

#include "fpc-types.h"

#ifdef __cplusplus
extern "C" {
#endif

int sgxcc_bind(enclave_id_t eid, report_t* report, ec256_public_t* pubkey);

int sgxcc_invoke(enclave_id_t eid,
    const char* args,
    const char* pk,  // client pk used for args encryption, if null
                     // no encryption used
    uint8_t* response,
    uint32_t response_len_in,
    uint32_t* response_len_out,
    ec256_signature_t* signature,
    void* ctx);

#ifdef __cplusplus
}
#endif /* __cplusplus */

#endif /* !_SGXCCLIB_H_ */
