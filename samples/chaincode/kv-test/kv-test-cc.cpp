/*
 * Copyright 2019 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "base64.h"
#include "shim.h"

#include <numeric>
#include <vector>

#define MAX_VALUE_SIZE (1 << 16)

int invoke(
    uint8_t* response, uint32_t max_response_len, uint32_t* actual_response_len, shim_ctx_ptr_t ctx)
{
    LOG_DEBUG("KV-Test: +++ Executing chaincode invocation +++");

    std::string function_name;
    std::vector<std::string> params;
    std::string result;
    get_func_and_params(function_name, params, ctx);

    LOG_DEBUG("Call function: %s", function_name.c_str());

    if (function_name == "put_state")
    {
        if (params.size() != 2)
        {
            result = std::string("put_state needs 2 parameters: key and value");
        }
        else
        {
            if (params[1].length() > MAX_VALUE_SIZE)
            {
                result = std::string("max value size exceeded");
            }
            else
            {
                put_state(params[0].c_str(), (uint8_t*)params[1].c_str(), params[1].length(), ctx);
                result = std::string("OK");
            }
        }
    }
    else if (function_name == "get_state")
    {
        uint8_t value[MAX_VALUE_SIZE];
        uint32_t actual_value_len = 0;

        if (params.size() != 1)
        {
            result = std::string("get_state needs 1 parameter: key");
        }
        else
        {
            get_state(params[0].c_str(), value, sizeof(value), &actual_value_len, ctx);
            if (actual_value_len == 0)
                result = std::string("NOT FOUND");
            else
                result = std::string((const char*)value, actual_value_len);
        }
    }
    else if (function_name == "del_state")
    {
        if (params.size() != 1)
        {
            result = std::string("del_state needs 1 parameter: key");
        }
        else
        {
            del_state(params[0].c_str(), ctx);
            result = std::string("OK");
        }
    }
    else
    {
        result = std::string("BAD FUNCTION");
    }

    // test more fpc shim functions

    std::string channel_id;
    get_channel_id(channel_id, ctx);
    LOG_DEBUG("Channel id: %s", channel_id.c_str());

    std::string tx_id;
    get_tx_id(tx_id, ctx);
    LOG_DEBUG("Tx id: %s", tx_id.c_str());

    ByteArray signed_proposal;
    get_signed_proposal(signed_proposal, ctx);
    LOG_DEBUG("Signed proposal: %s",
        base64_encode((unsigned char*)signed_proposal.data(), signed_proposal.size()).c_str());

    ByteArray creator;
    get_creator(creator, ctx);
    LOG_DEBUG("Creator: %s", base64_encode((unsigned char*)creator.data(), creator.size()).c_str());

    char creator_msp_id[1024];
    char creator_name[1024];
    get_creator_name(
        creator_msp_id, sizeof(creator_msp_id), creator_name, sizeof(creator_name), ctx);
    LOG_DEBUG("Creator msp: %s", creator_msp_id);
    LOG_DEBUG("Creator name: %s", creator_name);

    // check that result fits into response
    int neededSize = result.size();
    if (max_response_len < neededSize)
    {
        // ouch error
        LOG_ERROR("Response buffer too small");
        *actual_response_len = 0;
        return -1;
    }
    memcpy(response, result.c_str(), neededSize);
    *actual_response_len = neededSize;
    LOG_DEBUG("Response: %s", result.c_str());

    LOG_DEBUG("KV-Test: +++ Executing done +++");
    return 0;
}
