/*
 * Copyright 2019 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "echo_cc.h"
#include "logging.h"
#include "shim.h"

#include <vector>

#define MAX_VALUE_SIZE 1024

int init(const char* args,
    uint8_t* response,
    uint32_t max_response_len,
    uint32_t* actual_response_len,
    void* ctx)
{
    LOG_DEBUG("EchoCC: +++ Executing chaincode init +++");
    LOG_DEBUG("EchoCC: \tArgs (ignored): %s", args);

    *actual_response_len = 0;
    LOG_DEBUG("EchoCC: +++ Initialization done +++");
    return 0;
}

// implements chaincode logic for invoke
int invoke(const char* args,
    uint8_t* response,
    uint32_t max_response_len,
    uint32_t* actual_response_len,
    void* ctx)
{
    LOG_DEBUG("EchoCC: +++ Executing chaincode invocation +++");
    LOG_DEBUG("Args: %s", args);

    std::vector<std::string> argss;
    // parse json args
    unmarshal_args(argss, args);

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
