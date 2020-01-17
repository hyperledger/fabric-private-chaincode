/*
 * Copyright 2019 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#pragma once

#include "common.h"

namespace ClockAuction
{
class JsonUtils
{
public:
    static JSON_Object* openJsonObject(const char* c_str);
    static void closeJsonObject(JSON_Object* root_object, std::string* string);
    static bool isInteger(double d);
};
}  // namespace ClockAuction
