/*
 * Copyright 2019 Intel Corporation
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "common-sgxcclib.h"
#include <unistd.h>     // access
#include "enclave_u.h"  //ecall_init, ...
#include "sgx_attestation_type.h"
#include "sgx_eid.h"  // sgx_enclave_id_t
#include "sgx_error.h"
#include "sgx_uae_service.h"
#include "sgx_urts.h"

// TODO: separate logging in sgxcclib

// Prototypes of CGo functions implemented in ecc/enclave/enclave_stub.go
// - logging
#define NRM "\x1B[0m"
#define RED "\x1B[31m"
#define CYN "\x1B[36m"
extern void golog(const char* format, ...);

#include <stdarg.h>
#include <stdio.h>

#define BUF_SIZE 1024
#define LARGE_BUF_SIZE (BUF_SIZE * 2)
void LOG_ERROR(const char* fmt, ...)
{
    // create message
    char msg[BUF_SIZE];
    va_list ap;
    va_start(ap, fmt);
    vsnprintf(msg, BUF_SIZE, fmt, ap);
    va_end(ap);
    // color the message
    char colored_msg[LARGE_BUF_SIZE];
    snprintf(colored_msg, LARGE_BUF_SIZE, RED "ERROR: %s" NRM "\n", msg);
    // dump message
    golog(colored_msg);
}

void LOG_DEBUG(const char* fmt, ...)
{
    // create message
    char msg[BUF_SIZE];
    va_list ap;
    va_start(ap, fmt);
    vsnprintf(msg, BUF_SIZE, fmt, ap);
    va_end(ap);
    // color the message
    char colored_msg[LARGE_BUF_SIZE];
    snprintf(colored_msg, LARGE_BUF_SIZE, CYN "DEBUG: %s" NRM "\n", msg);
    // dump message
    golog(colored_msg);
}

int sgxcc_create_enclave(sgx_enclave_id_t* eid, const char* enclave_file)
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
    ret = ecall_init(*eid, &enclave_ret);
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

int sgxcc_get_remote_attestation_report(enclave_id_t eid,
    quote_t* quote,
    uint32_t quote_size,
    ec256_public_t* pubkey,
    spid_t* spid,
    uint8_t* p_sig_rl,
    uint32_t sig_rl_size)
{
    sgx_target_info_t qe_target_info = {0};
    sgx_epid_group_id_t gid = {0};
    sgx_report_t report;

    int ret = sgx_init_quote(&qe_target_info, &gid);
    CHECK_SGX_ERROR_AND_RETURN_ON_ERROR(ret)

    // get report from enclave
    int enclave_ret = SGX_ERROR_UNEXPECTED;
    ret = ecall_create_report(eid, &enclave_ret, &qe_target_info, &report, (uint8_t*)pubkey);
    CHECK_SGX_ERROR_AND_RETURN_ON_ERROR(ret)
    CHECK_SGX_ERROR_AND_RETURN_ON_ERROR(enclave_ret)

    uint32_t required_quote_size = 0;
    ret = sgxcc_get_quote_size(p_sig_rl, sig_rl_size, &required_quote_size);
    CHECK_SGX_ERROR_AND_RETURN_ON_ERROR(ret)
    if (quote_size < required_quote_size)
    {
        LOG_ERROR("Lib: ERROR - quote size too small. Required %u have %u", required_quote_size,
            quote_size);
        return SGX_ERROR_OUT_OF_MEMORY;
    }

    ret = sgx_get_quote(&report, SGX_QUOTE_SIGN_TYPE,
        (sgx_spid_t*)spid,  // spid
        NULL,               // nonce
        p_sig_rl,           // sig_rl
        sig_rl_size,        // sig_rl_size
        NULL,               // p_qe_report
        (sgx_quote_t*)quote, quote_size);
    CHECK_SGX_ERROR_AND_RETURN_ON_ERROR(ret)
    return SGX_SUCCESS;
}

int sgxcc_get_pk(enclave_id_t eid, ec256_public_t* pubkey)
{
    int enclave_ret = SGX_ERROR_UNEXPECTED;
    int ret = ecall_get_pk(eid, &enclave_ret, (uint8_t*)pubkey);
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
