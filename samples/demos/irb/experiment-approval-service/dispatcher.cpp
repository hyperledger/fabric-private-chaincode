/*
 * Copyright 2019 Intel Corporation
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "dispatcher.h"
#include <string>
#include "chaincode.h"
#include "errors.h"

Contract::Dispatcher::Dispatcher(const std::string& functionName,
    const std::vector<std::string>& functionParameters,
    uint8_t* response,
    const uint32_t max_response_len,
    uint32_t* actual_response_len,
    shim_ctx_ptr_t ctx)
    : functionName_(functionName),
      functionParameters_(functionParameters),
      response_(response),
      max_response_len_(max_response_len),
      actual_response_len_(actual_response_len),
      contract_(ctx)
{
    static std::map<std::string, contractFunctionP> fMap_ = {
        {"registerData", &Contract::ExperimentApprovalService::registerData},
        {"registerStudy", &Contract::ExperimentApprovalService::registerStudy},
        {"newExperiment", &Contract::ExperimentApprovalService::newExperiment},
        {"getExperimentProposal", &Contract::ExperimentApprovalService::getExperimentProposal},
        {"approveExperiment", &Contract::ExperimentApprovalService::approveExperiment},
        {"requestEvaluationPack", &Contract::ExperimentApprovalService::requestEvaluationPack}};

    LOG_DEBUG("Try dispatch function %s with parameters %s", functionName_.c_str(),
        functionParameters[0].c_str());

    // Call function
    auto fIter = fMap_.find(functionName_);
    if (fIter == fMap_.end())
    {
        // No such function
        CUSTOM_ERROR_REPORT(errorReport_, EC_BAD_FUNCTION_NAME, "Chaincode function not found");
    }
    else
    {
        LOG_DEBUG("Chaincode function found, call it");
        (contract_.*(fIter->second))(functionParameters[0], responseString_, errorReport_);
        LOG_DEBUG("Response string: %s", responseString_.c_str());
    }

    // prepare response string
    if (responseString_.length() == 0 || !errorReport_.isSuccess())
    {
        // an error occurred: fill the response with the error/status message
        errorReport_.toStatusProtoString(responseString_);
        LOG_DEBUG("Error response string set: %s", responseString_.c_str());
    }

    if (responseString_.length() > max_response_len_)
    {
        LOG_ERROR("Response string too long to be output");
        CUSTOM_ERROR_REPORT(
            errorReport_, EC_SHORT_RESPONSE_BUFFER, "Response string too long to be output");
        errorReport_.toStatusProtoString(responseString_);
    }

    // write response string (if possible)
    *actual_response_len_ =
        (responseString_.length() > max_response_len_ ? 0 : responseString_.length());
    memcpy(response_, responseString_.c_str(), *actual_response_len_);
    LOG_DEBUG("Response written (length %u)", *actual_response_len_);
}
