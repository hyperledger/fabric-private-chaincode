/*
 * Copyright 2019 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "error-codes.h"
#include <string>
#include "spectrum-auction-message.h"

ClockAuction::ErrorReport::ErrorReport() {}

void ClockAuction::ErrorReport::set(error_codes_e e, const std::string& s)
{
    ec_ = e;
    errorString_ = s;
}

void ClockAuction::ErrorReport::toStatusJsonString(std::string& jsonString)
{
    ClockAuction::SpectrumAuctionMessage msg;
    msg.toStatusJsonString(ec_, errorString_, jsonString);
}

void ClockAuction::ErrorReport::toWrappedStatusJsonString(std::string& jsonString)
{
    ClockAuction::SpectrumAuctionMessage msg;
    JSON_Object* root_object = ClockAuction::JsonUtils::openJsonObject(NULL);
    msg.toWrappedStatusJsonObject(root_object, ec_, errorString_);
    ClockAuction::JsonUtils::closeJsonObject(root_object, &jsonString);
}

bool ClockAuction::ErrorReport::isSuccess()
{
    if (ec_ == EC_SUCCESS)
    {
        return true;
    }
    return false;
}
