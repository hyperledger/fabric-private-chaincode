/*
 * Copyright IBM Corp. All Rights Reserved.
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

static sgx_thread_mutex_t global_mutex = SGX_THREAD_MUTEX_INITIALIZER;

extern sgx_ec256_public_t tlcc_pk;
extern sgx_cmac_128bit_key_t session_key;
extern sgx_aes_gcm_128bit_key_t state_encryption_key;

void get_creator_name(
    char* msp_id, uint32_t max_msp_id_len, char* dn, uint32_t max_dn_len, shim_ctx_ptr_t ctx)
{
    // TODO: right now the implementation is not secure yet as below function is unvalidated
    // from the (untrusted) peer.
    // To securely implement it, we will require the signed proposal to be passed
    // from the stub (see, e.g., ChaincodeStub in go shim core/chaincode/shim/stub.go)
    // and then verified. This in turn will require verification of certificates based
    // on the MSP info channel.  As TLCC already has to do keep track of MSP and do related
    // verification , we can off-load some of that to TLCC (as we anyway have to talk to it
    // to get channel MSP info)
    ocall_get_creator_name(msp_id, max_msp_id_len, dn, max_dn_len, ctx->u_shim_ctx);
}

void get_state(
    const char* key, uint8_t* val, uint32_t max_val_len, uint32_t* val_len, shim_ctx_ptr_t ctx)
{
    // read state
    ctx->read_set.insert(std::string(key));

    sgx_cmac_128bit_tag_t cmac = {0};

    uint8_t encoded_cipher[(max_val_len + SGX_AESGCM_IV_SIZE + SGX_AESGCM_MAC_SIZE + 2) / 3 * 4];
    uint32_t encoded_cipher_len;
    ocall_get_state(key, encoded_cipher, sizeof(encoded_cipher), &encoded_cipher_len,
        (sgx_cmac_128bit_tag_t*)cmac, ctx->u_shim_ctx);

    LOG_DEBUG("Enclave: got enc-cipher for key=%s len=%d val='%s'", key, encoded_cipher_len,
        (encoded_cipher_len > 0 ? (const char*)encoded_cipher : ""));

    // create state hash
    sgx_sha256_hash_t state_hash = {0};
    if (encoded_cipher_len > 0)
    {
        sgx_sha256_msg(encoded_cipher, encoded_cipher_len, &state_hash);
    }

    if (check_cmac(key, NULL, &state_hash, &session_key, &cmac) != 0)
    {
        LOG_ERROR("Enclave: VIOLATION!!! Oh oh! cmac does not match!");
        // TODO: proper error handling/abort rather than just continue
    }
    else
    {
        LOG_DEBUG("Enclave: State verification: cmac correct!! :D");
    }
    // if nothing read, no need for decryption
    if (encoded_cipher_len == 0)
    {
        *val_len = 0;
        return;
    }
    // base64 decode
    std::string cipher = base64_decode((const char*)encoded_cipher);
    if (cipher.size() < SGX_AESGCM_IV_SIZE + SGX_AESGCM_MAC_SIZE)
    {
        LOG_ERROR(
            "Enclave: base64 decoding failed/produced too shart a value with %d", cipher.size());
        *val_len = 0;
        return;
    }

    // decrypt
    *val_len = cipher.size() - SGX_AESGCM_IV_SIZE - SGX_AESGCM_MAC_SIZE;
    int ret = decrypt_state(
        &state_encryption_key, (uint8_t*)cipher.c_str(), cipher.size(), val, *val_len);
    if (ret != SGX_SUCCESS)
    {
        LOG_ERROR("Enclave: Error decrypting state: %d", ret);
        *val_len = 0;
    }
}

void put_state(const char* key, uint8_t* val, uint32_t val_len, shim_ctx_ptr_t ctx)
{
    // encrypt
    uint32_t cipher_len = val_len + SGX_AESGCM_IV_SIZE + SGX_AESGCM_MAC_SIZE;
    uint8_t cipher[cipher_len];
    int ret = encrypt_state(&state_encryption_key, val, val_len, cipher, cipher_len);
    if (ret != SGX_SUCCESS)
    {
        LOG_ERROR("Enclave: Error encrypting state");
    }

    // base64 encode
    std::string base64 = base64_encode((unsigned char*)cipher, cipher_len);

    // write state
    ctx->write_set.insert({key, base64});
    ocall_put_state(key, (uint8_t*)base64.c_str(), base64.size(), ctx->u_shim_ctx);
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
    for (int i = 0; i < json_array_get_count(pairs); i++)
    {
        JSON_Object* pair = json_array_get_object(pairs, i);
        const char* key = json_object_get_string(pair, "key");
        const char* value = json_object_get_string(pair, "value");
        values.insert({key, value});
    }
    json_value_free(root);
    return 1;
}

void get_state_by_partial_composite_key(
    const char* comp_key, std::map<std::string, std::string>& values, shim_ctx_ptr_t ctx)
{
    uint8_t json[262144];  // 128k needed for 1000 bids
    uint32_t len;

    sgx_cmac_128bit_tag_t cmac = {0};
    ocall_get_state_by_partial_composite_key(
        comp_key, json, sizeof(json), &len, (sgx_cmac_128bit_tag_t*)cmac, ctx->u_shim_ctx);

    unmarshal_values(values, (const char*)json, len);

    // create state hash
    sgx_sha256_hash_t state_hash = {0};
    sgx_sha_state_handle_t sha_handle;
    sgx_sha256_init(&sha_handle);

    for (auto& u : values)
    {
        ctx->read_set.insert(u.first);

        // but also compute hash
        sgx_sha256_update((const uint8_t*)u.first.c_str(), u.first.size(), sha_handle);
        sgx_sha256_update((const uint8_t*)u.second.c_str(), u.second.size(), sha_handle);

        // base64 decode
        std::string cipher = base64_decode(u.second.c_str());

        // decrypt
        uint32_t plain_len = cipher.size() - SGX_AESGCM_IV_SIZE - SGX_AESGCM_MAC_SIZE;
        uint8_t plain[plain_len];
        int ret = decrypt_state(
            &state_encryption_key, (uint8_t*)cipher.c_str(), cipher.size(), plain, plain_len);
        if (ret != SGX_SUCCESS)
        {
            LOG_ERROR("Enclave: Error decrypting state: %d", ret);
        }

        std::string s((const char*)plain, plain_len);
        u.second = s;
    }

    sgx_sha256_get_hash(sha_handle, &state_hash);
    sgx_sha256_close(sha_handle);

    if (check_cmac(comp_key, NULL, &state_hash, &session_key, &cmac) != 0)
    {
        LOG_ERROR("Enclave: VIOLATION!!! Oh oh! cmac does not match!");
        // TODO: proper error handling/abort rather than just continue
    }
    else
    {
        LOG_DEBUG("Enclave: State verification: cmac correct!! :D");
    }
}

int get_string_args(std::vector<std::string>& argss, shim_ctx_ptr_t ctx)
{
    JSON_Value* root = json_parse_string(ctx->json_args);
    if (json_value_get_type(root) != JSONArray)
    {
        LOG_ERROR("Shim: Cannot parse args '%s'", ctx->json_args);
        return -1;
    }

    JSON_Array* args = json_value_get_array(root);
    for (int i = 0; i < json_array_get_count(args); i++)
    {
        argss.push_back(json_array_get_string(args, i));
    }
    json_value_free(root);
    return 1;
}

int get_func_and_params(
    std::string& func_name, std::vector<std::string>& params, shim_ctx_ptr_t ctx)
{
    JSON_Value* root = json_parse_string(ctx->json_args);
    if (json_value_get_type(root) != JSONArray)
    {
        LOG_ERROR("Shim: Cannot parse args '%s'", ctx->json_args);
        return -1;
    }

    JSON_Array* args = json_value_get_array(root);
    if (0 < json_array_get_count(args))
    {
        func_name = json_array_get_string(args, 0);
    }
    else
    {
        LOG_ERROR("Shim: args '%s' do not contain a function name", ctx->json_args);
        return -1;
    }

    for (int i = 1; i < json_array_get_count(args); i++)
    {
        params.push_back(json_array_get_string(args, i));
    }
    json_value_free(root);
    return 1;
}
