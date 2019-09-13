/*
 * Copyright 2019 Intel Corporation
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "trusted_ledger.h"
#include "enclave_u.h"

#include <stdbool.h>
#include <string.h>
#include <unistd.h>

#include "sgx_eid.h"  // sgx_enclave_id_t
#include "sgx_quote.h"
#include "sgx_uae_service.h"
#include "sgx_urts.h"

#define NRM "\x1B[0m"
#define RED "\x1B[31m"
#define CYN "\x1B[36m"

#define PERR(fmt, ...) golog(CYN "ERROR" RED fmt NRM "\n", ##__VA_ARGS__)

// extern go printf
extern void golog(const char* format, ...);

int tlcc_init_with_genesis(enclave_id_t eid, uint8_t* genesis, uint32_t genesis_size)
{
    int enclave_ret = -1;
    int ret = -1;

    ret = ecall_join_channel(eid, &enclave_ret, genesis, genesis_size);
    if (ret != SGX_SUCCESS || enclave_ret != SGX_SUCCESS)
    {
        PERR("Lib: Unable to join channel. reason: %d %d", ret, enclave_ret);
        return -1;
    }

    return enclave_ret;
}

int tlcc_send_block(enclave_id_t eid, uint8_t* block, uint32_t block_size)
{
    int enclave_ret = -1;
    int ret = ecall_next_block(eid, &enclave_ret, block, block_size);
    if (ret != SGX_SUCCESS || enclave_ret != SGX_SUCCESS)
    {
        PERR("Lib: ERROR Process block within enclave. reason: %d %d", ret, enclave_ret);
        return ret;
    }

    return enclave_ret;
}

int tlcc_get_state_metadata(enclave_id_t eid, const char* key, uint8_t* nonce, cmac_t* cmac)
{
    int enclave_ret = -1;
    int ret = ecall_get_state_metadata(eid, (int*)&enclave_ret, key, nonce, cmac);
    if (ret != SGX_SUCCESS)
    {
        PERR("Lib: Error: %d", ret);
        return ret;
    }

    return SGX_SUCCESS;
}

int tlcc_get_multi_state_metadata(
    enclave_id_t eid, const char* comp_key, uint8_t* nonce, cmac_t* cmac)
{
    int enclave_ret = -1;
    int ret = ecall_get_multi_state_metadata(eid, (int*)&enclave_ret, comp_key, nonce, cmac);
    if (ret != SGX_SUCCESS)
    {
        PERR("Lib: Error: %d", ret);
        return ret;
    }

    return SGX_SUCCESS;
}

// this is only for debugging
int tlcc_print_state(enclave_id_t eid)
{
    int enclave_ret = -1;
    int ret = ecall_print_state(eid, (int*)&enclave_ret);
    if (ret != SGX_SUCCESS)
    {
        PERR("Lib: Error: %d", ret);
        return ret;
    }

    return SGX_SUCCESS;
}

/* OCall functions */
void ocall_print_string(const char* str)
{
    golog(str);
}
