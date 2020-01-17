/*
 * Copyright 2019 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#pragma once

#include "common.h"
#include "error-codes.h"

namespace ClockAuction
{
class Demand
{
public:
    Demand();
    Demand(uint32_t territoryId, uint32_t quantity, double price);
    uint32_t territoryId_;
    uint32_t quantity_;
    double price_;

    ErrorReport er_;

    bool fromJsonObject(const JSON_Object* root_object);
    void toJsonObject(JSON_Object* root_object) const;
};

class Bid
{
public:
    uint32_t auctionId_;
    uint32_t round_;
    std::vector<Demand> demands_;
    ErrorReport er_;

    bool fromJsonObject(const JSON_Object* root_object);
    void toJsonObject(JSON_Object* root_object) const;

    //            bool isValid(const ClockAuction::StaticAuctionState& sState, const
    //            ClockAuction::DynamicAuctionState& dState);
    uint32_t sumQuantityDemands() const;
    std::vector<uint32_t> getDemandedTerritoryIds() const;
};

class enqueuedBid
{
public:
    uint32_t bidderIndex;
    uint32_t territoryIndex;
    Demand demand;
    uint32_t pricePoint;
    uint32_t randomValue;
};
bool compareByDecreasingPricePoint(const enqueuedBid& b1, const enqueuedBid& b2);
bool compareByIncreasingPricePoint(const enqueuedBid& b1, const enqueuedBid& b2);
}  // namespace ClockAuction
