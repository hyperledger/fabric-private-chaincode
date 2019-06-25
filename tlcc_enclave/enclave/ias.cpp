/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "ias.h"

#include "base64.h"
#include "logging.h"
#include "parson.h"

#include <string.h>

// openssl
#include <openssl/bio.h>
#include <openssl/ecdsa.h>
#include <openssl/err.h>
#include <openssl/pem.h>
#include <openssl/x509.h>
#include "openssl/sha.h"

static const char* INTEL_PUB_PEM =
    "-----BEGIN PUBLIC KEY-----\n\
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAqXot4OZuphR8nudFrAFi\n\
aGxxkgma/Es/BA+tbeCTUR106AL1ENcWA4FX3K+E9BBL0/7X5rj5nIgX/R/1ubhk\n\
KWw9gfqPG3KeAtIdcv/uTO1yXv50vqaPvE1CRChvzdS/ZEBqQ5oVvLTPZ3VEicQj\n\
lytKgN9cLnxbwtuvLUK7eyRPfJW/ksddOzP8VBBniolYnRCD2jrMRZ8nBM2ZWYwn\n\
XnwYeOAHV+W9tOhAImwRwKF/95yAsVwd21ryHMJBcGH70qLagZ7Ttyt++qO/6+KA\n\
XJuKwZqjRlEtSEz8gZQeFfVYgcwSfo96oSMAzVr7V0L6HSDLRnpb6xxmbPdqNol4\n\
tQIDAQAB\n\
-----END PUBLIC KEY-----";

sgx_quote_t* quote_from_attestation_report_body(const char* report_body)
{
    JSON_Value* root = json_parse_string(report_body);
    const char* base64_quote = json_object_get_string(json_object(root), "isvEnclaveQuoteBody");
    std::string _quote = base64_decode(base64_quote);
    json_value_free(root);

    sgx_quote_t* quote = (sgx_quote_t*)malloc(_quote.size());
    memcpy(quote, _quote.c_str(), _quote.size());
    return quote;
}

// check that certain mrenclave and hash(enclave_pk) is in report body
int verify_mrenclave_in_quote(sgx_quote_t* quote, mrenclave_t* mrenclave)
{
    if (quote == NULL || mrenclave == NULL)
    {
        return -1;
    }
    return memcmp(quote->report_body.mr_enclave.m, mrenclave->m, 32);
}

int verify_enclave_pk_in_quote(
    sgx_quote_t* quote, const unsigned char* enclave_pk_der, int enclave_pk_len)
{
    if (quote == NULL || enclave_pk_der == NULL)
    {
        return -1;
    }

    const unsigned char* tmp = enclave_pk_der;

    // get ECDSA pk from DER
    EC_KEY* pubkey = NULL;
    d2i_EC_PUBKEY(&pubkey, &tmp, enclave_pk_len);
    if (pubkey == NULL)
    {
        LOG_ERROR("IAS: Pubkey error: %s", ERR_error_string(ERR_get_error(), NULL));
        return -1;
    }
    const EC_POINT* p = EC_KEY_get0_public_key(pubkey);
    const EC_GROUP* grp = EC_KEY_get0_group(pubkey);
    BIGNUM* x = BN_new();
    BIGNUM* y = BN_new();
    EC_POINT_get_affine_coordinates_GFp(grp, p, x, y, NULL);

    unsigned char* x_bin = (unsigned char*)malloc(BN_num_bytes(x));
    unsigned char* y_bin = (unsigned char*)malloc(BN_num_bytes(y));

    BN_bn2bin(x, x_bin);
    BN_bn2bin(y, y_bin);

    // get hash(enclave_pk)
    unsigned char pk_hash[32];
    {
        SHA256_CTX sha256;
        SHA256_Init(&sha256);
        SHA256_Update(&sha256, x_bin, BN_num_bytes(x));
        SHA256_Update(&sha256, y_bin, BN_num_bytes(y));
        SHA256_Final(pk_hash, &sha256);
    }

    free(x_bin);
    free(y_bin);
    BN_free(x);
    BN_free(y);
    EC_KEY_free(pubkey);

    return memcmp(quote->report_body.report_data.d, pk_hash, 32);
}

int verify_attestation_report(uint8_t* json_bytes, size_t json_len, mrenclave_t* mrenclave)
{
    std::string json((const char*)json_bytes, json_len);

    // first get intel pub key for verification
    BIO* mem = BIO_new(BIO_s_mem());
    if (BIO_puts(mem, INTEL_PUB_PEM) < 1)
    {
        LOG_ERROR("IAS: Mem NULL, error: %s", ERR_error_string(ERR_get_error(), NULL));
        return IAS_ERROR;
    }

    EVP_PKEY* intel_pkey = NULL;
    if (!PEM_read_bio_PUBKEY(mem, &intel_pkey, 0, 0))
    {
        LOG_ERROR(
            "IAS: Can not parse Intel PK from pem: %s", ERR_error_string(ERR_get_error(), NULL));
        return IAS_ERROR;
    }

    RSA* intel_pubkey_rsa = EVP_PKEY_get1_RSA(intel_pkey);

    JSON_Value* root = json_parse_string(json.c_str());
    if (root == NULL)
    {
        LOG_ERROR("IAS: Failed to parse JSON");
        return IAS_ERROR;
    }

    const char* base64_enclave_pk = json_object_get_string(json_object(root), "EnclavePk");
    const char* base64_signature = json_object_get_string(json_object(root), "IASReport-Signature");
    const char* base64_signing_cert =
        json_object_get_string(json_object(root), "IASReport-Signing-Certificate");
    const char* base64_report_body = json_object_get_string(json_object(root), "IASResponseBody");
    json_value_free(root);

    std::string signature = base64_decode(base64_signature);
    std::string report_body = base64_decode(base64_report_body);

    // next: compute hash IASReport Body
    unsigned char sig_hash[32];
    {
        SHA256_CTX sha256;
        SHA256_Init(&sha256);
        SHA256_Update(&sha256, report_body.c_str(), report_body.size());
        SHA256_Final(sig_hash, &sha256);
    }

    // next: verify
    int ret = RSA_verify(NID_sha256, sig_hash, 32, (const unsigned char*)signature.c_str(),
        signature.size(), intel_pubkey_rsa);
    if (ret == 0)
    {
        LOG_ERROR("IAS: Invalid IASReport signature");
        return IAS_ERROR;
    }
    else if (ret == -1)
    {
        LOG_ERROR("IAS: IASReport signature validation  error");
        return IAS_ERROR;
    }
    else
    {
        LOG_DEBUG("IAS: Valid IASReport");
    }

    sgx_quote_t* quote = quote_from_attestation_report_body(report_body.c_str());
    std::string enclave_pk = base64_decode(base64_enclave_pk);

    // check that IASReport includes enclave PK
    if (verify_enclave_pk_in_quote(
            quote, (const unsigned char*)enclave_pk.c_str(), enclave_pk.size()) != 0)
    {
        LOG_ERROR("IAS: Enclave PK does not match attestation report");
        return IAS_ERROR;
    }

    // check that correct MRENCLAVE is present in quote
    if (verify_mrenclave_in_quote(quote, mrenclave) != 0)
    {
        LOG_ERROR("IAS: Chaincode MRENCLAVE does not match attestation report");
        return IAS_ERROR;
    }

    EVP_PKEY_free(intel_pkey);
    RSA_free(intel_pubkey_rsa);

    return IAS_SUCCESS;
}
