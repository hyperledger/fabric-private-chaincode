/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "auction_cc.h"
#include "auction_json.h"
#include "shim.h"

#include <numeric>
#include <vector>

#define MAX_VALUE_SIZE 1024

#define OK "OK"
#define AUCTION_DRAW "DRAW"
#define AUCTION_NO_BIDS "NO_BIDS"
#define AUCTION_ALREADY_EXISTING "AUCTION_ALREADY_EXISTING"
#define AUCTION_NOT_EXISTING "AUCTION_NOT_EXISTING"
#define AUCTION_ALREADY_CLOSED "AUCTION_ALREADY_CLOSED"
#define AUCTION_STILL_OPEN "AUCTION_STILL_OPEN"

#define INITIALIZED_KEY "initialized"
#define AUCTION_HOUSE_NAME_KEY "auction_house_name"

const std::string SEP = ".";
const std::string PREFIX = SEP + "somePrefix" + SEP;

// implements chaincode logic for invoke
int invoke(
    uint8_t* response, uint32_t max_response_len, uint32_t* actual_response_len, shim_ctx_ptr_t ctx)
{
    bool _initialized;
    const char* _auction_house_name;
    char _auction_house_name_buf[128];

    uint32_t init_len = -1;
    get_state(INITIALIZED_KEY, (uint8_t*)&_initialized, sizeof(_initialized), &init_len, ctx);
    if ((init_len == 0) || !_initialized)
    {
        _initialized = false;
        _auction_house_name = "(uninitialized)";
    }
    else
    {
        uint32_t ahn_len = -1;
        get_state(AUCTION_HOUSE_NAME_KEY, (uint8_t*)_auction_house_name_buf,
            sizeof(_auction_house_name_buf) - 1, &ahn_len, ctx);
        if (ahn_len == 0)
        {
            _auction_house_name = "(uninitialized)";
        }
        else
        {
            _auction_house_name_buf[ahn_len + 1] = '\0';
            _auction_house_name = _auction_house_name_buf;
        }
    }

    LOG_DEBUG(
        "AuctionCC: +++ Executing '%s' auction chaincode invocation +++", _auction_house_name);

    std::string function_name;
    std::vector<std::string> params;
    get_func_and_params(function_name, params, ctx);

    LOG_DEBUG("AuctionCC: Function: %s, Params: %s", function_name.c_str(),
        (params.size() < 1
                ? "(none)"
                : std::accumulate(std::next(params.begin()), params.end(), params[0],
                      [](std::string a, std::string b) { return (a + std::string(", ") + b); })
                      .c_str()));

    std::string auction_name = params[0];
    std::string result;

    if (!_initialized && function_name != "init")
    {
        LOG_ERROR("AuctionCC: Auction not yet initialized / No re-initialized allowed");
        *actual_response_len = 0;
        return -1;
    }

    if (function_name == "init")
    {
        result = init_auction_house(params[0], ctx);
    }
    else if (function_name == "create")
    {
        result = auction_create(auction_name, ctx);
    }
    else if (function_name == "submit")
    {
        int value = std::stoi(params[2]);
        std::string bidder_name = params[1];
        // TODO: eventually replace bidder_name with get_creator_name but for now
        //  in our tests we have only one client, so leave passed bidder_name to
        //  allow for different bidders ...
        char real_bidder_name_msp_id[1024];
        char real_bidder_name_dn[1024];
        get_creator_name(real_bidder_name_msp_id, sizeof(real_bidder_name_msp_id),
            real_bidder_name_dn, sizeof(real_bidder_name_dn), ctx);
        LOG_INFO("AuctionCC: real bidder '(msp_id: %s, dn: %s)' masquerading as '%s'",
            real_bidder_name_msp_id, real_bidder_name_dn, bidder_name.c_str());

        result = auction_submit(auction_name, bidder_name, value, ctx);
    }
    else if (function_name == "close")
    {
        result = auction_close(auction_name, ctx);
    }
    else if (function_name == "eval")
    {
        result = auction_eval(auction_name, ctx);
    }
    else
    {
        // unknown function
        LOG_ERROR("AuctionCC: RECEIVED UNKOWN transaction");
        *actual_response_len = 0;
        return -1;
    }

    // check that result fits into response
    int neededSize = result.size();
    if (max_response_len < neededSize)
    {
        // ouch error
        LOG_ERROR("AuctionCC: Response buffer too small");
        *actual_response_len = 0;
        return -1;
    }

    // copy result to response
    memcpy(response, result.c_str(), neededSize);
    *actual_response_len = neededSize;
    LOG_DEBUG("AuctionCC: Response: %s", result.c_str());

    LOG_DEBUG("AuctionCC: +++ Executing done +++");
    return 0;
}

std::string init_auction_house(std::string auction_house_name, shim_ctx_ptr_t ctx)
{
    put_state(AUCTION_HOUSE_NAME_KEY, (uint8_t*)auction_house_name.c_str(),
        auction_house_name.size(), ctx);

    bool _initialized = true;
    put_state(INITIALIZED_KEY, (uint8_t*)&_initialized, sizeof(_initialized), ctx);

    return OK;
}

std::string auction_create(std::string auction_name, shim_ctx_ptr_t ctx)
{
    // check if auction already exists
    uint32_t auction_bytes_len = 0;
    uint8_t auction_bytes[MAX_VALUE_SIZE];
    get_state(auction_name.c_str(), auction_bytes, sizeof(auction_bytes), &auction_bytes_len, ctx);

    if (auction_bytes_len > 0)
    {
        // auction already exists
        LOG_DEBUG("AuctionCC: Auction already exists");
        return AUCTION_ALREADY_EXISTING;
    }

    // create new auction
    auction_t new_auction;
    new_auction.name = (char*)auction_name.c_str();
    new_auction.is_open = true;

    // convert to json string and store
    std::string json = marshal_auction(&new_auction);
    put_state(auction_name.c_str(), (uint8_t*)json.c_str(), json.size(), ctx);

    return OK;
}

std::string auction_submit(
    std::string auction_name, std::string bidder_name, int value, shim_ctx_ptr_t ctx)
{
    // check if auction already exists
    uint32_t auction_bytes_len = 0;
    uint8_t auction_bytes[MAX_VALUE_SIZE];
    get_state(auction_name.c_str(), auction_bytes, sizeof(auction_bytes), &auction_bytes_len, ctx);

    if (auction_bytes_len == 0)
    {
        LOG_DEBUG("AuctionCC: Auction does not exist");
        return AUCTION_NOT_EXISTING;
    }

    // get auction struct from json
    auction_t the_auction;
    unmarshal_auction(&the_auction, (const char*)auction_bytes, auction_bytes_len);

    if (!the_auction.is_open)
    {
        LOG_DEBUG("AuctionCC: Auction is already closed");
        return AUCTION_ALREADY_CLOSED;
    }

    // create composite key "auction_name.bidder_name"
    // if there is already a bid just overwrite it
    std::string new_key(PREFIX + auction_name + SEP + bidder_name + SEP);

    bid_t bid;
    bid.bidder_name = bidder_name;
    bid.value = value;

    // convert to json and store
    std::string json = marshal_bid(&bid);
    put_state(new_key.c_str(), (uint8_t*)json.c_str(), json.size(), ctx);

    return OK;
}

std::string auction_close(std::string auction_name, shim_ctx_ptr_t ctx)
{
    // check if auction already exists
    uint32_t auction_bytes_len = 0;
    uint8_t auction_bytes[MAX_VALUE_SIZE];
    get_state(auction_name.c_str(), auction_bytes, sizeof(auction_bytes), &auction_bytes_len, ctx);

    if (auction_bytes_len == 0)
    {
        LOG_DEBUG("AuctionCC: Auction does not exist");
        return AUCTION_NOT_EXISTING;
    }

    // get auction struct from json
    auction_t the_auction;
    unmarshal_auction(&the_auction, (const char*)auction_bytes, auction_bytes_len);

    if (!the_auction.is_open)
    {
        LOG_DEBUG("AuctionCC: Auction is already closed");
        return AUCTION_ALREADY_CLOSED;
    }

    // close auction
    the_auction.is_open = false;

    // converto to json and store
    std::string json = marshal_auction(&the_auction);
    put_state(auction_name.c_str(), (uint8_t*)json.c_str(), json.size(), ctx);

    return OK;
}

std::string auction_eval(std::string auction_name, shim_ctx_ptr_t ctx)
{
    // check if auction already exists
    uint32_t auction_bytes_len = 0;
    uint8_t auction_bytes[MAX_VALUE_SIZE];
    get_state(auction_name.c_str(), auction_bytes, sizeof(auction_bytes), &auction_bytes_len, ctx);

    if (auction_bytes_len == 0)
    {
        LOG_DEBUG("AuctionCC: Auction does not exist");
        return AUCTION_NOT_EXISTING;
    }

    // get auction struct from json
    auction_t the_auction;
    unmarshal_auction(&the_auction, (const char*)auction_bytes, auction_bytes_len);

    // check if auction is closed
    if (the_auction.is_open)
    {
        LOG_DEBUG("AuctionCC: Auction is still open");
        return AUCTION_STILL_OPEN;
    }

    // the result of the auction
    std::string auction_result;

    // get all bids
    std::string bid_composite_key = PREFIX + auction_name + SEP;
    std::map<std::string, std::string> values;
    get_state_by_partial_composite_key(bid_composite_key.c_str(), values, ctx);

    if (values.empty())
    {
        LOG_DEBUG("AuctionCC: No bids");
        auction_result = AUCTION_NO_BIDS;
    }
    else
    {
        // search highest bid
        bid_t winner;
        int high = -1;
        int draw = 0;

        LOG_DEBUG("AuctionCC: All concidered bids:");
        for (auto u : values)
        {
            bid_t b;
            unmarshal_bid(&b, u.second.c_str(), u.second.size());

            LOG_DEBUG("AuctionCC: \t%s value %d", b.bidder_name.c_str(), b.value);
            if (b.value > high)
            {
                draw = 0;
                high = b.value;
                winner = b;
            }
            else if (b.value == high)
            {
                draw = 1;
            }
        }

        if (draw != 1)
        {
            LOG_DEBUG("AuctionCC: Winner is: %s with %d", winner.bidder_name.c_str(), winner.value);
            auction_result = marshal_bid(&winner);
        }
        else
        {
            LOG_DEBUG("AuctionCC: DRAW");
            auction_result = AUCTION_DRAW;
        }
    }

    // publically store result ..
    std::string auction_result_key(auction_name + SEP + "outcome" + SEP);
    put_public_state(
        auction_result_key.c_str(), (uint8_t*)auction_result.c_str(), auction_result.size(), ctx);

    // .. but also return it to caller ...
    return auction_result;
}
