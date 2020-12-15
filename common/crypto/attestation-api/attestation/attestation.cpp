/*
 * Copyright 2020 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "attestation.h"
#include <string>
#include "attestation_tags.h"
#include "base64.h"
#include "error.h"
#include "logging.h"
#include "parson.h"
#include "pdo/common/crypto/crypto.h"
#include "pdo/common/jsonvalue.h"
#include "sgx_quote.h"
#include "sgx_utils.h"
#include "types.h"

/**********************************************************************************************************************
 * C prototype declarations for the ocalls
 * *******************************************************************************************************************/
#ifdef __cplusplus
extern "C" {
#endif
sgx_status_t ocall_init_quote(
    uint8_t* target, uint32_t target_len, uint8_t* egid, uint32_t egid_len);
sgx_status_t ocall_get_quote(uint8_t* spid,
    uint32_t spid_len,
    uint8_t* sig_rl,
    uint32_t sig_rl_len,
    uint32_t sign_type,
    uint8_t* report,
    uint32_t report_len,
    uint8_t* quote,
    uint32_t max_quote_len,
    uint32_t* actual_quote_len);
#ifdef __cplusplus
}
#endif /* __cplusplus */

/**********************************************************************************************************************
 * Attestation APIs
 * *******************************************************************************************************************/
typedef struct
{
    bool initialized;
    sgx_spid_t spid;
    ByteArray sig_rl;
    std::string attestation_type;
    uint32_t sign_type;
} attestation_state_t;

attestation_state_t g_attestation_state = {0};

bool init_attestation(uint8_t* params, uint32_t params_length)
{
    if (params == NULL)
    {
        LOG_ERROR("bad attestation init params");
        return false;
    }

    // open json
    std::string params_string((char*)params, params_length);
    JsonValue root(json_parse_string(params_string.c_str()));
    LOG_DEBUG("attestation params: %s", params_string.c_str());
    COND2LOGERR(root == NULL, "cannot parse attestation params");

    {  // set attestation type
        const char* p;
        p = json_object_get_string(json_object(root), ATTESTATION_TYPE_TAG);
        COND2LOGERR(p == NULL, "no attestation type provided");
        g_attestation_state.attestation_type.assign(p, p + strlen(p));

        if (g_attestation_state.attestation_type.compare(SIMULATED_TYPE_TAG) == 0)
        {
            // terminate init successfully
            goto init_success;
        }

        // check other types for EPID
        g_attestation_state.sign_type = -1;
        if (g_attestation_state.attestation_type.compare(EPID_LINKABLE_TYPE_TAG) == 0)
        {
            g_attestation_state.sign_type = SGX_LINKABLE_SIGNATURE;
        }

        if (g_attestation_state.attestation_type.compare(EPID_UNLINKABLE_TYPE_TAG) == 0)
        {
            g_attestation_state.sign_type = SGX_UNLINKABLE_SIGNATURE;
        }

        COND2LOGERR(g_attestation_state.sign_type == -1, "wrong attestation type");
    }

    // keep initializing EPID attestation with SPID and sig_rl

    {  // set SPID
        // std::string hex_spid;
        HexEncodedString hex_spid;
        const char* p;
        p = json_object_get_string(json_object(root), SPID_TAG);
        COND2LOGERR(p == NULL, "no spid provided");

        hex_spid.assign(p);
        COND2LOGERR(hex_spid.length() != sizeof(sgx_spid_t) * 2, "wrong spid length");
        // translate hex spid to binary
        try
        {
            ByteArray ba = HexEncodedStringToByteArray(hex_spid);
            memcpy(g_attestation_state.spid.id, ba.data(), ba.size());
        }
        catch (...)
        {
            COND2LOGERR(true, "bad hex spid");
        }
    }

    {  // set sig_rl
        const char* p;
        p = json_object_get_string(json_object(root), SIG_RL_TAG);
        COND2LOGERR(p == NULL, "no sig_rl provided");
        g_attestation_state.sig_rl.assign(p, p + strlen(p));
    }

init_success:
    g_attestation_state.initialized = true;
    return true;

err:
    return false;
}

bool get_attestation(uint8_t* statement,
    uint32_t statement_length,
    uint8_t* attestation,
    uint32_t attestation_max_length,
    uint32_t* attestation_length)
{
    sgx_report_t report;
    sgx_report_data_t report_data = {0};
    sgx_target_info_t qe_target_info = {0};
    sgx_epid_group_id_t egid = {0};
    std::string b64attestation;
    int ret;

    COND2LOGERR(!g_attestation_state.initialized, "attestation not initialized");
    COND2LOGERR(statement == NULL, "bad input statement");
    COND2LOGERR(attestation == NULL, "bad input attestation buffer");
    COND2LOGERR(attestation_length == NULL || attestation_max_length == 0,
        "bad input attestation buffer size");

    if (g_attestation_state.attestation_type.compare(SIMULATED_TYPE_TAG) == 0)
    {
        std::string zero("0");
        b64attestation = base64_encode((const unsigned char*)zero.c_str(), zero.length());
    }
    else
    {
        ocall_init_quote(
            (uint8_t*)&qe_target_info, sizeof(qe_target_info), (uint8_t*)&egid, sizeof(egid));

        ByteArray ba_statement(statement, statement + statement_length);
        ByteArray rd = pdo::crypto::ComputeMessageHash(ba_statement);
        // ComputeMessageHash uses sha256
        COND2LOGERR(rd.size() > sizeof(sgx_report_data_t), "report data too long");
        memcpy(&report_data, rd.data(), rd.size());

        ret = sgx_create_report(&qe_target_info, &report_data, &report);
        COND2LOGERR(SGX_SUCCESS != ret, "error creating report");

        ocall_get_quote((uint8_t*)&g_attestation_state.spid, (uint32_t)sizeof(sgx_spid_t),
            g_attestation_state.sig_rl.data(), g_attestation_state.sig_rl.size(),
            g_attestation_state.sign_type, (uint8_t*)&report, sizeof(report), attestation,
            attestation_max_length, attestation_length);
        COND2LOGERR(*attestation_length == 0, "error get quote");

        // convert to base64 (accepted by IAS)
        b64attestation = base64_encode((const unsigned char*)attestation, *attestation_length);
        COND2LOGERR(b64attestation.length() > attestation_max_length,
            "not enough space for b64 conversion");
        memcpy(attestation, b64attestation.c_str(), b64attestation.length());
        *attestation_length = b64attestation.length();
    }

    {
        // package the output
        size_t serialization_size = 0;
        JsonValue root_value(json_value_init_object());
        JSON_Object* root_object = json_value_get_object(root_value);
        COND2LOGERR(root_object == NULL, "can't create json");

        COND2LOGERR(JSONFailure == json_object_set_string(root_object, ATTESTATION_TYPE_TAG,
                                       g_attestation_state.attestation_type.c_str()),
            "error serializing attestation type");
        COND2LOGERR(JSONFailure == json_object_set_string(
                                       root_object, ATTESTATION_TAG, b64attestation.c_str()),
            "error serializing attestation");

        serialization_size = json_serialization_size(root_value);
        COND2LOGERR(
            serialization_size > attestation_max_length, "not enough space for b64 conversion");

        COND2LOGERR(JSONFailure == json_serialize_to_buffer(
                                       root_value, (char*)attestation, serialization_size),
            "error packaging attestation");
        // remove terminating null byte (if any)
        if (attestation[serialization_size] == '\0')
        {
            serialization_size -= 1;
        }

        *attestation_length = serialization_size;
    }

    return true;
err:
    return false;
}
