/*
 * Copyright 2019 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#pragma once

#include "common.h"
#include "error-codes.h"
#include "principal.h"

namespace ClockAuction
{
class Bidder
{
private:
    uint32_t id_;
    std::string displayName_;
    Principal principal_;

public:
    bool toJsonObject(JSON_Object* root_object) const;
    bool fromJsonObject(const JSON_Object* root_object);
    ErrorReport er_;

    bool matchPrincipal(const Principal& p) const;
    uint32_t getId() const;
    const Principal getPrincipal() const;
};
}  // namespace ClockAuction
