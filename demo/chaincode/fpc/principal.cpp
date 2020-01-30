/*
 * Copyright 2019 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "principal.h"

ClockAuction::Principal::Principal() {}

ClockAuction::Principal::Principal(std::string m, std::string d) : mspId_(m), dn_(d) {}

bool operator==(const ClockAuction::Principal& p1, const ClockAuction::Principal& p2)
{
    if (p1.getDn() == p2.getDn() && p1.getMspId() == p2.getMspId())
    {
        return true;
    }
    return false;
}

std::string ClockAuction::Principal::getDn() const
{
    return dn_;
}

std::string ClockAuction::Principal::getMspId() const
{
    return mspId_;
}

bool ClockAuction::Principal::toJsonObject(JSON_Object* root_object) const
{
    json_object_set_string(root_object, "mspId", mspId_.c_str());
    json_object_set_string(root_object, "dn", dn_.c_str());
    return true;
}

bool ClockAuction::Principal::fromJsonObject(const JSON_Object* root_object)
{
    {
        const char* str = json_object_get_string(root_object, "mspId");
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, str == 0);
        mspId_ = std::string(str);
    }
    {
        const char* str = json_object_get_string(root_object, "dn");
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, str == 0);
        dn_ = std::string(str);
    }
    return true;
}
