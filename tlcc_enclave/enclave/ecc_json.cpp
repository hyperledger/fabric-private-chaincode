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
