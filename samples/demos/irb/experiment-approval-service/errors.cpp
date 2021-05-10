/*
 * Copyright 2019 Intel Corporation
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "errors.h"
#include <string>
#include "messages.h"

Contract::ErrorReport::ErrorReport() {}

void Contract::ErrorReport::set(error_codes_e e, const std::string& s)
{
    ec_ = e;
    errorString_ = s;
}

void Contract::ErrorReport::toStatusJsonString(std::string& jsonString)
{
    //    ClockAuction::SpectrumAuctionMessage msg;
    //    msg.toStatusJsonString(ec_, errorString_, jsonString);
}

void Contract::ErrorReport::toWrappedStatusJsonString(std::string& jsonString)
{
    //    ClockAuction::SpectrumAuctionMessage msg;
    //    JSON_Object* root_object = ClockAuction::JsonUtils::openJsonObject(NULL);
    //    msg.toWrappedStatusJsonObject(root_object, ec_, errorString_);
    //    ClockAuction::JsonUtils::closeJsonObject(root_object, &jsonString);
}

bool Contract::ErrorReport::isSuccess()
{
    if (ec_ == EC_SUCCESS)
    {
        return true;
    }
    return false;
}
