/*
 * Copyright 2020 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include <sgx_uae_epid.h>
#include "error.h"
#include "fpc-types.h"
#include "logging.h"
#include "sgx_quote.h"

void ocall_init_quote(uint8_t* target, uint32_t target_len, uint8_t* egid, uint32_t egid_len)
{
    int ret = sgx_init_quote((sgx_target_info_t*)target, (sgx_epid_group_id_t*)egid);
    COND2LOGERR(ret != SGX_SUCCESS, "error init quote");
err:
    // nothing to do, error will be reported at quote time
    ;
}

void ocall_get_quote(uint8_t* spid,
    uint32_t spid_len,
    uint8_t* sig_rl,
    uint32_t sig_rl_len,
    uint32_t sign_type,
    uint8_t* report,
    uint32_t report_len,
    uint8_t* quote,
    uint32_t max_quote_len,
    uint32_t* actual_quote_len)
{
    int ret;
    uint32_t required_quote_size = 0;
    ret = sgx_calc_quote_size(sig_rl, sig_rl_len, &required_quote_size);
    COND2LOGERR(ret != SGX_SUCCESS, "cannot get quote size");
    COND2LOGERR(required_quote_size > max_quote_len, "not enough buffer for quote");

    ret = sgx_get_quote((const sgx_report_t*)report, (sgx_quote_sign_type_t)sign_type,
        (const sgx_spid_t*)spid,  // spid
        NULL,                     // nonce
        sig_rl,                   // sig_rl
        sig_rl_len,               // sig_rl_size
        NULL,                     // p_qe_report
        (sgx_quote_t*)quote, required_quote_size);
    COND2LOGERR(ret != SGX_SUCCESS, "error getting quote");
    *actual_quote_len = required_quote_size;

    return;

err:
    // if anything wrong, no quote
    *actual_quote_len = 0;
}
