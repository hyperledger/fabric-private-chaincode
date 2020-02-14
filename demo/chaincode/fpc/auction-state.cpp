/*
 * Copyright 2019 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "auction-state.h"
#include "common.h"
#include "utils.h"

const char* AUCTION_STATE_STRINGS[MAX_STATE_INDEX] = {
    "undefined", "clock", "assign", "failed_fsr", "done"};
const static std::map<std::string, auction_state_e> stateMap_ = {
    {AUCTION_STATE_STRINGS[STATE_UNDEFINED], STATE_UNDEFINED},
    {AUCTION_STATE_STRINGS[CLOCK_PHASE], CLOCK_PHASE},
    {AUCTION_STATE_STRINGS[ASSIGNMENT_PHASE], ASSIGNMENT_PHASE},
    {AUCTION_STATE_STRINGS[FSR_FAILED], FSR_FAILED}, {AUCTION_STATE_STRINGS[CLOSED], CLOSED}};

bool ClockAuction::StaticAuctionState::toJsonObject(JSON_Object* root_object) const
{
    {
        JSON_Object* o = ClockAuction::JsonUtils::openJsonObject(NULL);
        owner_.toJsonObject(o);
        json_object_set_value(root_object, "owner", json_object_get_wrapping_value(o));
        // do not close json object here -- it will be freed when parent object is closed
    }
    {
        json_object_set_string(root_object, "name", name_.c_str());
    }
    {
        json_object_set_value(root_object, "territories", json_value_init_array());
        JSON_Array* territory_array = json_object_get_array(root_object, "territories");
        for (unsigned int i = 0; i < territories_.size(); i++)
        {
            JSON_Value* v = json_value_init_object();
            JSON_Object* o = json_value_get_object(v);
            territories_[i].toJsonObject(o);
            json_array_append_value(territory_array, v);
        }
    }
    {
        json_object_set_value(root_object, "bidders", json_value_init_array());
        JSON_Array* bidder_array = json_object_get_array(root_object, "bidders");
        for (unsigned int i = 0; i < bidders_.size(); i++)
        {
            JSON_Value* v = json_value_init_object();
            JSON_Object* o = json_value_get_object(v);
            bidders_[i].toJsonObject(o);
            json_array_append_value(bidder_array, v);
        }
    }
    {
        json_object_set_value(root_object, "initialEligibilities", json_value_init_array());
        JSON_Array* eligibility_array = json_object_get_array(root_object, "initialEligibilities");
        for (unsigned int i = 0; i < initialEligibilities_.size(); i++)
        {
            JSON_Value* v = json_value_init_object();
            JSON_Object* o = json_value_get_object(v);
            initialEligibilities_[i].toJsonObject(o);
            json_array_append_value(eligibility_array, v);
        }
    }
    {
        json_object_set_number(
            root_object, "activityRequirementPercentage", activityRequirementPercentage_);
    }
    {
        json_object_set_number(
            root_object, "clockPriceIncrementPercentage", clockPriceIncrementPercentage_);
    }
    return true;
}

bool ClockAuction::StaticAuctionState::fromJsonObject(const JSON_Object* root_object)
{
    FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, root_object == NULL);

    {
        const char* str = json_object_get_string(root_object, "name");
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, str == 0);
        name_ = std::string(str);
    }
    {
        JSON_Array* territory_array = json_object_get_array(root_object, "territories");
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, territory_array == 0);
        unsigned int territoryN = json_array_get_count(territory_array);
        for (unsigned int i = 0; i < territoryN; i++)
        {
            JSON_Object* o = json_array_get_object(territory_array, i);
            Territory territory;
            FAST_FAIL_CHECK_EX(er_, &territory.er_, EC_INVALID_INPUT, !territory.fromJsonObject(o));
            territories_.push_back(territory);
        }
    }
    {
        JSON_Array* bidder_array = json_object_get_array(root_object, "bidders");
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, bidder_array == 0);
        unsigned int bidderN = json_array_get_count(bidder_array);
        for (unsigned int i = 0; i < bidderN; i++)
        {
            JSON_Object* o = json_array_get_object(bidder_array, i);
            Bidder bidder;
            FAST_FAIL_CHECK_EX(er_, &bidder.er_, EC_INVALID_INPUT, !bidder.fromJsonObject(o));
            bidders_.push_back(bidder);
        }
    }
    {
        JSON_Array* eligibility_array = json_object_get_array(root_object, "initialEligibilities");
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, eligibility_array == 0);
        unsigned int eligibilityN = json_array_get_count(eligibility_array);
        for (unsigned int i = 0; i < eligibilityN; i++)
        {
            JSON_Object* o = json_array_get_object(eligibility_array, i);
            Eligibility eligibility;
            FAST_FAIL_CHECK_EX(
                er_, &eligibility.er_, EC_INVALID_INPUT, !eligibility.fromJsonObject(o));
            initialEligibilities_.push_back(eligibility);
        }
    }
    {
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT,
            !json_object_has_value_of_type(
                root_object, "activityRequirementPercentage", JSONNumber));
        double d = json_object_get_number(root_object, "activityRequirementPercentage");
        activityRequirementPercentage_ = d;
        // TODO check integer
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, d < 0 || d > 100);
    }
    {
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT,
            !json_object_has_value_of_type(
                root_object, "clockPriceIncrementPercentage", JSONNumber));
        double d = json_object_get_number(root_object, "clockPriceIncrementPercentage");
        clockPriceIncrementPercentage_ = d;
        // TODO check integer
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, d < 0);
    }
    return true;
}

bool ClockAuction::StaticAuctionState::fromExtendedJsonObject(const JSON_Object* root_object)
{
    FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, root_object == NULL);
    {
        JSON_Value* v = json_object_get_value(root_object, "owner");
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, v == 0);
        JSON_Object* o = json_value_get_object(v);
        FAST_FAIL_CHECK_EX(er_, &owner_.er_, EC_INVALID_INPUT, !owner_.fromJsonObject(o));
    }
    return fromJsonObject(root_object);
}

bool ClockAuction::StaticAuctionState::checkValidity()
{
    FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, name_.length() == 0);
    FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, territories_.size() == 0);
    for (unsigned int j = 0; j < territories_.size(); j++)
    {
        std::vector<uint32_t> impairments = territories_[j].getChannelImpairments();
        for (unsigned int i = 0; i < impairments.size(); i++)
        {
            FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, impairments[i] > 100);
        }
    }
    FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, bidders_.size() == 0);
    FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, initialEligibilities_.size() == 0);
    FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, bidders_.size() != initialEligibilities_.size());
    FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, activityRequirementPercentage_ > 100);
    FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, clockPriceIncrementPercentage_ > 100);
    return true;
}

void ClockAuction::StaticAuctionState::setOwner(const Principal& p)
{
    owner_ = p;
}

ClockAuction::ErrorReport ClockAuction::StaticAuctionState::getErrorReport()
{
    return er_;
}

bool ClockAuction::StaticAuctionState::isTerritoryIdValid(uint32_t checkId)
{
    for (unsigned int i = 0; i < territories_.size(); i++)
    {
        if (checkId == territories_[i].getTerritoryId())
        {
            return true;
        }
    }
    return false;
}

const ClockAuction::Territory* ClockAuction::StaticAuctionState::getTerritory(
    uint32_t territoryId) const
{
    for (auto it = territories_.begin(); it != territories_.end(); ++it)
    {
        if (it->getTerritoryId() == territoryId)
        {
            return &(*it);
        }
    }
    return NULL;
}

int32_t ClockAuction::StaticAuctionState::getTerritoryIndex(uint32_t territoryId) const
{
    for (unsigned int i = 0; i < territories_.size(); i++)
    {
        if (territories_[i].getTerritoryId() == territoryId)
        {
            return i;
        }
    }
    return -1;
}

std::vector<uint32_t> ClockAuction::StaticAuctionState::getTerritoryIds() const
{
    std::vector<uint32_t> territoryIds;
    for (unsigned int i = 0; i < territories_.size(); i++)
    {
        territoryIds.push_back(territories_[i].getTerritoryId());
    }
    return territoryIds;
}

std::vector<uint32_t> ClockAuction::StaticAuctionState::getSupply() const
{
    std::vector<uint32_t> supply;
    for (unsigned int i = 0; i < territories_.size(); i++)
    {
        supply.push_back(territories_[i].numberOfChannels());
    }
    return supply;
}

std::vector<bool> ClockAuction::StaticAuctionState::getHighDemandVector() const
{
    std::vector<bool> highDemand;
    for (unsigned int i = 0; i < territories_.size(); i++)
    {
        highDemand.push_back(territories_[i].isHighDemand());
    }
    return highDemand;
}

bool ClockAuction::StaticAuctionState::isPrincipalOwner(const Principal& p) const
{
    return p == owner_;
}

bool ClockAuction::StaticAuctionState::isPrincipalBidder(const Principal& p)
{
    return fromPrincipalToBidderId(p) != 0;
}

uint32_t ClockAuction::StaticAuctionState::fromPrincipalToBidderId(const Principal& p) const
{
    for (unsigned int i = 0; i < bidders_.size(); i++)
    {
        if (bidders_[i].matchPrincipal(p))
        {
            return bidders_[i].getId();
        }
    }
    return 0;  // no bidder id
}

int32_t ClockAuction::StaticAuctionState::fromBidderIdToBidderIndex(uint32_t bidderId) const
{
    for (unsigned int i = 0; i < bidders_.size(); i++)
    {
        if (bidders_[i].getId() == bidderId)
        {
            return i;
        }
    }
    return -1;  // no bidder id/index
}

uint32_t ClockAuction::StaticAuctionState::fromBidderIndexToBidderId(uint32_t bidderIndex) const
{
    return bidders_[bidderIndex].getId();
}

const ClockAuction::Principal ClockAuction::StaticAuctionState::fromBidderIdToPrincipal(
    uint32_t bidderId) const
{
    for (unsigned int i = 0; i < bidders_.size(); i++)
    {
        if (bidders_[i].getId() == bidderId)
        {
            return bidders_[i].getPrincipal();
        }
    }
}

uint32_t ClockAuction::StaticAuctionState::getEligibilityNumber(uint32_t bidderId) const
{
    for (unsigned int i = 0; i < initialEligibilities_.size(); i++)
    {
        if (initialEligibilities_[i].matchBidderId(bidderId))
        {
            return initialEligibilities_[i].getNumber();
        }
    }
}

std::vector<double> ClockAuction::StaticAuctionState::getInitialPrices() const
{
    std::vector<double> initialPrices;
    for (unsigned int i = 0; i < territories_.size(); i++)
    {
        initialPrices.push_back(territories_[i].getMinPrice());
    }
    return initialPrices;
}

std::vector<uint32_t> ClockAuction::StaticAuctionState::getInitialEligibilities() const
{
    std::vector<uint32_t> initialEligibilities;
    for (unsigned int i = 0; i < initialEligibilities_.size(); i++)
    {
        initialEligibilities.push_back(initialEligibilities_[i].getNumber());
    }
    return initialEligibilities;
}

uint32_t ClockAuction::StaticAuctionState::getActivityRequirementPercentage() const
{
    return activityRequirementPercentage_;
}

uint32_t ClockAuction::StaticAuctionState::getClockPriceIncrementPercentage() const
{
    return clockPriceIncrementPercentage_;
}

uint32_t ClockAuction::StaticAuctionState::getBiddersN() const
{
    return bidders_.size();
}

uint32_t ClockAuction::StaticAuctionState::getTerritoryN() const
{
    return territories_.size();
}

/*
 * *************************************
 * DYNAMIC STATE ***********************
 * *************************************
 */

ClockAuction::DynamicAuctionState::DynamicAuctionState(shim_ctx_ptr_t ctx)
{
    identifySubmitter(ctx);
}

void ClockAuction::DynamicAuctionState::initialize(auction_state_e auctionState,
    uint32_t clockRound,
    bool roundActive,
    StaticAuctionState& staticAuctionState)
{
    staticAuctionState.setOwner(submitterPrincipal_);
    auctionState_ = auctionState;
    clockRound_ = clockRound;
    roundActive_ = roundActive;

    {  // initialize posted prices
        std::vector<double> initialPrices = staticAuctionState.getInitialPrices();
        postedPrice_.resize(2);
        postedPrice_[0] = initialPrices;               // posted price of round 0
        postedPrice_[1].resize(initialPrices.size());  // zero posted price vector for round 1
    }
    {                           // initialize clock prices
        clockPrice_.resize(1);  // dummy 0th
        clockPrice_.push_back(
            postedPrice_[0]);  // copy posted prices of round 0 in clock prices of round 1
        for (unsigned int i = 0; i < clockPrice_[clockRound_].size(); i++)
        {
            clockPrice_[clockRound_][i] *=
                (1 + ((double)staticAuctionState.getClockPriceIncrementPercentage() / 100));
        }
    }
    {                            // initialize eligibilities
        eligibility_.resize(1);  // dummy 0th
        eligibility_.push_back(staticAuctionState.getInitialEligibilities());
    }
    {
        // initialize clock bids
        clockBids_.resize(1);  // dummy 0th
    }
    {
        // initialize processed licenses
        processedLicenses_.resize(1);  // dummy 0th
    }
    {
        // initialized excess demand
        excessDemand_.resize(1);  // dummy 0th
    }
}

void ClockAuction::DynamicAuctionState::identifySubmitter(shim_ctx_ptr_t ctx)
{
    char mspId[1 << 12];
    char dn[1 << 12];
    get_creator_name(mspId, sizeof(mspId), dn, sizeof(dn), ctx);
    submitterPrincipal_ = Principal(mspId, dn);
    LOG_INFO("Identified submitter: mspid %s dn %s", mspId, dn);
}

bool ClockAuction::DynamicAuctionState::toJsonObject(JSON_Object* root_object) const
{
    json_object_set_string(root_object, "state", AUCTION_STATE_STRINGS[auctionState_]);
    json_object_set_number(root_object, "clockRound", clockRound_);
    json_object_set_boolean(root_object, "roundActive", (int)roundActive_);
    {
        json_object_set_value(root_object, "postedPrice", json_value_init_array());
        JSON_Array* json_array = json_object_get_array(root_object, "postedPrice");
        for (unsigned int i = 0; i < postedPrice_.size(); i++)
        {
            JSON_Value* perround_v = json_value_init_array();
            JSON_Array* perround_array = json_value_get_array(perround_v);
            for (unsigned int j = 0; j < postedPrice_[i].size(); j++)
            {
                json_array_append_number(perround_array, postedPrice_[i][j]);
            }
            json_array_append_value(json_array, perround_v);
        }
    }
    {
        json_object_set_value(root_object, "clockPrice", json_value_init_array());
        JSON_Array* json_array = json_object_get_array(root_object, "clockPrice");
        for (unsigned int i = 0; i < clockPrice_.size(); i++)
        {
            JSON_Value* perround_v = json_value_init_array();
            JSON_Array* perround_array = json_value_get_array(perround_v);
            for (unsigned int j = 0; j < clockPrice_[i].size(); j++)
            {
                json_array_append_number(perround_array, clockPrice_[i][j]);
            }
            json_array_append_value(json_array, perround_v);
        }
    }
    {
        json_object_set_value(root_object, "eligibility", json_value_init_array());
        JSON_Array* json_array = json_object_get_array(root_object, "eligibility");
        for (unsigned int i = 0; i < eligibility_.size(); i++)
        {
            JSON_Value* perround_v = json_value_init_array();
            JSON_Array* perround_array = json_value_get_array(perround_v);
            for (unsigned int j = 0; j < eligibility_[i].size(); j++)
            {
                json_array_append_number(perround_array, eligibility_[i][j]);
            }
            json_array_append_value(json_array, perround_v);
        }
    }
    {
        json_object_set_value(root_object, "clockBids", json_value_init_array());
        JSON_Array* json_array = json_object_get_array(root_object, "clockBids");
        for (unsigned int i = 0; i < clockBids_.size(); i++)
        {
            JSON_Value* perround_v = json_value_init_array();
            JSON_Array* perround_array = json_value_get_array(perround_v);
            for (unsigned int j = 0; j < clockBids_[i].size(); j++)
            {
                JSON_Value* v = json_value_init_object();
                JSON_Object* o = json_value_get_object(v);
                clockBids_[i][j].toJsonObject(o);
                json_array_append_value(perround_array, v);
            }
            json_array_append_value(json_array, perround_v);
        }
    }
    {
        json_object_set_value(root_object, "processedLicenses", json_value_init_array());
        JSON_Array* json_array = json_object_get_array(root_object, "processedLicenses");
        for (unsigned int i = 0; i < processedLicenses_.size(); i++)  // for each round
        {
            JSON_Value* perround_v = json_value_init_array();
            JSON_Array* perround_array = json_value_get_array(perround_v);
            for (unsigned int j = 0; j < processedLicenses_[i].size(); j++)  // for each bidder
            {
                JSON_Value* perbidder_v = json_value_init_array();
                JSON_Array* perbidder_array = json_value_get_array(perbidder_v);
                for (unsigned int k = 0; k < processedLicenses_[i][j].size();
                     k++)  // for each territory
                {
                    json_array_append_number(perbidder_array, processedLicenses_[i][j][k]);
                }
                json_array_append_value(perround_array, perbidder_v);
            }
            json_array_append_value(json_array, perround_v);
        }
    }
    {
        json_object_set_value(root_object, "excessDemand", json_value_init_array());
        JSON_Array* json_array = json_object_get_array(root_object, "excessDemand");
        for (unsigned int i = 0; i < excessDemand_.size(); i++)
        {
            JSON_Value* perround_v = json_value_init_array();
            JSON_Array* perround_array = json_value_get_array(perround_v);
            for (unsigned int j = 0; j < excessDemand_[i].size(); j++)
            {
                json_array_append_number(perround_array, excessDemand_[i][j]);
            }
            json_array_append_value(json_array, perround_v);
        }
    }
    {
        json_object_set_value(root_object, "winAssign", json_value_init_array());
        JSON_Array* json_array = json_object_get_array(root_object, "winAssign");
        for (unsigned int i = 0; i < winAssign_.size(); i++)  // for each territory
        {
            JSON_Value* perterritory_v = json_value_init_array();
            JSON_Array* perterritory_array = json_value_get_array(perterritory_v);
            for (unsigned int j = 0; j < winAssign_[i].size(); j++)  // for each channel
            {
                json_array_append_number(perterritory_array, winAssign_[i][j]);
            }
            json_array_append_value(json_array, perterritory_v);
        }
    }
    {
        json_object_set_value(root_object, "assignCost", json_value_init_array());
        JSON_Array* json_array = json_object_get_array(root_object, "assignCost");
        for (unsigned int i = 0; i < assignCost_.size(); i++)  // for each territory
        {
            JSON_Value* perterritory_v = json_value_init_array();
            JSON_Array* perterritory_array = json_value_get_array(perterritory_v);
            for (unsigned int j = 0; j < assignCost_[i].size(); j++)  // for each bidder
            {
                json_array_append_number(perterritory_array, assignCost_[i][j]);
            }
            json_array_append_value(json_array, perterritory_v);
        }
    }
    {
        json_object_set_value(root_object, "channelPrice", json_value_init_array());
        JSON_Array* json_array = json_object_get_array(root_object, "channelPrice");
        for (unsigned int i = 0; i < channelPrice_.size(); i++)  // for each territory
        {
            JSON_Value* perterritory_v = json_value_init_array();
            JSON_Array* perterritory_array = json_value_get_array(perterritory_v);
            for (unsigned int j = 0; j < channelPrice_[i].size(); j++)  // for each channel
            {
                json_array_append_number(perterritory_array, channelPrice_[i][j]);
            }
            json_array_append_value(json_array, perterritory_v);
        }
    }
    return true;
}

bool ClockAuction::DynamicAuctionState::fromJsonObject(const JSON_Object* root_object)
{
    {
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT,
            !json_object_has_value_of_type(root_object, "state", JSONString));
        const char* s = json_object_get_string(root_object, "state");
        auto fIter = stateMap_.find(s);
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, fIter == stateMap_.end());
        auctionState_ = fIter->second;
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, auctionState_ == 0);
    }
    {
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT,
            !json_object_has_value_of_type(root_object, "clockRound", JSONNumber));
        double d = json_object_get_number(root_object, "clockRound");
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, !ClockAuction::JsonUtils::isInteger(d));
        clockRound_ = (uint32_t)d;
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, clockRound_ == 0);
    }
    {
        int b = json_object_get_boolean(root_object, "roundActive");
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, roundActive_ == -1);
        roundActive_ = b;
    }
    {
        JSON_Array* postedprice_array = json_object_get_array(root_object, "postedPrice");
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, postedprice_array == 0);
        unsigned int postedPriceN = json_array_get_count(postedprice_array);
        for (unsigned int i = 0; i < postedPriceN; i++)
        {
            JSON_Value* perround_v = json_array_get_value(postedprice_array, i);
            JSON_Array* perround_array = json_value_get_array(perround_v);
            unsigned int roundPostedPriceN = json_array_get_count(perround_array);
            std::vector<double> roundPostedPrice;
            for (unsigned int j = 0; j < roundPostedPriceN; j++)
            {
                roundPostedPrice.push_back(json_array_get_number(perround_array, j));
            }
            postedPrice_.push_back(roundPostedPrice);
        }
    }
    {
        JSON_Array* clockprice_array = json_object_get_array(root_object, "clockPrice");
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, clockprice_array == 0);
        unsigned int clockPriceN = json_array_get_count(clockprice_array);
        for (unsigned int i = 0; i < clockPriceN; i++)
        {
            JSON_Value* perround_v = json_array_get_value(clockprice_array, i);
            JSON_Array* perround_array = json_value_get_array(perround_v);
            unsigned int roundClockPriceN = json_array_get_count(perround_array);
            std::vector<double> roundClockPrice;
            for (unsigned int j = 0; j < roundClockPriceN; j++)
            {
                roundClockPrice.push_back(json_array_get_number(perround_array, j));
            }
            clockPrice_.push_back(roundClockPrice);
        }
    }
    {
        JSON_Array* eligibility_array = json_object_get_array(root_object, "eligibility");
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, eligibility_array == 0);
        unsigned int eligibilityN = json_array_get_count(eligibility_array);
        for (unsigned int i = 0; i < eligibilityN; i++)
        {
            JSON_Value* perround_v = json_array_get_value(eligibility_array, i);
            JSON_Array* perround_array = json_value_get_array(perround_v);
            unsigned int roundEligibilityN = json_array_get_count(perround_array);
            std::vector<uint32_t> roundEligibility;
            for (unsigned int j = 0; j < roundEligibilityN; j++)
            {
                roundEligibility.push_back(json_array_get_number(perround_array, j));
            }
            eligibility_.push_back(roundEligibility);
        }
    }
    {
        JSON_Array* json_array = json_object_get_array(root_object, "clockBids");
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, json_array == 0);
        unsigned int arrayN = json_array_get_count(json_array);
        for (unsigned int i = 0; i < arrayN; i++)
        {
            JSON_Value* perround_v = json_array_get_value(json_array, i);
            JSON_Array* perround_array = json_value_get_array(perround_v);
            unsigned int roundArrayN = json_array_get_count(perround_array);
            std::vector<Bid> roundClockBids;
            for (unsigned int j = 0; j < roundArrayN; j++)
            {
                JSON_Object* o = json_array_get_object(perround_array, j);
                Bid bid;
                FAST_FAIL_CHECK_EX(er_, &bid.er_, EC_INVALID_INPUT, !bid.fromJsonObject(o));
                roundClockBids.push_back(bid);
            }
            clockBids_.push_back(roundClockBids);
        }
    }
    {
        JSON_Array* json_array = json_object_get_array(root_object, "processedLicenses");
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, json_array == 0);
        unsigned int arrayN = json_array_get_count(json_array);
        for (unsigned int i = 0; i < arrayN; i++)  // for each round
        {
            JSON_Value* perround_v = json_array_get_value(json_array, i);
            JSON_Array* perround_array = json_value_get_array(perround_v);
            unsigned int roundArrayN = json_array_get_count(perround_array);
            std::vector<std::vector<uint32_t> > perBidderProcessedLicenses;
            for (unsigned int j = 0; j < roundArrayN; j++)  // for each bidder
            {
                JSON_Value* perbidder_v = json_array_get_value(perround_array, j);
                JSON_Array* perbidder_array = json_value_get_array(perbidder_v);
                unsigned int bidderArrayN = json_array_get_count(perbidder_array);
                std::vector<uint32_t> perTerritoryProcessedLicenses;
                for (unsigned int k = 0; k < bidderArrayN; k++)  // for each territory
                {
                    perTerritoryProcessedLicenses.push_back(
                        json_array_get_number(perbidder_array, k));
                }
                perBidderProcessedLicenses.push_back(perTerritoryProcessedLicenses);
            }
            processedLicenses_.push_back(perBidderProcessedLicenses);
        }
    }
    {
        JSON_Array* json_array = json_object_get_array(root_object, "excessDemand");
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, json_array == 0);
        unsigned int arrayN = json_array_get_count(json_array);
        for (unsigned int i = 0; i < arrayN; i++)
        {
            JSON_Value* perround_v = json_array_get_value(json_array, i);
            JSON_Array* perround_array = json_value_get_array(perround_v);
            unsigned int roundArrayN = json_array_get_count(perround_array);
            std::vector<int32_t> roundExcessDemand;
            for (unsigned int j = 0; j < roundArrayN; j++)
            {
                roundExcessDemand.push_back(json_array_get_number(perround_array, j));
            }
            excessDemand_.push_back(roundExcessDemand);
        }
    }
    {
        JSON_Array* json_array = json_object_get_array(root_object, "winAssign");
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, json_array == 0);
        unsigned int arrayN = json_array_get_count(json_array);
        for (unsigned int i = 0; i < arrayN; i++)  // for each territory
        {
            JSON_Value* perterritory_v = json_array_get_value(json_array, i);
            JSON_Array* perterritory_array = json_value_get_array(perterritory_v);
            unsigned int territoryArrayN = json_array_get_count(perterritory_array);
            std::vector<int32_t> perTerritoryWinAssign;
            for (unsigned int j = 0; j < territoryArrayN; j++)  // for each channel
            {
                perTerritoryWinAssign.push_back(json_array_get_number(perterritory_array, j));
            }
            winAssign_.push_back(perTerritoryWinAssign);
        }
    }
    {
        JSON_Array* json_array = json_object_get_array(root_object, "assignCost");
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, json_array == 0);
        unsigned int arrayN = json_array_get_count(json_array);
        for (unsigned int i = 0; i < arrayN; i++)  // for each territory
        {
            JSON_Value* perterritory_v = json_array_get_value(json_array, i);
            JSON_Array* perterritory_array = json_value_get_array(perterritory_v);
            unsigned int territoryArrayN = json_array_get_count(perterritory_array);
            std::vector<double> perTerritoryAssignCost;
            for (unsigned int j = 0; j < territoryArrayN; j++)  // for each bidder
            {
                perTerritoryAssignCost.push_back(json_array_get_number(perterritory_array, j));
            }
            assignCost_.push_back(perTerritoryAssignCost);
        }
    }
    {
        JSON_Array* json_array = json_object_get_array(root_object, "channelPrice");
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, json_array == 0);
        unsigned int arrayN = json_array_get_count(json_array);
        for (unsigned int i = 0; i < arrayN; i++)  // for each territory
        {
            JSON_Value* perterritory_v = json_array_get_value(json_array, i);
            JSON_Array* perterritory_array = json_value_get_array(perterritory_v);
            unsigned int territoryArrayN = json_array_get_count(perterritory_array);
            std::vector<double> perTerritoryChannelPrice;
            for (unsigned int j = 0; j < territoryArrayN; j++)  // for each channel
            {
                perTerritoryChannelPrice.push_back(json_array_get_number(perterritory_array, j));
            }
            channelPrice_.push_back(perTerritoryChannelPrice);
        }
    }
    return true;
}

bool ClockAuction::DynamicAuctionState::toRoundInfoJsonObject(
    JSON_Object* root_object, const ClockAuction::StaticAuctionState& sState, uint32_t round) const
{
    {
        json_object_set_value(root_object, "prices", json_value_init_array());
        JSON_Array* json_array = json_object_get_array(root_object, "prices");
        std::vector<uint32_t> terIds = sState.getTerritoryIds();
        for (unsigned int i = 0; i < terIds.size(); i++)
        {
            JSON_Value* v = json_value_init_object();
            JSON_Object* o = json_value_get_object(v);
            json_object_set_number(o, "terId", terIds[i]);
            json_object_set_number(o, "minPrice", postedPrice_[round - 1][i]);
            json_object_set_number(o, "clockPrice", clockPrice_[round][i]);

            json_array_append_value(json_array, v);
        }
    }
    json_object_set_boolean(
        root_object, "active", (round == getRound() ? (int)roundActive_ : (int)false));
    return true;
}

bool ClockAuction::DynamicAuctionState::toBidderRoundResultsJsonObject(
    JSON_Object* root_object, const ClockAuction::StaticAuctionState& sState, uint32_t round) const
{
    uint32_t bidderIndex =
        sState.fromBidderIdToBidderIndex(sState.fromPrincipalToBidderId(submitterPrincipal_));
    {
        json_object_set_value(root_object, "result", json_value_init_array());
        JSON_Array* json_array = json_object_get_array(root_object, "result");
        std::vector<uint32_t> terIds = sState.getTerritoryIds();
        for (unsigned int i = 0; i < terIds.size(); i++)
        {
            if (processedLicenses_[round][bidderIndex][i] == 0)
            {
                continue;
            }

            JSON_Value* v = json_value_init_object();
            JSON_Object* o = json_value_get_object(v);
            json_object_set_number(o, "terId", terIds[i]);
            json_object_set_number(o, "postedPrice", postedPrice_[round][i]);
            json_object_set_number(o, "excessDemand", excessDemand_[round][i]);
            json_object_set_number(
                o, "processedLicenses", processedLicenses_[round][bidderIndex][i]);

            json_array_append_value(json_array, v);
        }
    }
    if (!isLastClockRound(round))
    {
        // NOTE: since this is not the last clock round, there exists a next (round+1) round
        uint32_t elig = eligibility_[round + 1][bidderIndex];
        json_object_set_number(root_object, "futureEligibility", elig);
        uint32_t requiredFutureElig = elig * sState.getActivityRequirementPercentage() / 100;
        json_object_set_number(root_object, "requiredFutureEligibility", requiredFutureElig);
    }

    return true;
}

bool ClockAuction::DynamicAuctionState::toOwnerRoundResultsJsonObject(
    JSON_Object* root_object, const ClockAuction::StaticAuctionState& sState, uint32_t round) const
{
    {
        json_object_set_value(root_object, "result", json_value_init_array());
        JSON_Array* json_array = json_object_get_array(root_object, "result");
        std::vector<uint32_t> terIds = sState.getTerritoryIds();
        uint32_t biddersN = sState.getBiddersN();
        for (unsigned int i = 0; i < terIds.size(); i++)
        {
            uint32_t activeBidders = 0;
            for (unsigned int j = 0; j < biddersN; j++)
            {
                activeBidders += (processedLicenses_[round][j][i] > 0 ? 1 : 0);
            }

            JSON_Value* v = json_value_init_object();
            JSON_Object* o = json_value_get_object(v);
            json_object_set_number(o, "terId", terIds[i]);
            json_object_set_number(o, "postedPrice", postedPrice_[round][i]);
            json_object_set_number(o, "excessDemand", excessDemand_[round][i]);
            json_object_set_number(o, "activeBidders", activeBidders);

            json_array_append_value(json_array, v);
        }
    }
    return true;
}

bool ClockAuction::DynamicAuctionState::toAssignmentResultsJsonObject(
    JSON_Object* root_object, const ClockAuction::StaticAuctionState& sState) const
{
    {
        json_object_set_value(root_object, "result", json_value_init_array());
        JSON_Array* json_array = json_object_get_array(root_object, "result");
        std::vector<uint32_t> terIds = sState.getTerritoryIds();
        uint32_t biddersN = sState.getBiddersN();
        for (unsigned int i = 0; i < biddersN; i++)
        {
            JSON_Value* v = json_value_init_object();
            JSON_Object* o = json_value_get_object(v);
            json_object_set_number(o, "bidderId", sState.fromBidderIndexToBidderId(i));
            json_object_set_value(o, "assignment", json_value_init_array());
            JSON_Array* assignment_json_array = json_object_get_array(o, "assignment");
            for (unsigned int j = 0; j < terIds.size(); j++)
            {
                if (processedLicenses_[getRound()][i][j] == 0)
                {
                    continue;
                }
                JSON_Value* assignment_v = json_value_init_object();
                JSON_Object* assignment_o = json_value_get_object(assignment_v);
                json_object_set_number(assignment_o, "territoryId", terIds[j]);
                json_object_set_number(assignment_o, "assignCost", assignCost_[j][i]);
                json_object_set_value(assignment_o, "channels", json_value_init_array());
                JSON_Array* channels_json_array = json_object_get_array(assignment_o, "channels");
                const Territory* t = sState.getTerritory(terIds[j]);
                std::vector<uint32_t> channelIds = t->getChannelIds();
                for (unsigned k = 0; k < channelIds.size(); k++)
                {
                    if (winAssign_[j][k] != i)
                    {
                        continue;
                    }

                    JSON_Value* channel_v = json_value_init_object();
                    JSON_Object* channel_o = json_value_get_object(channel_v);
                    json_object_set_number(channel_o, "channelId", channelIds[k]);
                    json_object_set_number(channel_o, "price", channelPrice_[j][k]);

                    json_array_append_value(channels_json_array, channel_v);
                }

                json_array_append_value(assignment_json_array, assignment_v);
            }

            json_array_append_value(json_array, v);
        }
    }
    return true;
}

bool ClockAuction::DynamicAuctionState::isRoundActive() const
{
    return roundActive_;
}

void ClockAuction::DynamicAuctionState::startRound(const ClockAuction::StaticAuctionState& sState)
{
    if (isStateClockPhase())
    {
        // add another vector of bids, which will be filled/updated as bidder submit bids
        if (clockBids_.size() < clockRound_ + 1)
        {
            clockBids_.resize(clockBids_.size() + 1);
            clockBids_[clockRound_].resize(sState.getBiddersN());
        }
    }

    roundActive_ = true;
}

void ClockAuction::DynamicAuctionState::endRound()
{
    roundActive_ = false;
}

void ClockAuction::DynamicAuctionState::endRoundAndAdvance()
{
    endRound();
    clockRound_++;
}

bool ClockAuction::DynamicAuctionState::isStateClockPhase() const
{
    return auctionState_ == CLOCK_PHASE;
}

bool ClockAuction::DynamicAuctionState::isStateAssignmentPhase() const
{
    return auctionState_ == ASSIGNMENT_PHASE;
}

bool ClockAuction::DynamicAuctionState::isStateClosedPhase() const
{
    return auctionState_ == CLOSED;
}

uint32_t ClockAuction::DynamicAuctionState::getRound() const
{
    return clockRound_;
}

bool ClockAuction::DynamicAuctionState::isLastClockRound(uint32_t round) const
{
    // last clock round is: the stored round when the auction is in assignment phase or closed
    return round == getRound() && (isStateAssignmentPhase() || isStateClosedPhase());
}

void ClockAuction::DynamicAuctionState::closeAuctionState()
{
    auctionState_ = CLOSED;
}

const ClockAuction::Principal ClockAuction::DynamicAuctionState::getSubmitter() const
{
    return submitterPrincipal_;
}

bool ClockAuction::DynamicAuctionState::isValidBid(const StaticAuctionState& sState, const Bid& bid)
{
    FAST_FAIL_CHECK(er_, EC_UNRECOGNIZED_SUBMITTER, !isValidBidder(sState));

    std::set<uint32_t> duplicateIdTracker;
    FAST_FAIL_CHECK(er_, EC_ROUND_NOT_CURRENT, clockRound_ != bid.round_);
    FAST_FAIL_CHECK(er_, EC_ROUND_NOT_ACTIVE, !roundActive_);

    for (unsigned int i = 0; i < bid.demands_.size(); i++)
    {
        const Territory* pTerritory = sState.getTerritory(bid.demands_[i].territoryId_);
        int32_t tIndex = sState.getTerritoryIndex(bid.demands_[i].territoryId_);

        // check territory existence
        FAST_FAIL_CHECK(er_, EC_UNRECOGNIZED_TERRITORY, pTerritory == NULL || tIndex == -1);

        // check if entered territory is duplicate
        auto it = duplicateIdTracker.find(bid.demands_[i].territoryId_);
        FAST_FAIL_CHECK(er_, EC_DUPLICATE_TERRITORIES, it != duplicateIdTracker.end());
        duplicateIdTracker.insert(bid.demands_[i].territoryId_);

        // check demand does not exceed supply
        FAST_FAIL_CHECK(
            er_, EC_TOO_MUCH_DEMAND, bid.demands_[i].quantity_ > pTerritory->numberOfChannels());

        if (clockRound_ == 1)
        {
            // check demand does not exceed eligibility
            uint32_t elig =
                sState.getEligibilityNumber(sState.fromPrincipalToBidderId(submitterPrincipal_));
            FAST_FAIL_CHECK(er_, EC_NOT_ENOUGH_ELIGIBILITY, bid.sumQuantityDemands() > elig);
            // no price check in first round
        }
        else
        {
            FAST_FAIL_CHECK(er_, EC_BELOW_POSTED_PRICE,
                bid.demands_[i].price_ < postedPrice_[clockRound_ - 1][tIndex]);
            FAST_FAIL_CHECK(er_, EC_ABOVE_CLOCK_PRICE,
                bid.demands_[i].price_ > clockPrice_[clockRound_][tIndex]);
        }
    }
    return true;
}

bool ClockAuction::DynamicAuctionState::isValidBidder(const StaticAuctionState& sState)
{
    uint32_t bidderId = sState.fromPrincipalToBidderId(submitterPrincipal_);
    return bidderId != 0;
}

bool ClockAuction::DynamicAuctionState::isValidOwner(const StaticAuctionState& sState)
{
    return sState.isPrincipalOwner(submitterPrincipal_);
}

void ClockAuction::DynamicAuctionState::storeBid(const StaticAuctionState& sState, const Bid& bid)
{
    uint32_t bidderId = sState.fromPrincipalToBidderId(submitterPrincipal_);
    int32_t bidderIndex = sState.fromBidderIdToBidderIndex(bidderId);
    clockBids_[clockRound_][bidderIndex] = bid;
}

void ClockAuction::DynamicAuctionState::fillMissingBids(
    const StaticAuctionState& sState, uint32_t auctionId)
{
    std::vector<uint32_t> territoryIds = sState.getTerritoryIds();
    sort(territoryIds.begin(), territoryIds.end());
    for (unsigned int i = 0; i < sState.getBiddersN(); i++)
    {
        // compute territory ids of missing bids
        std::vector<uint32_t> demandedTerritoryIds =
            clockBids_[clockRound_][i].getDemandedTerritoryIds();
        sort(demandedTerritoryIds.begin(), demandedTerritoryIds.end());
        std::vector<uint32_t> diff(sState.getTerritoryN());
        auto it = std::set_difference(territoryIds.begin(), territoryIds.end(),
            demandedTerritoryIds.begin(), demandedTerritoryIds.end(), diff.begin());
        diff.resize(it - diff.begin());

        // fill all fields in the available (possible empty) Bid
        clockBids_[clockRound_][i].auctionId_ = auctionId;
        clockBids_[clockRound_][i].round_ = clockRound_;
        for (unsigned int j = 0; j < diff.size(); j++)
        {
            uint32_t territoryIndex = sState.getTerritoryIndex(diff[j]);
            ClockAuction::Demand d(diff[j], 0, postedPrice_[clockRound_ - 1][territoryIndex]);
            clockBids_[clockRound_][i].demands_.push_back(d);
        }
    }
}
