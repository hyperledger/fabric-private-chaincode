/*
 * Copyright 2019 Intel Corporation
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "chaincode.h"
#include <string>
#include "errors.h"
#include "messages.h"

Contract::ExperimentApprovalService::ExperimentApprovalService(shim_ctx_ptr_t ctx) {}

CONTRACT_API_PROTOTYPE(Contract::ExperimentApprovalService::registerData)
{
    return true;
}

CONTRACT_API_PROTOTYPE(Contract::ExperimentApprovalService::registerStudy)
{
    return true;
}

CONTRACT_API_PROTOTYPE(Contract::ExperimentApprovalService::newExperiment)
{
    return true;
}

CONTRACT_API_PROTOTYPE(Contract::ExperimentApprovalService::getExperimentProposal)
{
    return true;
}

CONTRACT_API_PROTOTYPE(Contract::ExperimentApprovalService::approveExperiment)
{
    return true;
}

CONTRACT_API_PROTOTYPE(Contract::ExperimentApprovalService::requestEvaluationPack)
{
    return true;
}
