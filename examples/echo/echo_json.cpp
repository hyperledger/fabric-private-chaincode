/*
 * Copyright 2019 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "echo_json.h"
#include "parson.h"

int unmarshal(echo_t* echo, const char* json_bytes, uint32_t json_len)
{
    JSON_Value* root = json_parse_string(json_bytes);
    echo->echo_string = json_object_get_string(json_object(root), "echo_string");
    json_value_free(root);
    return 1;
}

std::string marshal(echo_t* echo)
{
    JSON_Value* root_value = json_value_init_object();
    JSON_Object* root_object = json_value_get_object(root_value);
    json_object_set_string(root_object, "echo_string", echo->echo_string.c_str());
    char* serialized_string = json_serialize_to_string(root_value);
    std::string out(serialized_string);
    json_free_serialized_string(serialized_string);
    json_value_free(root_value);
    return out;
}
