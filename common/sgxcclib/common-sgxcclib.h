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

int sgxcc_create_enclave(enclave_id_t* eid, const char* enclave_file);
int sgxcc_destroy_enclave(enclave_id_t eid);
int sgxcc_get_quote_size(uint8_t* p_sig_rl, uint32_t sig_rl_size, uint32_t* p_quote_size);
int sgxcc_get_target_info(enclave_id_t eid, target_info_t* target_info);
int sgxcc_get_local_attestation_report(
    enclave_id_t eid, target_info_t* target_info, report_t* report, ec256_public_t* pubkey);
int sgxcc_get_remote_attestation_report(enclave_id_t eid,
    quote_t* quote,
    uint32_t quote_size,
    ec256_public_t* pubkey,
    spid_t* spid,
    uint8_t* p_sig_rl,
    uint32_t sig_rl_size);
int sgxcc_get_pk(enclave_id_t eid, ec256_public_t* pubkey);
int sgxcc_get_egid(unsigned int* p_egid);

#define CHECK_SGX_ERROR_AND_RETURN_ON_ERROR(sgx_status_ret)                                        \
    if (sgx_status_ret != SGX_SUCCESS)                                                             \
    {                                                                                              \
        LOG_ERROR(                                                                                 \
            "Lib: ERROR - %s:%d: " #sgx_status_ret "=%d", __FUNCTION__, __LINE__, sgx_status_ret); \
        return sgx_status_ret;                                                                     \
    }

#ifdef __cplusplus
}
#endif /* __cplusplus */

#endif /* !_COMMON_SGXCCLIB_H_ */
