/*
 * Copyright 2019 Intel Corporation
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "dispatcher.h"
#include "logging.h"
#include "shim.h"

int invoke(
    uint8_t* response, uint32_t max_response_len, uint32_t* actual_response_len, shim_ctx_ptr_t ctx)
{
    int ret = -1;
    std::string functionName;
    std::vector<std::string> functionParameters;

    get_func_and_params(functionName, functionParameters, ctx);

    Contract::Dispatcher dispatcher(
        functionName, functionParameters, response, max_response_len, actual_response_len, ctx);
    if (!dispatcher.errorReport_.isSuccess())
    {
        LOG_ERROR("Execution failed.");
        ret = -1;
    }
    else
    {
        LOG_DEBUG("Execution successful");
        ret = 0;
    }

    // double check that the response has been filled
    if (*actual_response_len == 0)
    {
        LOG_ERROR("Response length is zero");
        ret = -1;
    }

    return ret;
}
