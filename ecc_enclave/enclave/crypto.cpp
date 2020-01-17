/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "crypto.h"

#include <assert.h>
#include <string.h>  // for memcpy etc

#include "sgx_trts.h"

int check_cmac(const char* key,
    uint8_t* nonce,
    sgx_sha256_hash_t* state_hash,
    sgx_cmac_128bit_key_t* cmac_key,
    sgx_cmac_128bit_tag_t* cmac)
{
    // hash( key || nonce || target_hash || result )
    sgx_cmac_128bit_tag_t tmp_cmac = {0};
    sgx_cmac_state_handle_t cmac_handle;
    sgx_cmac128_init(cmac_key, &cmac_handle);
    sgx_cmac128_update((const uint8_t*)key, strlen(key), cmac_handle);
    /* sgx_cmac128_update(nonce, 32, cmac_handle); */
    sgx_cmac128_update((const uint8_t*)state_hash, sizeof(sgx_sha256_hash_t), cmac_handle);
    sgx_cmac128_final(cmac_handle, &tmp_cmac);
    sgx_cmac128_close(cmac_handle);

    if (memcmp(&tmp_cmac, cmac, sizeof(sgx_cmac_128bit_tag_t)) != 0)
    {
        return -1;
    }
    return 0;
}

int encrypt_state(sgx_aes_gcm_128bit_key_t* key,
    uint8_t* plain,
    uint32_t plain_len,
    uint8_t* cipher,
    uint32_t cipher_len)
{
    // create buffer
    uint32_t needed_size = plain_len + SGX_AESGCM_IV_SIZE + SGX_AESGCM_MAC_SIZE;
    assert(cipher_len >= needed_size);

    // gen rnd iv
    sgx_read_rand(cipher, SGX_AESGCM_IV_SIZE);

    // encrypt
    return sgx_rijndael128GCM_encrypt(key, plain, plain_len,
        cipher + SGX_AESGCM_IV_SIZE + SGX_AESGCM_MAC_SIZE, cipher, SGX_AESGCM_IV_SIZE, NULL, 0,
        (sgx_aes_gcm_128bit_tag_t*)(cipher + SGX_AESGCM_IV_SIZE));
}

int decrypt_state(sgx_aes_gcm_128bit_key_t* key,
    uint8_t* cipher,
    uint32_t cipher_len,
    uint8_t* plain,
    uint32_t plain_len)
{
    // create buffer
    uint32_t needed_size = cipher_len - SGX_AESGCM_IV_SIZE - SGX_AESGCM_MAC_SIZE;
    assert(plain_len >= needed_size);

    // decrypt
    return sgx_rijndael128GCM_decrypt(key,
        cipher + SGX_AESGCM_IV_SIZE + SGX_AESGCM_MAC_SIZE,         /* cipher */
        plain_len, plain,                                          /* plain out */
        cipher, SGX_AESGCM_IV_SIZE,                                /* nonce */
        NULL, 0,                                                   /* aad */
        (sgx_aes_gcm_128bit_tag_t*)(cipher + SGX_AESGCM_IV_SIZE)); /* tag */
}

int get_random_bytes(uint8_t* buffer, size_t length)
{
    /* WARNING WARNING WARNING */
    /* WARNING WARNING WARNING */

    // the implementation of this function with SGX rand forces to have a single encalve endorser

    /* WARNING WARNING WARNING */
    /* WARNING WARNING WARNING */
    return sgx_read_rand(buffer, length);
}
