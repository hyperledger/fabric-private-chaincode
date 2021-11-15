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
class SignedApproval
{
public:
    std::string experimentId_;
    bool approved_;
    Id approverId_;
    ByteArray approvalBytes_;
    ByteArray signature_;
    ErrorReport er_;

    bool store(Contract::Storage& storage, const std::string& signedApprovalB64);
    bool retrieve(Contract::Storage& storage, std::string& signedApprovalB64);
};
}  // namespace Contract
