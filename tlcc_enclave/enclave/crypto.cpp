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

#include "crypto.h"

#include "logging.h"

// openssll
#include <openssl/bio.h>
#include <openssl/ecdsa.h>
#include <openssl/err.h>
#include <openssl/pem.h>
#include <openssl/x509_vfy.h>
#include "openssl/sha.h"

int store_root_cert(uint8_t* cert_bytes, uint32_t cert_bytes_len, X509_STORE* root_certs)
{
    X509* cert = NULL;

    // parse root cert
    BIO* mem = BIO_new_mem_buf(cert_bytes, cert_bytes_len);
    if (!PEM_read_bio_X509(mem, &cert, 0, NULL))
    {
        LOG_ERROR("Crypto: Can not parse cert");
        BIO_free_all(mem);
        return -1;
    }
    // add root cert to root cert store
    int ret = X509_STORE_add_cert(root_certs, cert);

    X509_free(cert);
    BIO_free_all(mem);

    return ret;
}

int validate_cert(uint8_t* cert_bytes, uint32_t cert_bytes_len, X509_STORE* root_certs)
{
    X509* cert = NULL;
    int ret = 1;

    // create cert
    BIO* mem = BIO_new_mem_buf(cert_bytes, cert_bytes_len);
    if (!PEM_read_bio_X509(mem, &cert, 0, NULL))
    {
        LOG_ERROR("Crypto: can not parse cert");
        BIO_free_all(mem);
        return -1;
    }

    // get cert store ctx
    X509_STORE_CTX* ctx = X509_STORE_CTX_new();
    X509_STORE_CTX_init(ctx, root_certs, cert, NULL);

    // verify
    if ((ret = X509_verify_cert(ctx)) == 0)
    {
        LOG_ERROR("Crypto: Invalid certificate: %s",
            X509_verify_cert_error_string(X509_STORE_CTX_get_error(ctx)));
    }

    X509_STORE_CTX_free(ctx);
    X509_free(cert);
    BIO_free_all(mem);

    return ret;
}

// return 1 on success; 0 verification fail; -1 for trouble
int verify_signature(const unsigned char** sig_bytes,
    uint32_t sig_bytes_len,
    uint8_t* input,
    uint32_t input_len,
    uint8_t* cert_bytes,
    uint32_t cert_bytes_len,
    X509_STORE* root_certs)
{
    X509* cert = NULL;

    // cert_bytes as PEM to x509 cert
    BIO* mem = BIO_new_mem_buf(cert_bytes, cert_bytes_len);
    if (!PEM_read_bio_X509(mem, &cert, 0, NULL))
    {
        LOG_ERROR("Crypto: can not parse cert");
        BIO_free_all(mem);
        return -1;
    }

    // check that cert has root cert
    X509_STORE_CTX* ctx = X509_STORE_CTX_new();
    X509_STORE_CTX_init(ctx, root_certs, cert, NULL);
    int ret = X509_verify_cert(ctx);
    X509_STORE_CTX_free(ctx);

    // check if we can already abort
    if (ret == 0)
    {
        LOG_ERROR("Crypto: Invalid certificate: %s",
            X509_verify_cert_error_string(X509_STORE_CTX_get_error(ctx)));
        X509_free(cert);
        BIO_free_all(mem);
        return 0;
    }

    // get public ecdsa key
    EVP_PKEY* pubkey = X509_get_pubkey(cert);
    EC_KEY* eckey = EVP_PKEY_get1_EC_KEY(pubkey);

    ECDSA_SIG* signature = d2i_ECDSA_SIG(NULL, sig_bytes, sig_bytes_len);
    ret = ECDSA_do_verify(input, input_len, signature, eckey);

    ECDSA_SIG_free(signature);
    EC_KEY_free(eckey);
    EVP_PKEY_free(pubkey);
    X509_free(cert);
    BIO_free_all(mem);

    return ret;
}

// return 1 on success; 0 verification fail; -1 for trouble
int verify_enclave_signature(const unsigned char** sig_bytes,
    uint32_t sig_bytes_len,
    uint8_t* input,
    uint32_t input_len,
    const unsigned char** pk,
    uint32_t pk_len)
{
    // get public ecdsa key
    EVP_PKEY* pubkey = d2i_PUBKEY(NULL, pk, pk_len);
    if (pubkey == NULL)
    {
        LOG_ERROR("pubkey: %s", ERR_error_string(ERR_get_error(), NULL));
    }

    EC_KEY* eckey = EVP_PKEY_get1_EC_KEY(pubkey);
    if (eckey == NULL)
    {
        LOG_ERROR("eckey: %s", ERR_error_string(ERR_get_error(), NULL));
    }

    ECDSA_SIG* signature = d2i_ECDSA_SIG(NULL, sig_bytes, sig_bytes_len);
    if (signature == NULL)
    {
        LOG_ERROR("signature %s", ERR_error_string(ERR_get_error(), NULL));
    }

    int ret = ECDSA_do_verify(input, input_len, signature, eckey);

    ECDSA_SIG_free(signature);
    EC_KEY_free(eckey);
    EVP_PKEY_free(pubkey);

    return ret;
}
