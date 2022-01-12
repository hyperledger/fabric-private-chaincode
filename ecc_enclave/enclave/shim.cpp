/*
 * Copyright IBM Corp. All Rights Reserved.
 * Copyright 2020 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "enclave_t.h"

#include "logging.h"
#include "shim.h"
#include "shim_internals.h"

#include "crypto.h"

#include "base64.h"
#include "parson.h"

#include "sgx_thread.h"

#include <mbusafecrt.h> /* for memcpy_s etc */
#include "cc_data.h"
#include "error.h"

static sgx_thread_mutex_t global_mutex = SGX_THREAD_MUTEX_INITIALIZER;

void get_creator_name(
    char* msp_id, uint32_t max_msp_id_len, char* dn, uint32_t max_dn_len, shim_ctx_ptr_t ctx)
{
    strcpy_s(msp_id, max_msp_id_len, ctx->creator_msp_id.c_str());
    strcpy_s(dn, max_dn_len, ctx->creator_name.c_str());

    return;
}

void get_state(
    const char* key, uint8_t* val, uint32_t max_val_len, uint32_t* val_len, shim_ctx_ptr_t ctx)
{
    // estimate max encrypted val length
    uint32_t max_encrypted_val_len = cc_data::estimate_encrypted_state_value_length(max_val_len);
    uint8_t encoded_cipher[max_encrypted_val_len];
    uint32_t encoded_cipher_len = 0;
    std::string encoded_cipher_s;
    std::string cipher;

    get_public_state(key, encoded_cipher, sizeof(encoded_cipher), &encoded_cipher_len, ctx);

    // if nothing read, no need for decryption
    if (encoded_cipher_len == 0)
    {
        *val_len = 0;
        return;
    }

    // if got value size larger than input array, report error
    COND2LOGERR(encoded_cipher_len > sizeof(encoded_cipher),
        "encoded_cipher_len greater than buffer length");

    // build the encoded cipher string
    encoded_cipher_s = std::string((const char*)encoded_cipher, encoded_cipher_len);
    COND2LOGERR(encoded_cipher_len != encoded_cipher_s.size(), "Unexpected string length");

    // base64 decode
    // TODO: double check if Fabric handles values as strings; if it handles binary values, remove
    // encoding
    cipher = base64_decode(encoded_cipher_s);

    // decrypt
    try
    {
        ByteArray value;
        ByteArray encrypted_value = ByteArray(cipher.c_str(), cipher.c_str() + cipher.size());
        g_cc_data->decrypt_state_value(encrypted_value, value);
        COND2ERR(memcpy_s(val, max_val_len, value.data(), value.size()) != 0);
        *val_len = value.size();
    }
    catch (...)
    {
        COND2LOGERR(true, "Error decrypting state");
    }

    return;

err:
    *val_len = 0;
}

void get_public_state(
    const char* key, uint8_t* val, uint32_t max_val_len, uint32_t* val_len, shim_ctx_ptr_t ctx)
{
    // TODO error checking and remove exceptions

    // read state
    ocall_get_state(key, val, max_val_len, val_len, ctx->u_shim_ctx);
    if (*val_len > max_val_len)
    {
        char s[] = "Enclave: val_len greater than max_val_len";
        LOG_ERROR("%s", s);
        throw std::runtime_error(s);
    }

    LOG_DEBUG("Enclave: got state for key=%s len=%d val='%s'", key, *val_len,
        (*val_len > 0 ? (std::string((const char*)val, *val_len)).c_str() : ""));

    // save in rw set: only save the first key/value-hash pair
    std::string key_s(key);
    ByteArray value_hash;
    compute_message_hash(ByteArray(val, val + *val_len), value_hash);
    auto it = ctx->read_set.find(key_s);
    if (it == ctx->read_set.end())
    {
        // key not found, insert it
        ctx->read_set.insert({std::string(key), value_hash});
    }
    else
    {
        // key found, ensure value is the same, or fail
        if (it->second != value_hash)
        {
            char s[] = "value read inconsistency";
            LOG_ERROR("%s", s);
            throw std::runtime_error(s);
        }

        // IMPORTANT:
        // * the read set stores the first key that is read;
        // * subsequent reads of this key must return the same value, **no matter any local
        // update/write**
        // * hence, if a read key is updated later, the get_state must return the original
        // (presumably committed) value
        // TODO: double check consistency levels with fabric maintainers
    }
}

void put_state(const char* key, uint8_t* val, uint32_t val_len, shim_ctx_ptr_t ctx)
{
    // TODO error checking and remove exceptions
    if (val_len == 0)
    {
        char s[] = "put state cannot accept zero length values";
        LOG_ERROR("%s", s);
        throw std::runtime_error(s);
    }

    ByteArray encrypted_val;

    // encrypt
    try
    {
        ByteArray value(val, val + val_len);
        g_cc_data->encrypt_state_value(value, encrypted_val);
    }
    catch (...)
    {
        LOG_ERROR("Enclave: Error encrypting state");
        return;
    }

    // base64 encode
    std::string base64 = base64_encode((unsigned char*)encrypted_val.data(), encrypted_val.size());

    // write state
    put_public_state(key, (uint8_t*)base64.c_str(), base64.size(), ctx);
}

void put_public_state(const char* key, uint8_t* val, uint32_t val_len, shim_ctx_ptr_t ctx)
{
    // TODO error checking, ensure write gets to fabric

    // save write -- only the last one for the same key
    ctx->write_set.erase(key);
    ctx->write_set.insert({key, ByteArray(val, val + val_len)});
    ctx->del_set.erase(key);

    ocall_put_state(key, val, val_len, ctx->u_shim_ctx);
}

int unmarshal_values(
    std::map<std::string, std::string>& values, const char* json_bytes, uint32_t json_len)
{
    JSON_Value* root = json_parse_string(json_bytes);
    if (json_value_get_type(root) != JSONArray)
    {
        LOG_ERROR("Shim: Cannot parse values");
        return -1;
    }

    JSON_Array* pairs = json_value_get_array(root);
    COND2ERR(pairs == NULL);

    for (int i = 0; i < json_array_get_count(pairs); i++)
    {
        JSON_Object* pair = json_array_get_object(pairs, i);
        const char* key = json_object_get_string(pair, "key");
        const char* value = json_object_get_string(pair, "value");
        values.insert({key, value});
    }
    json_value_free(root);
    return 1;

err:
    return -1;
}

void get_state_by_partial_composite_key(
    const char* comp_key, std::map<std::string, std::string>& values, shim_ctx_ptr_t ctx)
{
    get_public_state_by_partial_composite_key(comp_key, values, ctx);
    for (auto& u : values)
    {
        // base64 decode
        std::string cipher = base64_decode(u.second.c_str());

        // decrypt
        try
        {
            ByteArray value;
            g_cc_data->decrypt_state_value(
                ByteArray(cipher.c_str(), cipher.c_str() + cipher.size()), value);
            std::string s((const char*)value.data(), value.size());
            u.second = s;
        }
        catch (...)
        {
            COND2LOGERR(true, "Error decrypting state");
        }
    }

    return;

err:
    // delete all values
    for (auto& u : values)
        u.second.clear();
}

void get_public_state_by_partial_composite_key(
    const char* comp_key, std::map<std::string, std::string>& values, shim_ctx_ptr_t ctx)
{
    uint8_t json[262144];  // 128k needed for 1000 bids
    uint32_t len = 0;

    ocall_get_state_by_partial_composite_key(comp_key, json, sizeof(json), &len, ctx->u_shim_ctx);
    if (len > sizeof(json))
    {
        char s[] = "Enclave: len greater than json buffer size";
        LOG_ERROR("%s", s);
        throw std::runtime_error(s);
    }

    unmarshal_values(values, (const char*)json, len);
}

void del_state(const char* key, shim_ctx_ptr_t ctx)
{
    ctx->write_set.erase(key);
    ctx->del_set.insert(key);

    ocall_del_state(key, ctx->u_shim_ctx);
    return;
}

int get_string_args(std::vector<std::string>& argss, shim_ctx_ptr_t ctx)
{
    argss = ctx->string_args;
    return 1;
}

int get_func_and_params(
    std::string& func_name, std::vector<std::string>& params, shim_ctx_ptr_t ctx)
{
    COND2LOGERR(ctx->string_args.size() == 0, "no function name");
    func_name = ctx->string_args[0];
    params = ctx->string_args;
    params.erase(params.begin());

    return 1;

err:
    return -1;
}

void get_signed_proposal(ByteArray& signed_proposal, shim_ctx_ptr_t ctx)
{
    signed_proposal = ctx->signed_proposal;
}

void get_channel_id(std::string& channel_id, shim_ctx_ptr_t ctx)
{
    channel_id = ctx->channel_id;
}

void get_tx_id(std::string& tx_id, shim_ctx_ptr_t ctx)
{
    tx_id = ctx->tx_id;
}

void get_creator(ByteArray& creator, shim_ctx_ptr_t ctx)
{
    creator = ctx->creator;
}
