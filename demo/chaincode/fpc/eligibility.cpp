/*
 * Copyright 2019 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */
#include "eligibility.h"
#include "utils.h"

bool ClockAuction::Eligibility::toJsonObject(JSON_Object* root_object) const
{
    json_object_set_number(root_object, "bidderId", bidderId_);
    json_object_set_number(root_object, "number", number_);
    return true;
}

bool ClockAuction::Eligibility::fromJsonObject(const JSON_Object* root_object)
{
    {
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT,
            !json_object_has_value_of_type(root_object, "bidderId", JSONNumber));
        double d = json_object_get_number(root_object, "bidderId");
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, !ClockAuction::JsonUtils::isInteger(d));
        bidderId_ = (uint32_t)d;
    }
    {
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT,
            !json_object_has_value_of_type(root_object, "number", JSONNumber));
        double d = json_object_get_number(root_object, "number");
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, !ClockAuction::JsonUtils::isInteger(d));
        number_ = (uint32_t)d;
    }
    return true;
}

bool ClockAuction::Eligibility::matchBidderId(uint32_t bidderId) const
{
    return bidderId == bidderId_;
}

uint32_t ClockAuction::Eligibility::getNumber() const
{
    return number_;
}
