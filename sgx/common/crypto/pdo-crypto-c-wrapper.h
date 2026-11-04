/*
 * Copyright 2020 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#pragma once

#ifdef __cplusplus
extern "C" {
#endif

extern const unsigned int SYM_KEY_LEN;
extern const unsigned int IV_LEN;
extern const unsigned int TAG_LEN;

extern const unsigned int RSA_PLAINTEXT_LEN;
extern const unsigned int RSA_KEY_SIZE;

bool compute_hash(uint8_t* message,
    uint32_t message_len,
    uint8_t* hash,
    uint32_t max_hash_len,
    uint32_t* actual_hash_len);

bool verify_signature(uint8_t* public_key,
        uint32_t public_key_len,
        uint8_t* message,
        uint32_t message_len,
        uint8_t* signature,
        uint32_t signature_len);

bool sign_message(uint8_t* private_key,
        uint32_t private_key_len,
        uint8_t* message,
        uint32_t message_len,
        uint8_t* signature,
        uint32_t signature_len,
        uint32_t* signature_actual_len);

bool pk_encrypt_message(uint8_t* public_key,
        uint32_t public_key_len,
        uint8_t* message,
        uint32_t message_len,
        uint8_t* encrypted_message,
        uint32_t encrypted_message_len,
        uint32_t* encrypted_message_actual_len);

bool pk_decrypt_message(uint8_t* private_key,
        uint32_t private_key_len,
        uint8_t* encrypted_message,
        uint32_t encrypted_message_len,
        uint8_t* message,
        uint32_t message_len,
        uint32_t* message_actual_len);

bool encrypt_message(uint8_t* key,
        uint32_t key_len,
        uint8_t* message,
        uint32_t message_len,
        uint8_t* encrypted_message,
        uint32_t encrypted_message_len,
        uint32_t* encrypted_message_actual_len);

bool decrypt_message(uint8_t* key,
        uint32_t key_len,
        uint8_t* encrypted_message,
        uint32_t encrypted_message_len,
        uint8_t* message,
        uint32_t message_len,
        uint32_t* message_actual_len);

bool new_rsa_key(uint8_t* public_key,
        uint32_t public_key_len,
        uint32_t* public_key_actual_len,
        uint8_t* private_key,
        uint32_t private_key_len,
        uint32_t* private_key_actual_len);

bool new_ecdsa_key(uint8_t* public_key,
        uint32_t public_key_len,
        uint32_t* public_key_actual_len,
        uint8_t* private_key,
        uint32_t private_key_len,
        uint32_t* private_key_actual_len);

#ifdef __cplusplus
}
#endif
