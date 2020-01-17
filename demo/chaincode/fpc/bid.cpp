/*
 * Copyright 2019 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "bid.h"
#include "auction-state.h"
#include "utils.h"

ClockAuction::Demand::Demand() {}

ClockAuction::Demand::Demand(uint32_t territoryId, uint32_t quantity, double price)
    : territoryId_(territoryId), quantity_(quantity), price_(price)
{
}

bool ClockAuction::Demand::fromJsonObject(const JSON_Object* root_object)
{
    {
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT,
            !json_object_has_value_of_type(root_object, "terId", JSONNumber));
        double d = json_object_get_number(root_object, "terId");
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, !ClockAuction::JsonUtils::isInteger(d));
        territoryId_ = (uint32_t)d;
    }
    {
        FAST_FAIL_CHECK(
            er_, EC_INVALID_INPUT, !json_object_has_value_of_type(root_object, "qty", JSONNumber));
        double d = json_object_get_number(root_object, "qty");
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, !ClockAuction::JsonUtils::isInteger(d));
        quantity_ = (uint32_t)d;
    }
    {
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT,
            !json_object_has_value_of_type(root_object, "price", JSONNumber));
        double d = json_object_get_number(root_object, "price");
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, !ClockAuction::JsonUtils::isInteger(d));
        price_ = (uint32_t)d;
    }
    return true;
}

void ClockAuction::Demand::toJsonObject(JSON_Object* root_object) const
{
    json_object_set_number(root_object, "terId", territoryId_);
    json_object_set_number(root_object, "qty", quantity_);
    json_object_set_number(root_object, "price", price_);
}

bool ClockAuction::Bid::fromJsonObject(const JSON_Object* root_object)
{
    {
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT,
            !json_object_has_value_of_type(root_object, "auctionId", JSONNumber));
        double d = json_object_get_number(root_object, "auctionId");
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, !ClockAuction::JsonUtils::isInteger(d));
        auctionId_ = (uint32_t)d;
    }
    {
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT,
            !json_object_has_value_of_type(root_object, "round", JSONNumber));
        double d = json_object_get_number(root_object, "round");
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, !ClockAuction::JsonUtils::isInteger(d));
        round_ = (uint32_t)d;
    }
    {
        JSON_Array* demand_array = json_object_get_array(root_object, "bids");
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, demand_array == 0);
        unsigned int bidsN = json_array_get_count(demand_array);
        // NOTE: the following check (bidsN == 0 && auctionId_ != 0) allows to have
        // no bids when auctionid is 0. This is useful to immediately store/retrieve
        // missing bids in the dynamic state for each bidder.
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, bidsN == 0 && auctionId_ != 0);
        for (unsigned int i = 0; i < bidsN; i++)
        {
            JSON_Object* o = json_array_get_object(demand_array, i);
            Demand d;
            FAST_FAIL_CHECK_EX(er_, &d.er_, EC_INVALID_INPUT, !d.fromJsonObject(o));
            demands_.push_back(d);
        }
    }
    return true;
}

void ClockAuction::Bid::toJsonObject(JSON_Object* root_object) const
{
    json_object_set_number(root_object, "auctionId", auctionId_);
    json_object_set_number(root_object, "round", round_);
    json_object_set_value(root_object, "bids", json_value_init_array());
    JSON_Array* demand_array = json_object_get_array(root_object, "bids");
    for (unsigned int i = 0; i < demands_.size(); i++)
    {
        JSON_Value* v = json_value_init_object();
        JSON_Object* o = json_value_get_object(v);
        demands_[i].toJsonObject(o);
        json_array_append_value(demand_array, v);
    }
}

uint32_t ClockAuction::Bid::sumQuantityDemands() const
{
    uint32_t total = 0;
    for (unsigned int i = 0; i < demands_.size(); i++)
    {
        total += demands_[i].quantity_;
    }
    return total;
}

std::vector<uint32_t> ClockAuction::Bid::getDemandedTerritoryIds() const
{
    std::vector<uint32_t> demandedTerritoryIds;
    for (unsigned int i = 0; i < demands_.size(); i++)
    {
        demandedTerritoryIds.push_back(demands_[i].territoryId_);
    }
    return demandedTerritoryIds;
}

bool ClockAuction::compareByDecreasingPricePoint(const enqueuedBid& b1, const enqueuedBid& b2)
{
    if (b1.pricePoint > b2.pricePoint)
        return true;
    if (b1.pricePoint < b2.pricePoint)
        return false;
    if (b1.randomValue > b2.randomValue)
        return true;
    return false;
}

bool ClockAuction::compareByIncreasingPricePoint(const enqueuedBid& b1, const enqueuedBid& b2)
{
    return !compareByDecreasingPricePoint(b1, b2);
}
