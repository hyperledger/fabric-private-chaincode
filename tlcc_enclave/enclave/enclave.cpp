/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "enclave_t.h"

#include "logging.h"
#include "utils.h"

#include <assert.h>
#include <stdlib.h>  // for malloc etc
#include <string.h>  // for memcpy etc

#include "sgx_utils.h"

#include "ledger.h"

// this is currently hardcoded to simplifying prototyping!!!
// Note that you need a session_key per chaincode enclave
static sgx_cmac_128bit_key_t session_key = {
    0x3F, 0xE2, 0x59, 0xDF, 0x62, 0x7F, 0xEF, 0x99, 0x5B, 0x4B, 0x00, 0xDE, 0x44, 0xC1, 0x26, 0x33};

// creates new identity if not exists
int ecall_join_channel(uint8_t* genesis, uint32_t gen_len)
{
    // init ledger
    init_ledger();

    // parse genesis block
    int sgx_ret = parse_block(genesis, gen_len);
    if (sgx_ret != SGX_SUCCESS)
    {
        LOG_ERROR("Parsing genesis block failed: %d", sgx_ret);
        return sgx_ret;
    }

    return SGX_SUCCESS;
}

int ecall_next_block(uint8_t* block_bytes, uint32_t block_size)
{
    return parse_block(block_bytes, block_size);
}

int ecall_get_state_metadata(const char* key, uint8_t* nonce, sgx_cmac_128bit_tag_t* cmac)
{
    sgx_sha256_hash_t state_hash = {0};
    ledger_get_state_hash(key, (uint8_t*)&state_hash);

    // remove channel name prefix from key if exists
    std::string kkey(key);
    size_t found = kkey.find_first_of(".");
    if (found != std::string::npos)
    {
        kkey.erase(0, found + 1);
    }

    // hash( key || nonce || target_hash || result )
    sgx_cmac_state_handle_t cmac_handle;
    sgx_cmac128_init(&session_key, &cmac_handle);
    sgx_cmac128_update((const uint8_t*)kkey.c_str(), kkey.size(), cmac_handle);
    // TODO use the nonce
    /* sgx_cmac128_update(nonce, 32, cmac_handle); */
    sgx_cmac128_update(state_hash, sizeof(sgx_sha256_hash_t), cmac_handle);
    sgx_cmac128_final(cmac_handle, cmac);
    sgx_cmac128_close(cmac_handle);

    return SGX_SUCCESS;
}

int ecall_get_multi_state_metadata(
    const char* comp_key, uint8_t* nonce, sgx_cmac_128bit_tag_t* cmac)
{
    // create state hash
    sgx_sha256_hash_t state_hash = {0};
    ledger_get_multi_state_hash(comp_key, (uint8_t*)&state_hash);

    // remove prefix
    std::string k(comp_key);
    size_t found = k.find_first_of(".");
    if (found != std::string::npos)
    {
        k.erase(0, found);
    }

    // hash( key || nonce || target_hash || result )
    sgx_cmac_state_handle_t cmac_handle;
    sgx_cmac128_init(&session_key, &cmac_handle);
    sgx_cmac128_update((const uint8_t*)k.c_str(), k.size(), cmac_handle);
    // TODO use the nonce
    /* sgx_cmac128_update(nonce, 32, cmac_handle); */
    sgx_cmac128_update(state_hash, sizeof(sgx_sha256_hash_t), cmac_handle);
    sgx_cmac128_final(cmac_handle, cmac);
    sgx_cmac128_close(cmac_handle);

    return SGX_SUCCESS;
}

int ecall_print_state()
{
    return print_state();
}
