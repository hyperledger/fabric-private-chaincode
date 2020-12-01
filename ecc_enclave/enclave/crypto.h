/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#pragma once

#include "sgx_tcrypto.h"

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
