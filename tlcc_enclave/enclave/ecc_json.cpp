/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "ecc_json.h"
#include "base64.h"
#include "parson.h"

int unmarshal_ecc_response(const uint8_t* json_bytes,
    uint32_t json_len,
    uint8_t* response_data,
    uint32_t* response_len,
    uint8_t* signature,
    uint32_t* signature_len,
    uint8_t* pk,
    uint32_t* pk_len)
{
    JSON_Value* root = json_parse_string((const char*)json_bytes);
    const char* base64_response = json_object_get_string(json_object(root), "ResponseData");
    std::string response = base64_decode(base64_response);
    memcpy(response_data, response.c_str(), response.size());
    *response_len = response.size();

    const char* base64_signature = json_object_get_string(json_object(root), "Signature");
    std::string _signature = base64_decode(base64_signature);
    memcpy(signature, _signature.c_str(), _signature.size());
    *signature_len = _signature.size();

    const char* base64_pk = json_object_get_string(json_object(root), "PublicKey");
    std::string _pk = base64_decode(base64_pk);
    memcpy(pk, _pk.c_str(), _pk.size());
    *pk_len = _pk.size();

    json_value_free(root);
    return 1;
}
