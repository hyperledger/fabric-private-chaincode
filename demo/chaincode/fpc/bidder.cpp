/*
 * Copyright 2019 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "bidder.h"
#include "common.h"
#include "utils.h"

bool ClockAuction::Bidder::toJsonObject(JSON_Object* root_object) const
{
    json_object_set_number(root_object, "id", id_);
    json_object_set_string(root_object, "displayName", displayName_.c_str());
    JSON_Value* v = json_value_init_object();
    JSON_Object* o = json_value_get_object(v);
    principal_.toJsonObject(o);
    json_object_set_value(root_object, "principal", v);
    return true;
}

bool ClockAuction::Bidder::fromJsonObject(const JSON_Object* root_object)
{
    {
        FAST_FAIL_CHECK(
            er_, EC_INVALID_INPUT, !json_object_has_value_of_type(root_object, "id", JSONNumber));
        double d = json_object_get_number(root_object, "id");
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, !ClockAuction::JsonUtils::isInteger(d));
        id_ = (uint32_t)d;
    }
    {
        const char* str = json_object_get_string(root_object, "displayName");
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, str == 0);
        displayName_ = std::string(str);
    }
    {
        JSON_Value* v = json_object_get_value(root_object, "principal");
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, v == 0);
        JSON_Object* o = json_value_get_object(v);
        FAST_FAIL_CHECK_EX(er_, &principal_.er_, EC_INVALID_INPUT, !principal_.fromJsonObject(o));
    }
    return true;
}

bool ClockAuction::Bidder::matchPrincipal(const Principal& p) const
{
    return p == principal_;
}

uint32_t ClockAuction::Bidder::getId() const
{
    return id_;
}

const ClockAuction::Principal ClockAuction::Bidder::getPrincipal() const
{
    return principal_;
}
