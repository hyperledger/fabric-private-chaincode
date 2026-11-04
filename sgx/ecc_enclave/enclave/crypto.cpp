/*
 * Copyright IBM Corp. All Rights Reserved.
 * Copyright 2020 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "crypto.h"
#include <openssl/bio.h>
#include <openssl/pem.h>
#include <openssl/x509.h>
#include <openssl/x509v3.h>
#include <stdexcept>
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

bool validate_message_signature(
    const ByteArray signature, const ByteArray message, const ByteArray signer_cert)
{
    try
    {
        // deserialize public key
        pdo::crypto::sig::PublicKey pk(extract_encoded_public_key(signer_cert));

        // check signature
        int r = pk.VerifySignature(message, signature);
        COND2ERR(r != 1);
    }
    catch (const std::exception& e)
    {
        COND2LOGERR(true, e.what());
    }

    return true;

err:
    return false;
}

std::string extract_encoded_public_key(const ByteArray cert_pem)
{
    BIO* certBio = BIO_new(BIO_s_mem());
    if (!certBio)
    {
        throw std::runtime_error("cannot create cert bio");
    }

    int r = BIO_write(certBio, cert_pem.data(), cert_pem.size());
    if (r <= 0)
    {
        BIO_free_all(certBio);
        throw std::runtime_error("cannot read pem data");
    }

    X509* cert = PEM_read_bio_X509(certBio, NULL, NULL, NULL);
    if (!cert)
    {
        BIO_free_all(certBio);
        throw std::runtime_error("cannot parse cert");
    }

    // extract pk from cert and convert it back to pem format since PDO crypto requires encoded pk
    EVP_PKEY* pubkey = X509_get_pubkey(cert);
    if (!pubkey)
    {
        X509_free(cert);
        BIO_free_all(certBio);
        throw std::runtime_error("cannot get public key from cert");
    }
    EC_KEY* eckey = EVP_PKEY_get1_EC_KEY(pubkey);
    if (!eckey)
    {
        EVP_PKEY_free(pubkey);
        X509_free(cert);
        BIO_free_all(certBio);
        throw std::runtime_error("cannot get EC public key");
    }

    BIO* pkBio = BIO_new(BIO_s_mem());
    if (!pkBio)
    {
        EVP_PKEY_free(pubkey);
        EC_KEY_free(eckey);
        X509_free(cert);
        BIO_free_all(certBio);
        throw std::runtime_error("cannot create cert bio");
    }

    r = PEM_write_bio_EC_PUBKEY(pkBio, eckey);
    if (r == 0)
    {
        EVP_PKEY_free(pubkey);
        EC_KEY_free(eckey);
        X509_free(cert);
        BIO_free_all(certBio);
        throw std::runtime_error("cannot create cert bio");
    }
    int keylen = BIO_pending(pkBio);
    ByteArray pem_str(keylen + 1);
    r = BIO_read(pkBio, pem_str.data(), keylen);
    if (r <= 0)
    {
        EVP_PKEY_free(pubkey);
        EC_KEY_free(eckey);
        X509_free(cert);
        BIO_free_all(certBio);
        throw std::runtime_error("cannot serialize public key");
    }
    pem_str[keylen] = '\0';
    BIO_free_all(pkBio);

    EVP_PKEY_free(pubkey);
    EC_KEY_free(eckey);
    X509_free(cert);
    BIO_free_all(certBio);

    return std::string((const char*)pem_str.data(), pem_str.size());
}

std::string extract_subject_from_cert(const ByteArray cert_pem)
{
    BIO* certBio = BIO_new(BIO_s_mem());
    if (!certBio)
    {
        throw std::runtime_error("cannot create cert bio");
    }

    int r = BIO_write(certBio, cert_pem.data(), cert_pem.size());
    if (r <= 0)
    {
        BIO_free_all(certBio);
        throw std::runtime_error("cannot read pem data");
    }

    X509* cert = PEM_read_bio_X509(certBio, NULL, NULL, NULL);
    if (!cert)
    {
        BIO_free_all(certBio);
        throw std::runtime_error("cannot parse cert");
    }

    char* subj = X509_NAME_oneline(X509_get_subject_name(cert), NULL, 0);
    if (!subj)
    {
        X509_free(cert);
        BIO_free_all(certBio);
        throw std::runtime_error("cert does not contain a subject");
    }
    std::string out(subj);
    free(subj);

    X509_free(cert);
    BIO_free_all(certBio);

    // Go: CN=peer0.org1.example.com,OU=COP,L=San Francisco,ST=California,C=US
    // This: /C=US/ST=California/L=San Francisco/OU=COP/CN=peer0.org1.example.com

    return out;
}
