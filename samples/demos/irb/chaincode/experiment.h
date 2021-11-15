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
class Experiment
{
public:
    std::string studyId_;
    std::string experimentId_;
    Id workerId_;
    ErrorReport er_;

    Experiment();
    Experiment(std::string studyId, std::string experimentId, Contract::Id& workerId);

    bool store(Contract::Storage& storage, const std::string& experimentProposalB64);
    bool retrieve(Contract::Storage& storage, std::string& experimentProposalB64);
};
}  // namespace Contract
