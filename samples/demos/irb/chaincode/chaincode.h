/*
 * Copyright 2019 Intel Corporation
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#pragma once

#include <string>
#include "errors.h"
#include "shim.h"
#include "storage.h"

#define CONTRACT_API_PROTOTYPE(api_name) \
    bool api_name(                       \
        const std::string& inputString, std::string& outputString, Contract::ErrorReport& er)

namespace Contract
{
class ExperimentApprovalService
{
private:
    Contract::Storage storage_;

public:
    ExperimentApprovalService(shim_ctx_ptr_t ctx);
    ErrorReport er_;

    CONTRACT_API_PROTOTYPE(registerData);
    CONTRACT_API_PROTOTYPE(registerStudy);
    CONTRACT_API_PROTOTYPE(newExperiment);
    CONTRACT_API_PROTOTYPE(getExperimentProposal);
    CONTRACT_API_PROTOTYPE(approveExperiment);
    CONTRACT_API_PROTOTYPE(requestEvaluationPack);
};
}  // namespace Contract

typedef CONTRACT_API_PROTOTYPE((Contract::ExperimentApprovalService::*contractFunctionP));
