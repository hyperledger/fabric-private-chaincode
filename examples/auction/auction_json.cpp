/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "auction_json.h"
#include "parson.h"

int unmarshal_auction(auction_t* auction, const char* json_bytes, uint32_t json_len)
{
    JSON_Value* root = json_parse_string(json_bytes);
    auction->name = json_object_get_string(json_object(root), "name");
    auction->is_open = json_object_get_boolean(json_object(root), "is_open");
    json_value_free(root);
    return 1;
}

int unmarshal_bid(bid_t* bid, const char* json_bytes, uint32_t json_len)
{
    JSON_Value* root = json_parse_string(json_bytes);
    bid->bidder_name = json_object_get_string(json_object(root), "bidder");
    bid->value = json_object_get_number(json_object(root), "value");
    json_value_free(root);
    return 1;
}

std::string marshal_auction(auction_t* the_auction)
{
    JSON_Value* root_value = json_value_init_object();
    JSON_Object* root_object = json_value_get_object(root_value);
    json_object_set_string(root_object, "name", the_auction->name.c_str());
    json_object_set_boolean(root_object, "is_open", the_auction->is_open);
    char* serialized_string = json_serialize_to_string(root_value);
    std::string out(serialized_string);
    json_free_serialized_string(serialized_string);
    json_value_free(root_value);
    return out;
}

std::string marshal_bid(bid_t* bid)
{
    JSON_Value* root_value = json_value_init_object();
    JSON_Object* root_object = json_value_get_object(root_value);
    json_object_set_string(root_object, "bidder", bid->bidder_name.c_str());
    json_object_set_number(root_object, "value", bid->value);
    char* serialized_string = json_serialize_to_string(root_value);
    std::string out(serialized_string);
    json_free_serialized_string(serialized_string);
    json_value_free(root_value);
    return out;
}
