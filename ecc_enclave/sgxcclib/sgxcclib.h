/*
 * Copyright 2019 Intel Corporation
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#ifndef _SGXCCLIB_H_
#define _SGXCCLIB_H_

#include "types.h"

#ifdef __cplusplus
extern "C" {
#endif

int sgxcc_get_remote_attestation_report(enclave_id_t eid,
    quote_t* quote,
    uint32_t quote_size,
    ec256_public_t* pubkey,
    spid_t* spid,
    uint8_t* p_sig_rl,
    uint32_t sig_rl_size);

int sgxcc_bind(enclave_id_t eid, report_t* report, ec256_public_t* pubkey);

int sgxcc_init(enclave_id_t eid,
    const char* args,
    // Note: no PK, init only/mainly useful iff it involves explicit org-approval and then should be
    // public
    uint8_t* response,
    uint32_t response_len_in,
    uint32_t* response_len_out,
    ec256_signature_t* signature,
    void* ctx);

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
