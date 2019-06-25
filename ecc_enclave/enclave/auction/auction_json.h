/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#pragma once

#include <stdbool.h>
#include <stdint.h>
#include <string>

typedef struct auction
{
    std::string name;
    bool is_open;
} auction_t;

typedef struct bid
{
    std::string bidder_name;
    int value;
} bid_t;

int unmarshal_auction(auction_t* auction, const char* json_bytes, uint32_t json_len);
int unmarshal_bid(bid_t* bids, const char* json_bytes, uint32_t json_len);
std::string marshal_auction(auction_t* auction);
std::string marshal_bid(bid_t* bid);
