/*
 * Copyright 2020 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include <string.h>
#include "crypto.h"
#include "logging.h"
#include "error.h"
#include "types.h"

#ifdef __cplusplus
extern "C" {
#endif

extern "C" const unsigned int SYM_KEY_LEN = pdo::crypto::constants::SYM_KEY_LEN;
extern "C" const unsigned int RSA_PLAINTEXT_LEN = pdo::crypto::constants::RSA_PLAINTEXT_LEN;
extern "C" const unsigned int RSA_KEY_SIZE = pdo::crypto::constants::RSA_KEY_SIZE;

bool compute_hash(uint8_t* message,
    uint32_t message_len,
    uint8_t* hash,
    uint32_t max_hash_len,
    uint32_t* actual_hash_len)
{
    ByteArray ba;

    COND2ERR(message == NULL);

    ba = pdo::crypto::ComputeMessageHash(ByteArray(message, message + message_len));
    COND2ERR(ba.size() > max_hash_len);

    memcpy(hash, ba.data(), ba.size());
    *actual_hash_len = ba.size();
    return true;

err:
    return false;
}

bool verify_signature(uint8_t* public_key, uint32_t public_key_len, uint8_t* message, uint32_t message_len, uint8_t* signature, uint32_t signature_len)
{
    try
    {
        std::string pk_string((const char*)public_key, public_key_len);
        ByteArray msg(message, message + message_len);
        ByteArray sig(signature, signature + signature_len);

        //deserialize public key
        pdo::crypto::sig::PublicKey pk(pk_string);

        //check signature
        int r = pk.VerifySignature(msg, sig);
        COND2ERR(r != 1);
    }
    catch(...)
    {
        COND2ERR(true);
    }

    // verification successful
    return true;

err:
    return false;
}

bool pk_encrypt_message(uint8_t* public_key,
        uint32_t public_key_len,
        uint8_t* message,
        uint32_t message_len,
        uint8_t* encrypted_message,
        uint32_t encrypted_message_len,
        uint32_t* encrypted_message_actual_len)
{
    try
    {
        std::string pk_string((const char*)public_key, public_key_len);
        ByteArray msg(message, message + message_len);
        ByteArray encr_msg;

        //deserialize public key
        pdo::crypto::pkenc::PublicKey pk(pk_string);

        //encrypt message
        encr_msg = pk.EncryptMessage(msg);

        LOG_DEBUG("encr msg size %d buffer len %d",  encr_msg.size(), encrypted_message_len);
        COND2LOGERR(encrypted_message_len < encr_msg.size(), "buffer too small for encrypted msg");
        memcpy(encrypted_message, encr_msg.data(), encr_msg.size());
        *encrypted_message_actual_len = encr_msg.size();
    }
    catch(const std::exception& e)
    {
        COND2LOGERR(true, e.what());
    }

    // encryption successful
    return true;

err:
    return false;
}

bool decrypt_message(uint8_t* key,
        uint32_t key_len,
        uint8_t* encrypted_message,
        uint32_t encrypted_message_len,
        uint8_t* message,
        uint32_t message_len,
        uint32_t* message_actual_len)
{
    try
    {
        ByteArray ba_key(key, key + key_len);
        ByteArray encr_msg(encrypted_message, encrypted_message + encrypted_message_len);
        ByteArray msg;

        //decrypt message
        msg = pdo::crypto::skenc::DecryptMessage(ba_key, encr_msg);

        COND2ERR(message_len < msg.size());
        memcpy(message, msg.data(), msg.size());
        *message_actual_len = msg.size();
    }
    catch(const std::exception& e)
    {
        COND2LOGERR(true, e.what());
    }

    //decryption successful
    return true;

err:
    return false;
}

#ifdef __cplusplus
}
#endif /* __cplusplus */
