/*
 * Copyright 2019 Intel Corporation
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "common-sgxcclib.h"

#include <sgx_uae_epid.h>  // epid-based attestation ...
#include <unistd.h>        // access

#include "check-sgx-error.h"
#include "enclave_u.h"  //ecall_init, ...
#include "logging.h"

int sgxcc_create_enclave(sgx_enclave_id_t* eid,
    const char* enclave_file,
    uint8_t* attestation_parameters,
    uint32_t ap_size,
    uint8_t* cc_parameters,
    uint32_t ccp_size,
    uint8_t* host_parameters,
    uint32_t hp_size,
    uint8_t* credentials,
    uint32_t credentials_max_size,
    uint32_t* credentials_size)
{
    if (access(enclave_file, F_OK) == -1)
    {
        LOG_ERROR("Lib: enclave file does not exist! %s", enclave_file);
        return SGX_ERROR_UNEXPECTED;
    }

    sgx_launch_token_t token = {0};
    int updated = 0;

    int ret = sgx_create_enclave(enclave_file, SGX_DEBUG_FLAG, &token, &updated, eid, NULL);
    CHECK_SGX_ERROR_AND_RETURN_ON_ERROR(ret)

    int enclave_ret = SGX_ERROR_UNEXPECTED;
    ret = ecall_init(*eid, &enclave_ret, attestation_parameters, ap_size, cc_parameters, ccp_size,
        host_parameters, hp_size, credentials, credentials_max_size, credentials_size);
    CHECK_SGX_ERROR_AND_RETURN_ON_ERROR(ret)
    CHECK_SGX_ERROR_AND_RETURN_ON_ERROR(enclave_ret)

    return SGX_SUCCESS;
}

int sgxcc_destroy_enclave(enclave_id_t eid)
{
    int ret = sgx_destroy_enclave((sgx_enclave_id_t)eid);
    CHECK_SGX_ERROR_AND_RETURN_ON_ERROR(ret)
    return SGX_SUCCESS;
}

int sgxcc_get_quote_size(uint8_t* p_sig_rl, uint32_t sig_rl_size, uint32_t* p_quote_size)
{
    *p_quote_size = 0;
    if (sig_rl_size > 0 && p_sig_rl == NULL)
    {
        LOG_ERROR("Lib: Error: sigrlsize not zero but null sig_rl");
        return SGX_ERROR_UNEXPECTED;
    }
    if (p_quote_size == NULL)
    {
        LOG_ERROR("Lib: Error: pquotesize is null");
        return SGX_ERROR_UNEXPECTED;
    }

    sgx_status_t ret = sgx_calc_quote_size(p_sig_rl, sig_rl_size, p_quote_size);
    CHECK_SGX_ERROR_AND_RETURN_ON_ERROR(ret)

    return SGX_SUCCESS;
}

int sgxcc_get_target_info(enclave_id_t eid, target_info_t* p_target_info)
{
    int ret = sgx_get_target_info(eid, (sgx_target_info_t*)p_target_info);
    CHECK_SGX_ERROR_AND_RETURN_ON_ERROR(ret)
    return SGX_SUCCESS;
}

int sgxcc_get_local_attestation_report(
    enclave_id_t eid, target_info_t* target_info, report_t* report, ec256_public_t* pubkey)
{
    int enclave_ret = SGX_ERROR_UNEXPECTED;
    int ret = ecall_create_report(eid, &enclave_ret, (sgx_target_info_t*)target_info,
        (sgx_report_t*)report, (uint8_t*)pubkey);
    CHECK_SGX_ERROR_AND_RETURN_ON_ERROR(ret)
    CHECK_SGX_ERROR_AND_RETURN_ON_ERROR(enclave_ret)
    return SGX_SUCCESS;
}

int sgxcc_get_egid(unsigned int* p_egid)
{
    sgx_target_info_t target_info = {0};
    int ret = sgx_init_quote(&target_info, (sgx_epid_group_id_t*)p_egid);
    LOG_DEBUG("EGID: %u", *p_egid);
    CHECK_SGX_ERROR_AND_RETURN_ON_ERROR(ret)
    return SGX_SUCCESS;
}
