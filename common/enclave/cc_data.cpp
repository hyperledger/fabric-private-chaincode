/*
 * Copyright 2020 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include <string>

#include "base64.h"
#include "cc_data.h"
#include "error.h"
#include "logging.h"

#include <pb_decode.h>
#include <pb_encode.h>
#include "protos/fpc/fpc.pb.h"

#include "attestation-api/attestation/attestation.h"
#include "crypto.h"

// ecc enclave global variable -- allocated dynamically
cc_data* g_cc_data = NULL;

bool cc_data::generate()
{
    return generate_keys();
}

bool cc_data::generate_keys()
{
    try
    {
        signature_key_.Generate();                          // enclave_sk, private signing key
        verification_key_ = signature_key_.GetPublicKey();  // enclave_vk, public verifying key
        decryption_key_.Generate();                         // enclave_dk, private decryption key
        encryption_key_ = decryption_key_.GetPublicKey();   // enclave_ek, public encryption key
        cc_decryption_key_.Generate();                      // chaincode_dk, private decryption key
        cc_encryption_key_ =
            cc_decryption_key_.GetPublicKey();  // chaincode_ek, public encryption key

        // generate state encryption key
        state_encryption_key_ = pdo::crypto::skenc::GenerateKey();

        std::string s;
        s = verification_key_.Serialize();
        LOG_DEBUG("enclave verification key: %s", s.c_str());
        s = get_enclave_id();
        LOG_DEBUG("enclave id: %s", s.c_str());
    }
    catch (...)
    {
        LOG_ERROR("error creating cryptographic keys");
        return false;
    }

    return true;
}

bool cc_data::build_attested_data(ByteArray& attested_data)
{
    // estimate attested data size
    // NOTE: the size is roughly estimated, it should be adapted
    size_t attested_data_proto_size = attestation_parameters_.size() + cc_parameters_.size() +
                                      host_parameters_.size() + (1 << 13);
    uint8_t* attested_data_proto;
    pb_ostream_t ostream;
    bool b;

    CATCH(b, attested_data.resize(attested_data_proto_size));
    COND2ERR(!b);

    attested_data_proto = attested_data.data();

    ostream = pb_ostream_from_buffer(attested_data_proto, attested_data_proto_size);

    {
        // fpc_AttestedData_cc_params_tag
        COND2ERR(!pb_encode_tag(&ostream, PB_WT_STRING, fpc_AttestedData_cc_params_tag));
        COND2ERR(!pb_encode_string(
            &ostream, (const unsigned char*)cc_parameters_.data(), cc_parameters_.size()));
    }

    {
        // fpc_AttestedData_host_params_tag
        COND2ERR(!pb_encode_tag(&ostream, PB_WT_STRING, fpc_AttestedData_host_params_tag));
        COND2ERR(!pb_encode_string(
            &ostream, (const unsigned char*)host_parameters_.data(), host_parameters_.size()));
    }

    {
        // fpc_AttestedData_enclave_vk_tag
        std::string s = verification_key_.Serialize();
        COND2ERR(!pb_encode_tag(&ostream, PB_WT_STRING, fpc_AttestedData_enclave_vk_tag));
        COND2ERR(!pb_encode_string(&ostream, (const unsigned char*)s.c_str(), s.length()));
    }

    {
        // NOTE: for the one-chaincode-one-enclave FPC-Lite version, the chaincode encryption key
        // is serialized directly in the attested data message.
        // This is a (momentary) short-cut over the FPC and FPC Lite specification in
        // `docs/design/fabric-v2+/fpc-registration.puml` and
        // `docs/design/fabric-v2+/fpc-key-dist.puml`

        // fpc_AttestedData_chaincode_ek_tag
        std::string s = cc_encryption_key_.Serialize();
        COND2ERR(!pb_encode_tag(&ostream, PB_WT_STRING, fpc_AttestedData_chaincode_ek_tag));
        COND2ERR(!pb_encode_string(&ostream, (const unsigned char*)s.c_str(), s.length()));
    }

    // resize array to fit written data
    attested_data.resize(ostream.bytes_written);

    return true;

err:
    return false;
}

bool cc_data::get_credentials(const uint8_t* attestation_parameters,
    uint32_t ap_size,
    const uint8_t* cc_parameters,
    uint32_t ccp_size,
    const uint8_t* host_parameters,
    uint32_t hp_size,
    uint8_t* credentials,
    uint32_t credentials_max_size,
    uint32_t* credentials_size)
{
    ByteArray attested_data;
    ByteArray attestation;
    // NOTE: attestation's max length should be adapted
    const uint32_t attestation_max_length = 1 << 12;
    uint32_t attestation_length;
    bool b;
    std::string attestation_parameters_b64;
    std::string attestation_parameters_json;

    // init parameters
    CATCH(b,
        attestation_parameters_.assign(attestation_parameters, attestation_parameters + ap_size));
    COND2ERR(!b);

    CATCH(b, cc_parameters_.assign(cc_parameters, cc_parameters + ccp_size));
    COND2ERR(!b);

    CATCH(b, host_parameters_.assign(host_parameters, host_parameters + hp_size));
    COND2ERR(!b);

    // build attested data
    b = build_attested_data(attested_data);
    COND2ERR(!b);

    // init attestation
    attestation_parameters_b64 = std::string((const char*)attestation_parameters, ap_size);
    attestation_parameters_json = base64_decode(attestation_parameters_b64);
    b = init_attestation(
        (uint8_t*)attestation_parameters_json.c_str(), attestation_parameters_json.length());
    COND2LOGERR(!b, "cannot init attestation");

    // get attestation
    CATCH(b, attestation.resize(attestation_max_length));
    COND2ERR(!b);

    b = get_attestation(attested_data.data(), attested_data.size(), attestation.data(),
        attestation_max_length, &attestation_length);
    COND2LOGERR(!b, "cannot get attestation");

    {
        // build credentials (Attested_Data || Attestation)
        pb_ostream_t ostream;
        ostream = pb_ostream_from_buffer(credentials, credentials_max_size);
        {
            pb_ostream_t ostream_any;
            ByteArray buffer;
            // NOTE: buffer size should be adapted
            CATCH(b, buffer.resize(attested_data.size() + 1024));
            COND2ERR(!b);

            {
                // serialize the Any type
                ostream_any = pb_ostream_from_buffer(buffer.data(), buffer.size());

                COND2ERR(
                    !pb_encode_tag(&ostream_any, PB_WT_STRING, google_protobuf_Any_type_url_tag));
                // NOTE: the url type string is necessary,
                //       and the type after last '/' must match the serialized message type
                std::string s("github.com/fpc/fpc.AttestedData");
                COND2ERR(
                    !pb_encode_string(&ostream_any, (const unsigned char*)s.c_str(), s.length()));

                COND2ERR(!pb_encode_tag(&ostream_any, PB_WT_STRING, google_protobuf_Any_value_tag));
                COND2ERR(!pb_encode_string(&ostream_any, (const unsigned char*)attested_data.data(),
                    attested_data.size()));
            }

            // fpc_Credentials_serialized_attested_data_tag
            COND2ERR(!pb_encode_tag(
                &ostream, PB_WT_STRING, fpc_Credentials_serialized_attested_data_tag));
            COND2ERR(!pb_encode_string(
                &ostream, (const unsigned char*)buffer.data(), ostream_any.bytes_written));
        }

        {
            // fpc_Credentials_attestation_tag
            COND2ERR(!pb_encode_tag(&ostream, PB_WT_STRING, fpc_Credentials_attestation_tag));
            COND2ERR(!pb_encode_string(
                &ostream, (const unsigned char*)attestation.data(), attestation_length));
        }

        // set output credential size
        *credentials_size = ostream.bytes_written;
    }

    return true;

err:
    return false;
}

std::string cc_data::get_enclave_id()
{
    // get enclave vk
    std::string s = verification_key_.Serialize();
    // hash
    ByteArray h = pdo::crypto::ComputeMessageHash(ByteArray(s.c_str(), s.c_str() + s.length()));
    // hex encode
    std::string hex = ByteArrayToHexEncodedString(h);
    // normalize
    std::transform(
        hex.begin(), hex.end(), hex.begin(), [](unsigned char c) { return std::toupper(c); });

    return hex;
}

std::string cc_data::get_channel_id()
{
    fpc_CCParameters cc_params = fpc_CCParameters_init_zero;
    pb_istream_t istream =
        pb_istream_from_buffer((const unsigned char*)cc_parameters_.data(), cc_parameters_.size());
    bool b = pb_decode(&istream, fpc_CCParameters_fields, &cc_params);
    COND2LOGERR(!b, PB_GET_ERROR(&istream));

    return std::string(cc_params.channel_id);

err:
    // return empty string in case of error
    return std::string();
}

bool cc_data::sign_message(const ByteArray& message, ByteArray& signature) const
{
    bool b;
    CATCH(b, signature = signature_key_.SignMessage(message));
    COND2LOGERR(!b, "message signing failed");

    return true;

err:
    return false;
}

// decrypts a key transport message using asymmetric encryption provided by pdo::crypto with
// the chaincode decryption key (cc_data.cc_decryption_key_)
bool cc_data::decrypt_key_transport_message(
    const ByteArray& encrypted_key_transport_message, ByteArray& key_transport_message) const
{
    bool b;
    CATCH(b,
        key_transport_message = cc_decryption_key_.DecryptMessage(encrypted_key_transport_message));
    COND2LOGERR(!b, "key transport message decryption failed");

    return true;

err:
    return false;
}

bool cc_data::decrypt_state_value(const ByteArray& encrypted_value, ByteArray& value) const
{
    bool b;
    COND2LOGERR(
        encrypted_value.size() <= pdo::crypto::constants::IV_LEN + pdo::crypto::constants::TAG_LEN,
        "encrypted value too short");
    CATCH(b, value = pdo::crypto::skenc::DecryptMessage(state_encryption_key_, encrypted_value));
    COND2LOGERR(!b, "state decryption failed");

    return true;

err:
    return false;
}

bool cc_data::encrypt_state_value(const ByteArray& value, ByteArray& encrypted_value) const
{
    bool b;
    CATCH(b, encrypted_value = pdo::crypto::skenc::EncryptMessage(state_encryption_key_, value));
    COND2LOGERR(!b, "state encryption failed");

    return true;

err:
    return false;
}

int cc_data::estimate_encrypted_state_value_length(const int value_len)
{
    return (value_len + pdo::crypto::constants::IV_LEN + pdo::crypto::constants::TAG_LEN) * 2;
}
