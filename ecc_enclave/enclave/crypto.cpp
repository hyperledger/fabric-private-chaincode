/*
 * Copyright IBM Corp. All Rights Reserved.
 * Copyright 2020 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "crypto.h"
#include "error.h"
#include "logging.h"
#include "pdo/common/crypto/crypto.h"
#include "sgx_trts.h"

int get_random_bytes(uint8_t* buffer, size_t length)
{
    /* WARNING WARNING WARNING */
    /* WARNING WARNING WARNING */

    // the implementation of this function with SGX rand forces to have a single encalve endorser

    /* WARNING WARNING WARNING */
    /* WARNING WARNING WARNING */
    return sgx_read_rand(buffer, length);
}

bool validate_key_length(const ByteArray key)
{
    return key.size() == pdo::crypto::constants::SYM_KEY_LEN;
}

bool decrypt_message(const ByteArray key, const ByteArray& encrypted_message, ByteArray& message)
{
    bool b;
    COND2LOGERR(!validate_key_length(key), "invalid decryption key length");
    CATCH(b, message = pdo::crypto::skenc::DecryptMessage(key, encrypted_message));
    COND2LOGERR(!b, "message decryption failed");

    return true;

err:
    return false;
}

bool encrypt_message(const ByteArray key, const ByteArray& message, ByteArray& encrypted_message)
{
    bool b;
    COND2LOGERR(!validate_key_length(key), "invalid encryption key length");
    CATCH(b, encrypted_message = pdo::crypto::skenc::EncryptMessage(key, message));
    COND2LOGERR(!b, "message encryption failed");

    return true;

err:
    return false;
}

bool compute_message_hash(const ByteArray message, ByteArray& message_hash)
{
    bool b;
    CATCH(b, message_hash = pdo::crypto::ComputeMessageHash(message););
    COND2LOGERR(!b, "computing message hash failed");

    return true;

err:
    return false;
}
