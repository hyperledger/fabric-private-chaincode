/*
 * Copyright 2019 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "utils.h"

JSON_Object* ClockAuction::JsonUtils::openJsonObject(const char* c_str)
{
    JSON_Value* root_value;

    if (c_str == NULL)  // create an empty json object
    {
        root_value = json_value_init_object();
    }
    else  // parse the input json string
    {
        root_value = json_parse_string(c_str);
        if (root_value == NULL)
        {
            return NULL;
        }
    }

    return json_value_get_object(root_value);
}

void ClockAuction::JsonUtils::closeJsonObject(JSON_Object* root_object, std::string* str)
{
    if (!root_object)
    {
        return;
    }

    JSON_Value* root_value = json_object_get_wrapping_value(root_object);
    if (str != NULL)  // serialize the json object to string
    {
        char* serialized_string = json_serialize_to_string(root_value);
        str->assign(serialized_string);
        json_free_serialized_string(serialized_string);
    }

    // eventually, free the root value
    json_value_free(root_value);
}

bool ClockAuction::JsonUtils::isInteger(double d)
{
    return d == ((double)(int)d);
}
