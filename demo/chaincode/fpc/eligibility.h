/*
 * Copyright 2019 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#pragma once

#include "common.h"
#include "error-codes.h"

namespace ClockAuction
{
class Eligibility
{
private:
    uint32_t bidderId_;
    uint32_t number_;

public:
    bool toJsonObject(JSON_Object* root_object) const;
    bool fromJsonObject(const JSON_Object* root_object);
    ErrorReport er_;

    bool matchBidderId(uint32_t bidderId) const;
    uint32_t getNumber() const;
};
}  // namespace ClockAuction
