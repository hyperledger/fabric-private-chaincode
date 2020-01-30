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
class Principal
{
private:
    // uint32_t id_;
    std::string mspId_;
    std::string dn_;
    // std::string name_;

public:
    Principal();
    Principal(std::string m, std::string d);
    bool toJsonObject(JSON_Object* root_object) const;
    bool fromJsonObject(const JSON_Object* root_object);
    std::string getDn() const;
    std::string getMspId() const;
    ErrorReport er_;
};
}  // namespace ClockAuction

bool operator==(const ClockAuction::Principal& p1, const ClockAuction::Principal& p2);
