/*
 * Copyright 2021 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#pragma once

#include "common.h"
#include "errors.h"
#include "id.h"
#include "storage.h"

namespace Contract
{
class Study
{
public:
    std::string studyId_;
    std::string metadata_;
    std::vector<Id> userIds_;
    ErrorReport er_;

    bool store(Contract::Storage& storage, const std::string& studyDetailsB64);
    bool retrieve(Contract::Storage& storage, std::string& studyDetailsB64);
};
}  // namespace Contract
