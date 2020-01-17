/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#pragma once

#include "sgx_tcrypto.h"

int check_cmac(const char* key,
    uint8_t* nonce,
    sgx_sha256_hash_t* state_hash,
    sgx_cmac_128bit_key_t* cmac_key,
    sgx_cmac_128bit_tag_t* cmac);
int encrypt_state(sgx_aes_gcm_128bit_key_t* key,
    uint8_t* plain,
    uint32_t plain_len,
    uint8_t* cipher,
    uint32_t cipher_len);
int decrypt_state(sgx_aes_gcm_128bit_key_t* key,
    uint8_t* cipher,
    uint32_t cipher_len,
    uint8_t* plain,
    uint32_t plain_len);
int get_random_bytes(uint8_t* buffer, size_t length);
