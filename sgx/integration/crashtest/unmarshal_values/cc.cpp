#include <string>
#include "logging.h"
#include "shim.h"

const std::string SEP = ".";
const std::string PREFIX = SEP + "somePrefix" + SEP;

std::string store(std::string auction_name, std::string bidder_name, int value, shim_ctx_ptr_t ctx)
{
    std::string new_key(PREFIX + auction_name + SEP + bidder_name + SEP);
    put_public_state(new_key.c_str(), (uint8_t*)&value, sizeof(int), ctx);
    return "OK";
}

std::string retrieve(std::string auction_name, shim_ctx_ptr_t ctx)
{
    std::map<std::string, std::string> values;
    std::string bid_composite_key = PREFIX + auction_name + SEP;
    get_public_state_by_partial_composite_key(bid_composite_key.c_str(), values, ctx);
    return "STILL_ALIVE";
}

int invoke(
    uint8_t* response, uint32_t max_response_len, uint32_t* actual_response_len, shim_ctx_ptr_t ctx)
{
    LOG_DEBUG("[+] +++ Executing chaincode invocation +++");

    std::string function_name;
    std::vector<std::string> params;
    get_func_and_params(function_name, params, ctx);
    std::string result;

    if (function_name == "store")
        result = store(params[0], params[1], std::stoi(params[2]), ctx);
    else if (function_name == "retrieve")
        result = retrieve(params[0], ctx);

    int neededSize = result.size();
    if (max_response_len < neededSize)
    {
        LOG_DEBUG("[+] Response buffer too small");
        *actual_response_len = 0;
        return -1;
    }

    memcpy(response, result.c_str(), neededSize);
    *actual_response_len = neededSize;
    LOG_DEBUG("[+] Response: %s", result.c_str());
    LOG_DEBUG("[+] +++ Executing done +++");
    return 0;
}
