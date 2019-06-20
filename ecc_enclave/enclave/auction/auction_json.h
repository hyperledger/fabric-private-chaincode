/*
 * Copyright IBM Corp. 2018 All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
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
