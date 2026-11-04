/*
 * Copyright 2019 Intel Corporation
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include <map>
#include <vector>
#include "chaincode.h"

namespace Contract
{
class Dispatcher
{
private:
    const std::string functionName_;
    const std::vector<std::string> functionParameters_;
    uint8_t* response_;
    const uint32_t max_response_len_;
    uint32_t* actual_response_len_;
    std::string responseString_;

    Contract::ExperimentApprovalService contract_;

public:
    Contract::ErrorReport errorReport_;

    Dispatcher(const std::string& functionName,
        const std::vector<std::string>& functionParameters,
        uint8_t* response,
        const uint32_t max_response_len,
        uint32_t* actual_response_len,
        shim_ctx_ptr_t ctx);
};
}  // namespace Contract
