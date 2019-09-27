/*
 * Copyright 2019 Intel Corporation
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "trusted_ledger.h"
#include "common-sgxcclib.h"
#include "enclave_u.h"

#include <stdbool.h>
#include <string.h>
#include <unistd.h>

#include "sgx_eid.h"  // sgx_enclave_id_t
#include "sgx_quote.h"
#include "sgx_uae_service.h"
#include "sgx_urts.h"

int tlcc_init_with_genesis(enclave_id_t eid, uint8_t* genesis, uint32_t genesis_size)
{
    int enclave_ret = SGX_ERROR_UNEXPECTED;
    int ret = ecall_join_channel(eid, &enclave_ret, genesis, genesis_size);
    CHECK_SGX_ERROR_AND_RETURN_ON_ERROR(ret)
    CHECK_SGX_ERROR_AND_RETURN_ON_ERROR(enclave_ret)
    return SGX_SUCCESS;
}

int tlcc_send_block(enclave_id_t eid, uint8_t* block, uint32_t block_size)
{
    int enclave_ret = SGX_ERROR_UNEXPECTED;
    int ret = ecall_next_block(eid, &enclave_ret, block, block_size);
    CHECK_SGX_ERROR_AND_RETURN_ON_ERROR(ret)
    CHECK_SGX_ERROR_AND_RETURN_ON_ERROR(enclave_ret)
    return SGX_SUCCESS;
}

int tlcc_get_state_metadata(enclave_id_t eid, const char* key, uint8_t* nonce, cmac_t* cmac)
{
    int enclave_ret = SGX_ERROR_UNEXPECTED;
    int ret = ecall_get_state_metadata(eid, (int*)&enclave_ret, key, nonce, cmac);
    CHECK_SGX_ERROR_AND_RETURN_ON_ERROR(ret)
    CHECK_SGX_ERROR_AND_RETURN_ON_ERROR(enclave_ret)
    return SGX_SUCCESS;
}

int tlcc_get_multi_state_metadata(
    enclave_id_t eid, const char* comp_key, uint8_t* nonce, cmac_t* cmac)
{
    int enclave_ret = SGX_ERROR_UNEXPECTED;
    int ret = ecall_get_multi_state_metadata(eid, (int*)&enclave_ret, comp_key, nonce, cmac);
    CHECK_SGX_ERROR_AND_RETURN_ON_ERROR(ret)
    CHECK_SGX_ERROR_AND_RETURN_ON_ERROR(enclave_ret)
    return SGX_SUCCESS;
}

// this is only for debugging
int tlcc_print_state(enclave_id_t eid)
{
    int enclave_ret = SGX_ERROR_UNEXPECTED;
    int ret = ecall_print_state(eid, (int*)&enclave_ret);
    CHECK_SGX_ERROR_AND_RETURN_ON_ERROR(ret)
    CHECK_SGX_ERROR_AND_RETURN_ON_ERROR(enclave_ret)
    return SGX_SUCCESS;
}

/* OCall functions */
void ocall_print_string(const char* str)
{
    golog(str);
}
