/*
 * Copyright 2020 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#pragma once

#include "pdo/common/crypto/crypto.h"
#include "pdo/common/types.h"

class cc_data
{
private:
    pdo::crypto::sig::PublicKey verification_key_;      // enclave_vk
    pdo::crypto::sig::PrivateKey signature_key_;        // enclave_sk
    pdo::crypto::pkenc::PublicKey encryption_key_;      // enclave_ek
    pdo::crypto::pkenc::PrivateKey decryption_key_;     // enclave_dk
    pdo::crypto::pkenc::PublicKey cc_encryption_key_;   // chaincode_ek
    pdo::crypto::pkenc::PrivateKey cc_decryption_key_;  // chaincode_dk

    ByteArray state_encryption_key_;  // or sek, for encrypting key-value pairs

    ByteArray attestation_parameters_;
    ByteArray cc_parameters_;
    ByteArray host_parameters_;

    bool generate_keys();

    bool build_attested_data(ByteArray& attested_data);

public:
    bool generate();

    bool get_credentials(const uint8_t* attestation_parameters,
        uint32_t ap_size,
        const uint8_t* cc_parameters,
        uint32_t ccp_size,
        const uint8_t* host_parameters,
        uint32_t hp_size,
        uint8_t* credentials,
        uint32_t credentials_max_size,
        uint32_t* credentials_size);

    std::string get_enclave_id();
    std::string get_channel_id();

    bool sign_message(const ByteArray& message, ByteArray& signature) const;
    bool decrypt_key_transport_message(
        const ByteArray& encrypted_key_transport_message, ByteArray& key_transport_message) const;
    bool decrypt_state_value(const ByteArray& encrypted_value, ByteArray& value) const;
    bool encrypt_state_value(const ByteArray& value, ByteArray& encrypted_value) const;
    static int estimate_encrypted_state_value_length(const int value_len);
};

extern cc_data* g_cc_data;
