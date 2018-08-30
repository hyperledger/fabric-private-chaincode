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

#include "enclave_t.h"

#include "chaincode.h"
#include "logging.h"
#include "shim.h"
#include "utils.h"

#include "base64.h"

#include "sgx_utils.h"

extern sgx_ec256_private_t enclave_sk;
extern sgx_ec256_public_t enclave_pk;

// this is tlcc binding
sgx_ec256_public_t tlcc_pk = {0};

// state verification key; hardcoded for debugging
// note that key must be negociated during "binding phase" with the ledger enclave; for prototyping its is hardcoded at the moment!!!!
sgx_cmac_128bit_key_t session_key = {
    0x3F, 0xE2, 0x59, 0xDF, 0x62, 0x7F, 0xEF, 0x99, 0x5B, 0x4B, 0x00, 0xDE, 0x44, 0xC1, 0x26, 0x33};

// state encryption key; hardcoded for debugging
sgx_aes_gcm_128bit_key_t state_encryption_key = {
    0x6A, 0xB0, 0x46, 0xB3, 0x8D, 0x14, 0x2D, 0x17, 0x3F, 0x52, 0xF3, 0x9F, 0xDA, 0x1D, 0x63, 0x4A};

int ecall_bind_tlcc(const sgx_report_t *report, const uint8_t *pubkey)
{
    // IMPORTANT!!!
    // here is out testing backdoor for starting ecc without a tlcc instance
    if (report == NULL && pubkey == NULL) {
        LOG_DEBUG("Start without TLCC!!!!");
        return SGX_SUCCESS;
    }

    sgx_sha256_hash_t pk_hash;
    sgx_sha256_msg(pubkey, 64, &pk_hash);
    std::string base64_hash = base64_encode((const unsigned char *)pk_hash, SGX_SHA256_HASH_SIZE);
    LOG_DEBUG("Received pk hash: %s", base64_hash.c_str());

    if (memcmp(&pk_hash, &(report->body.report_data), SGX_HASH_SIZE) != 0) {
        LOG_ERROR("PK does not match the one in report !");
        return SGX_ERROR_INVALID_PARAMETER;
    }

    int ver_ret = sgx_verify_report(report);
    if (ver_ret != SGX_SUCCESS) {
        LOG_ERROR("Attestation report verification failed!");
        return ver_ret;
    }

    memcpy(&tlcc_pk, pubkey, 64);

    // TODO negociate session key with tlcc
    // for prototyping this is hardcoded right now
    LOG_DEBUG("Binding successfull");
    return SGX_SUCCESS;
}

int invoke_enc(const char *args, const char *pk, uint8_t *response, uint32_t max_response_len,
    uint32_t *actual_response_len, void *ctx)
{
    LOG_DEBUG("Encrypted invcation");
    sgx_ec256_public_t client_pk = {0};

    std::string _pk = base64_decode(pk);
    uint8_t *pk_bytes = (uint8_t *)_pk.c_str();
    bytes_swap(pk_bytes, 32);
    bytes_swap(pk_bytes + 32, 32);
    memcpy(&client_pk, pk_bytes, sizeof(sgx_ec256_public_t));

    sgx_ec256_dh_shared_t shared_dhkey;

    sgx_ecc_state_handle_t ecc_handle = NULL;
    sgx_ecc256_open_context(&ecc_handle);
    int sgx_ret =
        sgx_ecc256_compute_shared_dhkey(&enclave_sk, &client_pk, &shared_dhkey, ecc_handle);
    if (sgx_ret != SGX_SUCCESS) {
        LOG_ERROR("Compute shared dhkey: %d\n", sgx_ret);
        return sgx_ret;
    }
    sgx_ecc256_close_context(ecc_handle);
    bytes_swap(&shared_dhkey, 32);

    sgx_sha256_hash_t h;
    sgx_sha256_msg(
        (const uint8_t *)&shared_dhkey, sizeof(sgx_ec256_dh_shared_t), (sgx_sha256_hash_t *)&h);

    sgx_aes_gcm_128bit_key_t key;
    memcpy(key, h, sizeof(sgx_aes_gcm_128bit_key_t));

    std::string _cipher = base64_decode(args);
    uint8_t *cipher = (uint8_t *)_cipher.c_str();
    int cipher_len = _cipher.size();

    uint32_t needed_size = cipher_len - SGX_AESGCM_IV_SIZE - SGX_AESGCM_MAC_SIZE;
    // need one byte more for string terminator
    char plain[needed_size + 1];
    plain[needed_size] = '\0';

    // decrypt
    sgx_ret = sgx_rijndael128GCM_decrypt(&key,
        cipher + SGX_AESGCM_IV_SIZE + SGX_AESGCM_MAC_SIZE,          /* cipher */
        needed_size, (uint8_t *)plain,                              /* plain out */
        cipher, SGX_AESGCM_IV_SIZE,                                 /* nonce */
        NULL, 0,                                                    /* aad */
        (sgx_aes_gcm_128bit_tag_t *)(cipher + SGX_AESGCM_IV_SIZE)); /* tag */
    if (sgx_ret != SGX_SUCCESS) {
        LOG_ERROR("Decrypt error: %x\n", sgx_ret);
        return sgx_ret;
    }

    return invoke((const char *)plain, response, max_response_len, actual_response_len, ctx);
}

// chaincode call
// output, response <- F(args, input)
// signature <- sign (hash,sk)
int ecall_invoke(const char *args, const char *pk, uint8_t *response, uint32_t response_len_in,
    uint32_t *response_len_out, sgx_ec256_signature_t *signature, void *ctx)
{
    // register ctx
    read_set_t readset;
    write_set_t writeset;

    register_rwset(ctx, &readset, &writeset);

    // call chaincode invoke logic: creates output and response
    // output, response <- F(args, input)
    int ret;
    if (strlen(pk) == 0) {
        // clear input
        ret = invoke(args, response, response_len_in, response_len_out, ctx);
    } else {
        // encrypted input
        ret = invoke_enc(args, pk, response, response_len_in, response_len_out, ctx);
    }

    if (ret != 0) {
        return SGX_ERROR_UNEXPECTED;
    }

    // create Hash <- H(args || result || read-write set)
    sgx_sha256_hash_t hash;
    sgx_sha_state_handle_t sha_handle;
    sgx_sha256_init(&sha_handle);
    sgx_sha256_update((const uint8_t *)args, strlen(args), sha_handle);
    sgx_sha256_update(response, *response_len_out, sha_handle);

    // hash read and write set
    LOG_DEBUG("read_set:");
    for (auto &it : readset) {
        LOG_DEBUG("\\-> %s", it.c_str());
        sgx_sha256_update((const uint8_t *)it.c_str(), it.size(), sha_handle);
    }

    LOG_DEBUG("write_set:");
    for (auto &it : writeset) {
        LOG_DEBUG("\\-> %s - %s", it.first.c_str(), it.second.c_str());
        sgx_sha256_update((const uint8_t *)it.first.c_str(), it.first.size(), sha_handle);
        sgx_sha256_update((const uint8_t *)it.second.c_str(), it.second.size(), sha_handle);
    }

    sgx_sha256_get_hash(sha_handle, &hash);
    sgx_sha256_close(sha_handle);

    // clean context
    free_rwset(ctx);

    // sig <- sign (hash,sk)
    uint8_t sig[sizeof(sgx_ec256_signature_t)];
    sgx_ecc_state_handle_t ecc_handle = NULL;
    sgx_ecc256_open_context(&ecc_handle);
    ret = sgx_ecdsa_sign((uint8_t *)&hash, SGX_SHA256_HASH_SIZE, &enclave_sk,
        (sgx_ec256_signature_t *)sig, ecc_handle);
    sgx_ecc256_close_context(ecc_handle);
    if (ret != SGX_SUCCESS) {
        LOG_ERROR("Signing failed!! Reason: %#08x\n", ret);
        return ret;
    }
    LOG_DEBUG("Response signature created!");
    // convert signature to big endian and copy out
    bytes_swap(sig, 32);
    bytes_swap(sig + 32, 32);
    memcpy(signature, sig, sizeof(sgx_ec256_signature_t));

    std::string base64_hash = base64_encode((const unsigned char *)hash, 32);
    LOG_DEBUG("ecc sig hash: %s", base64_hash.c_str());

    std::string base64_sig =
        base64_encode((const unsigned char *)sig, sizeof(sgx_ec256_signature_t));
    LOG_DEBUG("ecc sig sig: %s", base64_sig.c_str());

    std::string base64_pk =
        base64_encode((const unsigned char *)&enclave_pk, sizeof(sgx_ec256_public_t));
    LOG_DEBUG("ecc sig pk: %s", base64_pk.c_str());

    return ret;
}
