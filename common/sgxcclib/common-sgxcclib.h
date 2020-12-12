/*
 * Copyright 2019 Intel Corporation
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#ifndef _COMMON_SGXCCLIB_H_
#define _COMMON_SGXCCLIB_H_

#include "fpc-types.h"

#ifdef __cplusplus
extern "C" {
#endif

int sgxcc_create_enclave(enclave_id_t* eid,
    const char* enclave_file,
    uint8_t* attestation_parameters,
    uint32_t ap_size,
    uint8_t* cc_parameters,
    uint32_t ccp_size,
    uint8_t* host_parameters,
    uint32_t hp_size,
    uint8_t* credentials,
    uint32_t credentials_max_size,
    uint32_t* credentials_size);
int sgxcc_destroy_enclave(enclave_id_t eid);
int sgxcc_get_quote_size(uint8_t* p_sig_rl, uint32_t sig_rl_size, uint32_t* p_quote_size);
int sgxcc_get_target_info(enclave_id_t eid, target_info_t* target_info);
int sgxcc_get_local_attestation_report(
    enclave_id_t eid, target_info_t* target_info, report_t* report, ec256_public_t* pubkey);
int sgxcc_get_egid(unsigned int* p_egid);

#ifdef __cplusplus
}
#endif /* __cplusplus */

#endif /* !_COMMON_SGXCCLIB_H_ */
