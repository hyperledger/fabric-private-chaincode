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

void Contract::ErrorReport::toStatusProtoString(std::string& outputString)
{
Contract:
    EASMessage m;
    m.toStatus(errorString_, ec_, outputString);
}

bool Contract::ErrorReport::isSuccess()
{
    if (ec_ == EC_SUCCESS)
    {
        return true;
    }
    return false;
}
