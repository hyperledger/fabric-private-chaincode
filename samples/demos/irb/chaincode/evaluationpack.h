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
class EvaluationPack
{
public:
    std::string experimentId_;
    ErrorReport er_;

    bool build(Contract::Storage& storage, std::string& encryptedEvaluationPackB64);
};
}  // namespace Contract
