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
#include "utils.h"

#include "base64.h"

#include "sgx_utils.h"

extern sgx_ec256_private_t enclave_sk;
extern sgx_ec256_public_t enclave_pk;

// this is tlcc binding
sgx_ec256_public_t tlcc_pk = {0};

// state verification key; hardcoded for debugging
// note that key must be negociated during "binding phase" with the ledger enclave; for prototyping
// its is hardcoded at the moment!!!!
sgx_cmac_128bit_key_t session_key = {
    0x3F, 0xE2, 0x59, 0xDF, 0x62, 0x7F, 0xEF, 0x99, 0x5B, 0x4B, 0x00, 0xDE, 0x44, 0xC1, 0x26, 0x33};

// state encryption key; hardcoded for debugging
sgx_aes_gcm_128bit_key_t state_encryption_key = {
    0x6A, 0xB0, 0x46, 0xB3, 0x8D, 0x14, 0x2D, 0x17, 0x3F, 0x52, 0xF3, 0x9F, 0xDA, 0x1D, 0x63, 0x4A};

int ecall_bind_tlcc(const sgx_report_t* report, const uint8_t* pubkey)
{
    LOG_DEBUG("ecall_bind_tlcc: \tArgs: &report=%p, &pk=%p", report, pubkey);

    // IMPORTANT!!!
    // here is our testing backdoor for starting ecc without a tlcc instance
    if (report == NULL && pubkey == NULL)
    {
        LOG_WARNING("Start without TLCC!!!!");
        return SGX_SUCCESS;
    }

    sgx_sha256_hash_t pk_hash;
    sgx_sha256_msg(pubkey, 64, &pk_hash);
    std::string base64_hash = base64_encode((const unsigned char*)pk_hash, SGX_SHA256_HASH_SIZE);
    LOG_DEBUG("Received pk hash: %s", base64_hash.c_str());

    if (memcmp(&pk_hash, &(report->body.report_data), SGX_HASH_SIZE) != 0)
    {
        LOG_ERROR("PK does not match the one in report !");
        return SGX_ERROR_INVALID_PARAMETER;
    }

    int ver_ret = sgx_verify_report(report);
    if (ver_ret != SGX_SUCCESS)
    {
        LOG_ERROR("Attestation report verification failed!");
        return ver_ret;
    }

    memcpy(&tlcc_pk, pubkey, 64);

    // TODO negociate session key with tlcc
    // for prototyping this is hardcoded right now
    LOG_DEBUG("Binding successfull");
    return SGX_SUCCESS;
}

int gen_response(const char* txType,
    uint8_t* response,
    uint32_t* response_len_out,
    sgx_ec256_signature_t* signature,
    shim_ctx_ptr_t ctx)
{
    int ret;

    // Note: below signature is verified in
    // - ecc/crypto/ecdsa.go::Verify (for VSCC)
    // - tlcc_enclave/enclave/ledger.cpp::int parse_endorser_transaction (for TLCC)

    // create Hash <- H(txType in {"invoke"} || encoded_args || result || read set || write
    // set)
    // TODO: we should encode the hash below in an unambiguous fashion (which is not true with
    //    simple concatenation as done below!)
    //    Probably easiest by prefixing each field by length in fixed-size format?
    sgx_sha256_hash_t hash;
    sgx_sha_state_handle_t sha_handle;
    sgx_sha256_init(&sha_handle);
    LOG_DEBUG("txType: %s", txType);
    sgx_sha256_update((const uint8_t*)txType, strlen(txType), sha_handle);
    LOG_DEBUG("encoded_args: %s", ctx->encoded_args);
    sgx_sha256_update((const uint8_t*)ctx->encoded_args, strlen(ctx->encoded_args), sha_handle);
    LOG_DEBUG("response_data len: %d", *response_len_out);
    sgx_sha256_update(response, *response_len_out, sha_handle);

    // hash read and write set
    LOG_DEBUG("read_set:");
    for (auto& it : ctx->read_set)
    {
        LOG_DEBUG("\\-> %s", it.c_str());
        sgx_sha256_update((const uint8_t*)it.c_str(), it.size(), sha_handle);
    }

    LOG_DEBUG("write_set:");
    for (auto& it : ctx->write_set)
    {
        LOG_DEBUG("\\-> %s - %s", it.first.c_str(), it.second.c_str());
        sgx_sha256_update((const uint8_t*)it.first.c_str(), it.first.size(), sha_handle);
        sgx_sha256_update((const uint8_t*)it.second.c_str(), it.second.size(), sha_handle);
    }

    sgx_sha256_get_hash(sha_handle, &hash);
    sgx_sha256_close(sha_handle);

    // sig <- sign (hash,sk)
    uint8_t sig[sizeof(sgx_ec256_signature_t)];
    sgx_ecc_state_handle_t ecc_handle = NULL;
    sgx_ecc256_open_context(&ecc_handle);
    ret = sgx_ecdsa_sign((uint8_t*)&hash, SGX_SHA256_HASH_SIZE, &enclave_sk,
        (sgx_ec256_signature_t*)sig, ecc_handle);
    sgx_ecc256_close_context(ecc_handle);
    if (ret != SGX_SUCCESS)
    {
        LOG_ERROR("Signing failed!! Reason: %#08x", ret);
        return ret;
    }
    LOG_DEBUG("Response signature created!");
    // convert signature to big endian and copy out
    bytes_swap(sig, 32);
    bytes_swap(sig + 32, 32);
    memcpy(signature, sig, sizeof(sgx_ec256_signature_t));

    std::string base64_hash = base64_encode((const unsigned char*)hash, 32);
    LOG_DEBUG("ecc sig hash (base64): %s", base64_hash.c_str());

    std::string base64_sig =
        base64_encode((const unsigned char*)sig, sizeof(sgx_ec256_signature_t));
    LOG_DEBUG("ecc sig sig (base64): %s", base64_sig.c_str());

    std::string base64_pk =
        base64_encode((const unsigned char*)&enclave_pk, sizeof(sgx_ec256_public_t));
    LOG_DEBUG("ecc sig pk (base64): %s", base64_pk.c_str());

    return ret;
}

// chaincode call processing when we have secure channel ..
int invoke_enc(const char* pk,
    uint8_t* response,
    uint32_t max_response_len,
    uint32_t* actual_response_len,
    shim_ctx_ptr_t ctx)
{
    LOG_DEBUG("Encrypted invocation");
    sgx_ec256_public_t client_pk = {0};

    std::string _pk = base64_decode(pk);
    uint8_t* pk_bytes = (uint8_t*)_pk.c_str();
    bytes_swap(pk_bytes, 32);
    bytes_swap(pk_bytes + 32, 32);
    memcpy(&client_pk, pk_bytes, sizeof(sgx_ec256_public_t));

    sgx_ec256_dh_shared_t shared_dhkey;

    sgx_ecc_state_handle_t ecc_handle = NULL;
    sgx_ecc256_open_context(&ecc_handle);
    int sgx_ret =
        sgx_ecc256_compute_shared_dhkey(&enclave_sk, &client_pk, &shared_dhkey, ecc_handle);
    if (sgx_ret != SGX_SUCCESS)
    {
        LOG_ERROR("Compute shared dhkey: %d", sgx_ret);
        return sgx_ret;
    }
    sgx_ecc256_close_context(ecc_handle);
    bytes_swap(&shared_dhkey, 32);

    sgx_sha256_hash_t h;
    sgx_sha256_msg(
        (const uint8_t*)&shared_dhkey, sizeof(sgx_ec256_dh_shared_t), (sgx_sha256_hash_t*)&h);

    sgx_aes_gcm_128bit_key_t key;
    memcpy(key, h, sizeof(sgx_aes_gcm_128bit_key_t));

    std::string _cipher = base64_decode(ctx->encoded_args);
    uint8_t* cipher = (uint8_t*)_cipher.c_str();
    int cipher_len = _cipher.size();

    uint32_t needed_size = cipher_len - SGX_AESGCM_IV_SIZE - SGX_AESGCM_MAC_SIZE;
    // need one byte more for string terminator
    char plain[needed_size + 1];
    plain[needed_size] = '\0';

    // decrypt
    sgx_ret = sgx_rijndael128GCM_decrypt(&key,
        cipher + SGX_AESGCM_IV_SIZE + SGX_AESGCM_MAC_SIZE,         /* cipher */
        needed_size, (uint8_t*)plain,                              /* plain out */
        cipher, SGX_AESGCM_IV_SIZE,                                /* nonce */
        NULL, 0,                                                   /* aad */
        (sgx_aes_gcm_128bit_tag_t*)(cipher + SGX_AESGCM_IV_SIZE)); /* tag */
    if (sgx_ret != SGX_SUCCESS)
    {
        LOG_ERROR("Decrypt error: %x", sgx_ret);
        return sgx_ret;
    }
    LOG_DEBUG("invoke_enc: \tdecrypted args: %s", plain);
    ctx->json_args = plain;

    return invoke(response, max_response_len, actual_response_len, ctx);
}

#include "pdo/common/crypto/crypto.h"
bool crypto_created = false;

// chaincode call
// output, response <- F(args, input)
// signature <- sign (hash,sk)
int ecall_cc_invoke(const char* encoded_args,
    const char* pk,
    uint8_t* response,
    uint32_t response_len_in,
    uint32_t* response_len_out,
    sgx_ec256_signature_t* signature,
    void* u_shim_ctx)
{
    LOG_DEBUG("ecall_cc_invoke: \tArgs: %s", encoded_args);

    // NOTE/TODO: this crypto part will be removed
    if (!crypto_created)
    {
        pdo::crypto::sig::PublicKey verification_key_;
        pdo::crypto::sig::PrivateKey signature_key_;
        signature_key_.Generate();  // private key
        verification_key_ = signature_key_.GetPublicKey();
        // debug
        std::string s = verification_key_.Serialize();
        LOG_DEBUG("enclave verification key: %s", s.c_str());
        crypto_created = true;
    }
    else
    {
        LOG_DEBUG("enclave crypto material created");
    }

    t_shim_ctx_t ctx;
    ctx.u_shim_ctx = u_shim_ctx;
    ctx.encoded_args = encoded_args;

    // call chaincode invoke logic: creates output and response
    // output, response <- F(args, input)
    int ret;
    if (strlen(pk) == 0)
    {
        // clear input
        ctx.json_args = ctx.encoded_args;
        ret = invoke(response, response_len_in, response_len_out, &ctx);
    }
    else
    {
        // encrypted input
        ret = invoke_enc(pk, response, response_len_in, response_len_out, &ctx);
    }

    if (ret != 0)
    {
        return SGX_ERROR_UNEXPECTED;
    }

    ret = gen_response("invoke", response, response_len_out, signature, &ctx);

    return ret;
}
