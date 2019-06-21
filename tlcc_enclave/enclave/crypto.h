/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#pragma once

#include <openssl/x509.h>
#include <stdint.h>

int validate_cert(uint8_t* cert_bytes, uint32_t cert_bytes_len, X509_STORE* root_certs);
int store_root_cert(uint8_t* cert_bytes, uint32_t cert_bytes_len, X509_STORE* root_certs);
int verify_signature(const unsigned char** sig_bytes,
    uint32_t sig_bytes_len,
    uint8_t* input,
    uint32_t input_len,
    uint8_t* cert_bytes,
    uint32_t cert_bytes_len,
    X509_STORE* root_certs);
int verify_enclave_signature(const unsigned char** sig_bytes,
    uint32_t sig_bytes_len,
    uint8_t* input,
    uint32_t input_len,
    const unsigned char** pk,
    uint32_t pk_len);
