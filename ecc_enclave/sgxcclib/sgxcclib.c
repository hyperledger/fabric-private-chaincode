/*
 * Copyright 2019 Intel Corporation
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "sgxcclib.h"
#include "enclave_u.h"
#include "sgx_attestation_type.h"

#include <stdbool.h>
#include <string.h>
#include <unistd.h>

// for RA:
#include "sgx_quote.h"
#include "sgx_report.h"
#include "sgx_uae_service.h"

#include "sgx_eid.h"  // sgx_enclave_id_t
#include "sgx_urts.h"

#define NRM "\x1B[0m"
#define RED "\x1B[31m"
#define CYN "\x1B[36m"

//TODO: separate logging in sgxcclib

// Prototypes of CGo functions implemented in ecc/enclave/enclave_stub.go
// - logging
extern void golog(const char* format, ...);

#include <stdio.h>
#include <stdarg.h>

#define BUF_SIZE 1024
#define LARGE_BUF_SIZE (BUF_SIZE*2)
void LOG_ERROR(const char* fmt, ...) {
    //create message
    char msg[BUF_SIZE];
    va_list ap;
    va_start(ap, fmt);
    vsnprintf(msg, BUF_SIZE, fmt, ap);
    va_end(ap);
    //color the message
    char colored_msg[LARGE_BUF_SIZE];
    snprintf(colored_msg, LARGE_BUF_SIZE, RED "ERROR: %s" NRM "\n", msg);
    //dump message
    golog(colored_msg);
}

void LOG_DEBUG(const char* fmt, ...) {
    //create message
    char msg[BUF_SIZE];
    va_list ap;
    va_start(ap, fmt);
    vsnprintf(msg, BUF_SIZE, fmt, ap);
    va_end(ap);
    //color the message
    char colored_msg[LARGE_BUF_SIZE];
    snprintf(colored_msg, LARGE_BUF_SIZE, CYN "DEBUG: %s" NRM "\n", msg);
    //dump message
    golog(colored_msg);
}

// - creator access
extern void get_creator_name(
    const char* msp_id, uint32_t max_msp_id_len, const char* dn, uint32_t max_dn_len, void* ctx);

// - for accessing ledger kvs
extern void get_state(const char* key,
    uint8_t* val,
    uint32_t max_val_len,
    uint32_t* val_len,
    cmac_t* cmac,
    void* ctx);
extern void get_state_by_partial_composite_key(const char* comp_key,
    uint8_t* values,
    uint32_t max_len,
    uint32_t* values_len,
    cmac_t* cmac,
    void* ctx);
extern void put_state(const char* key, uint8_t* val, uint32_t val_len, void* ctx);

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
    if (ret != SGX_SUCCESS)
    {
        LOG_ERROR("Lib: Unable to create enclave. reason: %d", ret);
        return ret;
    }

    int enclave_ret = SGX_ERROR_UNEXPECTED;
    ret = ecall_init(*eid, &enclave_ret);
    if (ret != SGX_SUCCESS || enclave_ret != SGX_SUCCESS)
    {
        LOG_ERROR("Lib: Unable to initialize enclave. reason: %d", ret);
    }
    return enclave_ret;
}

int sgxcc_destroy_enclave(enclave_id_t eid)
{
    int ret = sgx_destroy_enclave((sgx_enclave_id_t)eid);
    if (ret != SGX_SUCCESS)
    {
        LOG_ERROR("Lib: Error: %d", ret);
    }
    return ret;
}

int sgxcc_get_quote_size(uint8_t *p_sig_rl, uint32_t sig_rl_size, uint32_t *p_quote_size)
{
    *p_quote_size = 0;
    if(sig_rl_size > 0 && p_sig_rl == NULL) {
        LOG_ERROR("Lib: Error: sigrlsize not zero but null sig_rl");
        return SGX_ERROR_UNEXPECTED;
    }
    if(p_quote_size == NULL) {
        LOG_ERROR("Lib: Error: pquotesize is null");
        return SGX_ERROR_UNEXPECTED;
    }

    sgx_status_t ret = sgx_calc_quote_size(p_sig_rl, sig_rl_size, p_quote_size);
    if(ret != SGX_SUCCESS) {
        LOG_ERROR("Lib: Error: %d", ret);
    }

    return (int)ret;
}

int sgxcc_get_target_info(enclave_id_t eid, target_info_t* p_target_info)
{
    int ret = sgx_get_target_info(eid, (sgx_target_info_t*)p_target_info);
    if (ret != SGX_SUCCESS)
    {
        LOG_ERROR("Lib: ERROR - ecall_get_target_info: %d", ret);
    }

    return ret;
}

int sgxcc_get_local_attestation_report(
    enclave_id_t eid, target_info_t* target_info, report_t* report, ec256_public_t* pubkey)
{
    int enclave_ret = SGX_ERROR_UNEXPECTED;
    int ret = ecall_create_report(eid, &enclave_ret, (sgx_target_info_t*)target_info,
        (sgx_report_t*)report, (uint8_t*)pubkey);
    if (ret != SGX_SUCCESS || enclave_ret != SGX_SUCCESS)
    {
        LOG_ERROR("Lib: ERROR - ecall_create_report: %d", ret);
    }
    return enclave_ret;
}

int sgxcc_bind(enclave_id_t eid, report_t* report, ec256_public_t* pubkey)
{
    int enclave_ret = SGX_ERROR_UNEXPECTED;
    int ret = ecall_bind_tlcc(eid, &enclave_ret, (sgx_report_t*)report, (uint8_t*)pubkey);
    if (ret != SGX_SUCCESS || enclave_ret != SGX_SUCCESS)
    {
        LOG_ERROR("Lib: ERROR - ecall_bind_tlcc: %d", ret);
    }
    return enclave_ret;
}

int sgxcc_get_remote_attestation_report(
    enclave_id_t eid, quote_t* quote, uint32_t quote_size, ec256_public_t* pubkey, spid_t* spid, uint8_t *p_sig_rl, uint32_t sig_rl_size)
{
    sgx_target_info_t qe_target_info = {0};
    sgx_epid_group_id_t gid = {0};
    sgx_report_t report;

    int ret = sgx_init_quote(&qe_target_info, &gid);
    if (ret != SGX_SUCCESS)
    {
        LOG_ERROR("Lib: ERROR - sgx_init_quote: %d", ret);
        return ret;
    }

    // get report from enclave
    int enclave_ret = SGX_ERROR_UNEXPECTED;
    ret = ecall_create_report(eid, &enclave_ret, &qe_target_info, &report, (uint8_t*)pubkey);
    if (ret != SGX_SUCCESS || enclave_ret != SGX_SUCCESS)
    {
        LOG_ERROR("Lib: ERROR - ecall_create_report: %d", ret);
        return enclave_ret;
    }

    uint32_t required_quote_size = 0;
    ret = sgxcc_get_quote_size(p_sig_rl, sig_rl_size, &required_quote_size);
    if(ret != SGX_SUCCESS)
    {
        LOG_ERROR("Lib: ERROR - get quote size, %d", ret);
        return ret;
    }
    if (quote_size < required_quote_size)
    {
        LOG_ERROR("Lib: ERROR - quote size too small. Required %u have %u", required_quote_size, quote_size);
        return SGX_ERROR_OUT_OF_MEMORY;
    }

    ret = sgx_get_quote(&report, SGX_QUOTE_SIGN_TYPE,
        (sgx_spid_t*)spid,  // spid
        NULL,               // nonce
        p_sig_rl,           // sig_rl
        sig_rl_size,        // sig_rl_size
        NULL,               // p_qe_report
        (sgx_quote_t*)quote, quote_size);

    if (ret != SGX_SUCCESS)
    {
        LOG_ERROR("Lib: ERROR - sgx_get_quote: %d", ret);
    }
    return ret;
}

int sgxcc_get_pk(enclave_id_t eid, ec256_public_t* pubkey)
{
    int enclave_ret = SGX_ERROR_UNEXPECTED;
    int ret = ecall_get_pk(eid, &enclave_ret, (uint8_t*)pubkey);
    if (ret != SGX_SUCCESS || enclave_ret != SGX_SUCCESS)
    {
        LOG_ERROR("Lib: ERROR - ecall_get_pk: %d", ret);
    }
    return enclave_ret;
}

int sgxcc_init(enclave_id_t eid,
    const char* encoded_args,
    uint8_t* response,
    uint32_t response_len_in,
    uint32_t* response_len_out,
    ec256_signature_t* signature,
    void* ctx)
{
    int enclave_ret = SGX_ERROR_UNEXPECTED;
    int ret = ecall_cc_init(eid, &enclave_ret,
        encoded_args,                                 // args (encoded and potentially encrypted)
        response, response_len_in, response_len_out,  // response
        (sgx_ec256_signature_t*)signature,            // signature
        ctx);                                         // context for callback
    if (ret != SGX_SUCCESS || enclave_ret != SGX_SUCCESS)
    {
        LOG_ERROR("Lib: ERROR - invoke: %d", ret);
    }
    return enclave_ret;
}

int sgxcc_invoke(enclave_id_t eid,
    const char* encoded_args,
    const char* pk,
    uint8_t* response,
    uint32_t response_len_in,
    uint32_t* response_len_out,
    ec256_signature_t* signature,
    void* ctx)
{
    int enclave_ret = SGX_ERROR_UNEXPECTED;
    int ret = ecall_cc_invoke(eid, &enclave_ret,
        encoded_args,  // args  (encoded and potentially encrypted)
        pk,            // client pk used for args encryption, if null no encryption used
        response, response_len_in, response_len_out,  // response
        (sgx_ec256_signature_t*)signature,            // signature
        ctx);                                         // context for callback
    if (ret != SGX_SUCCESS || enclave_ret != SGX_SUCCESS)
    {
        LOG_ERROR("Lib: ERROR - invoke: %d", ret);
    }

    return enclave_ret;
}

int sgxcc_get_egid(unsigned int* p_egid) {
    sgx_target_info_t target_info = {0};
    int ret = sgx_init_quote(&target_info, (sgx_epid_group_id_t*)p_egid);
    if (ret != SGX_SUCCESS)
    {
        LOG_ERROR("Lib: ERROR - sgx_get_egid: %d", ret);
    }
    LOG_DEBUG("EGID: %u", *p_egid);
    return ret;
}

/* OCall functions */
void ocall_get_creator_name(
    char* msp_id, uint32_t max_msp_id_len, char* dn, uint32_t max_dn_len, void* ctx)
{
    get_creator_name(msp_id, max_msp_id_len, dn, max_dn_len, ctx);
}

void ocall_get_state(const char* key,
    uint8_t* val,
    uint32_t max_val_len,
    uint32_t* val_len,
    sgx_cmac_128bit_tag_t* cmac,
    void* ctx)
{
    get_state(key, val, max_val_len, val_len, (cmac_t*)cmac, ctx);
}

void ocall_put_state(const char* key, uint8_t* val, uint32_t val_len, void* ctx)
{
    put_state(key, val, val_len, ctx);
}

void ocall_get_state_by_partial_composite_key(const char* key,
    uint8_t* bids_bytes,
    uint32_t max_len,
    uint32_t* bids_bytes_len,
    sgx_cmac_128bit_tag_t* cmac,
    void* ctx)
{
    get_state_by_partial_composite_key(
        key, bids_bytes, max_len, bids_bytes_len, (cmac_t*)cmac, ctx);
}

void ocall_print_string(const char* str)
{
    golog(str);
}
