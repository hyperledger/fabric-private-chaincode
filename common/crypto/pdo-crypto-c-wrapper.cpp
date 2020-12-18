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

#ifdef __cplusplus
}
#endif /* __cplusplus */
