/*
 * Copyright IBM Corp. 2018 All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
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
