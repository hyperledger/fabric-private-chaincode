/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "ecc_json.h"
#include "base64.h"
#include "logging.h"
#include "parson.h"

int marshal_ecc_args(std::vector<std::string>& argss,
    std::string& encoded_string)  // assumed to be initialized but empty
{
    JSON_Value* root = json_value_init_array();
    JSON_Array* list = json_value_get_array(root);

    for (std::size_t i = 0; i < argss.size(); i++)
    {
        json_array_append_string(list, argss[i].c_str());
    }

    char* ss = json_serialize_to_string(root);
    encoded_string.append(ss);

    json_free_serialized_string(ss);
    json_value_free(root);
    return 1;
}

int unmarshal_ecc_response(const uint8_t* json_bytes,
    uint32_t json_len,
    uint8_t* response_data,
    uint32_t* response_len,
    uint8_t* signature,
    uint32_t* signature_len,
    uint8_t* pk,
    uint32_t* pk_len)
{
    if ((response_data == NULL) || (response_len == NULL) || (signature == NULL) ||
        (signature_len == NULL) || (pk == NULL) || (pk_len == NULL))
    {
        LOG_ERROR("illegal parameters");
        return 0;
    }
    JSON_Value* root = json_parse_string((const char*)json_bytes);
    const char* base64_response = json_object_get_string(json_object(root), "ResponseData");
    std::string _response = base64_decode(base64_response);
    if (_response.size() > *response_len)
    {
        LOG_ERROR(
            "response buffer too short: required %d, got %d", _response.size(), *response_len);
        return 0;
    }
    memcpy(response_data, _response.c_str(), _response.size());
    *response_len = _response.size();

    const char* base64_signature = json_object_get_string(json_object(root), "Signature");
    std::string _signature = base64_decode(base64_signature);
    if (_signature.size() > *signature_len)
    {
        LOG_ERROR(
            "signature buffer too short: required %d, got %d", _signature.size(), *signature_len);
        return 0;
    }
    memcpy(signature, _signature.c_str(), _signature.size());
    *signature_len = _signature.size();

    const char* base64_pk = json_object_get_string(json_object(root), "PublicKey");
    std::string _pk = base64_decode(base64_pk);
    if (_pk.size() > *pk_len)
    {
        LOG_ERROR("pk buffer too short: required %d, got %d", _pk.size(), *pk_len);
        return 0;
    }
    memcpy(pk, _pk.c_str(), _pk.size());
    *pk_len = _pk.size();

    json_value_free(root);
    return 1;
}
