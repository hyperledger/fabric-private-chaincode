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
#include "error.h"

#include <mbusafecrt.h> /* for memcpy_s etc */
#include "sgx_utils.h"

#include "fpc/fpc.pb.h"
#include "pb_decode.h"
#include "pb_encode.h"

extern sgx_ec256_private_t enclave_sk;
extern sgx_ec256_public_t enclave_pk;

// this is tlcc binding
sgx_ec256_public_t tlcc_pk = {0};

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
/*
int gen_response(const char* txType,
    uint8_t* response,
    uint32_t* response_len_out,
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
    //memcpy(signature, sig, sizeof(sgx_ec256_signature_t));

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
*/
/*
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
        cipher + SGX_AESGCM_IV_SIZE + SGX_AESGCM_MAC_SIZE,         // cipher
        needed_size, (uint8_t*)plain,                              // plain out
        cipher, SGX_AESGCM_IV_SIZE,                                // nonce
        NULL, 0,                                                   // aad
        (sgx_aes_gcm_128bit_tag_t*)(cipher + SGX_AESGCM_IV_SIZE)); // tag
    if (sgx_ret != SGX_SUCCESS)
    {
        LOG_ERROR("Decrypt error: %x", sgx_ret);
        return sgx_ret;
    }
    LOG_DEBUG("invoke_enc: \tdecrypted args: %s", plain);
    ctx->json_args = plain;

    return invoke(response, max_response_len, actual_response_len, ctx);
}
*/

#include "pdo/common/crypto/crypto.h"
bool crypto_created = false;

// chaincode call
// output, response <- F(args, input)
// signature <- sign (hash,sk)
int ecall_cc_invoke(const uint8_t* signed_proposal_proto_bytes,
    uint32_t signed_proposal_proto_bytes_len,
    const uint8_t* cc_request_message_proto,
    uint32_t cc_request_message_proto_len,
    uint8_t* cc_response_message_proto,
    uint32_t cc_response_message_proto_len_in,
    uint32_t* cc_response_message_proto_len_out,
    void* u_shim_ctx)
{
    LOG_DEBUG("ecall_cc_invoke");
    LOG_DEBUG("signed proposal length %u", signed_proposal_proto_bytes_len);

    bool b;
    fpc_ChaincodeRequestMessage cc_request_message = {};
    fpc_CleartextChaincodeRequest cleartext_cc_request = {};
    t_shim_ctx_t ctx;
    int ret;
    // estimate max response len (take into account other fields and b64 encoding)
    uint32_t response_len = cc_response_message_proto_len_in / 4 * 3 - 1024;
    uint8_t response[cc_response_message_proto_len_in / 4 * 3];
    uint32_t response_len_out = 0;
    std::string b64_response;

    ctx.u_shim_ctx = u_shim_ctx;

    {
        pb_istream_t istream;

        // set stream for ChaincodeRequestMessage
        istream = pb_istream_from_buffer(
            (const unsigned char*)cc_request_message_proto, cc_request_message_proto_len);

        b = pb_decode(&istream, fpc_ChaincodeRequestMessage_fields, &cc_request_message);
        COND2LOGERR(!b, PB_GET_ERROR(&istream));
        COND2LOGERR(cc_request_message.encrypted_request->size == 0, "zero size request");

        // set stream for CleartextChaincodeRequestMessage
        istream = pb_istream_from_buffer(
            (const unsigned char*)cc_request_message.encrypted_request->bytes,
            cc_request_message.encrypted_request->size);
        b = pb_decode(&istream, fpc_CleartextChaincodeRequest_fields, &cleartext_cc_request);
        COND2LOGERR(!b, PB_GET_ERROR(&istream));
        COND2LOGERR(!cleartext_cc_request.has_input, "no input in cleartext request");

        for (int i = 0; i < cleartext_cc_request.input.args_count; i++)
        {
            ctx.string_args.push_back(
                std::string((const char*)cleartext_cc_request.input.args[i]->bytes,
                    cleartext_cc_request.input.args[i]->size));
        }

        // the dynamic memory in the message is release at the end
    }

    ret = invoke(response, response_len, &response_len_out, &ctx);
    COND2ERR(ret != 0);

    b64_response = base64_encode((const unsigned char*)response, response_len_out);

    {
        // TODO put response in protobuf and encode it

        fpc_ChaincodeResponseMessage crm;
        pb_ostream_t ostream;
        std::string b64_crm_proto;

        // create proto struct to encode
        crm = {};
        crm.encrypted_response = (pb_bytes_array_t*)pb_realloc(
            crm.encrypted_response, PB_BYTES_ARRAY_T_ALLOCSIZE(b64_response.length()));
        COND2LOGERR(crm.encrypted_response == NULL, "cannot allocate encrypted message");
        crm.encrypted_response->size = b64_response.length();
        ret = memcpy_s(crm.encrypted_response->bytes, crm.encrypted_response->size,
            b64_response.c_str(), b64_response.length());
        COND2LOGERR(ret != 0, "cannot encode field");

        // encode proto
        ostream =
            pb_ostream_from_buffer(cc_response_message_proto, cc_response_message_proto_len_in);
        b = pb_encode(&ostream, fpc_ChaincodeResponseMessage_fields, &crm);
        COND2LOGERR(!b, "error encoding proto");

        pb_release(fpc_ChaincodeResponseMessage_fields, &crm);

        *cc_response_message_proto_len_out = ostream.bytes_written;
    }

    // release dynamic allocations (TODO:release in case of error)
    pb_release(fpc_ChaincodeRequestMessage_fields, &cc_request_message);

    return 0;

err:
    *cc_response_message_proto_len_out = 0;
    return SGX_ERROR_UNEXPECTED;
}
