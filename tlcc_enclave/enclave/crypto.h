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

#pragma once

#include <openssl/x509.h>
#include <stdint.h>

int validate_cert(uint8_t* cert_bytes, uint32_t cert_bytes_len, X509_STORE* root_certs);
int store_root_cert(uint8_t* cert_bytes, uint32_t cert_bytes_len, X509_STORE* root_certs);
int verify_signature(const unsigned char** sig_bytes, uint32_t sig_bytes_len, uint8_t* input,
    uint32_t input_len, uint8_t* cert_bytes, uint32_t cert_bytes_len, X509_STORE* root_certs);
int verify_enclave_signature(const unsigned char** sig_bytes, uint32_t sig_bytes_len,
    uint8_t* input, uint32_t input_len, const unsigned char** pk, uint32_t pk_len);
