/*
 * Copyright 2019 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "echo_cc.h"
#include "shim.h"

#include <numeric>
#include <vector>

#define MAX_VALUE_SIZE 1024

// implements chaincode logic for invoke
int invoke(
    uint8_t* response, uint32_t max_response_len, uint32_t* actual_response_len, shim_ctx_ptr_t ctx)
{
    LOG_DEBUG("EchoCC: +++ Executing chaincode invocation +++");

    std::vector<std::string> argss;
    // parse json args
    get_string_args(argss, ctx);

    LOG_DEBUG("EchoCC: Args: %s",
        (argss.size() < 1
                ? "(none)"
                : std::accumulate(std::next(argss.begin()), argss.end(), argss[0],
                      [](std::string a, std::string b) { return (a + std::string(", ") + b); })
                      .c_str()));

    echo_t echo;
    echo.echo_string = argss[0];

    std::string result = echo.echo_string;

    // check that result fits into response
    int neededSize = result.size();
    if (max_response_len < neededSize)
    {
        // ouch error
        LOG_DEBUG("EchoCC: Response buffer too small");
        *actual_response_len = 0;
        return -1;
    }
    memcpy(response, result.c_str(), neededSize);
    *actual_response_len = neededSize;
    LOG_DEBUG("EchoCC: Response: %s", result.c_str());

    LOG_DEBUG("EchoCC: +++ Executing done +++");
    return 0;
}
