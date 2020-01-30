/*
 * Copyright 2019 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#pragma once

#include "bid.h"
#include "bidder.h"
#include "common.h"
#include "eligibility.h"
#include "error-codes.h"
#include "principal.h"
#include "territory.h"

typedef enum
{
    STATE_UNDEFINED,
    CLOCK_PHASE,
    ASSIGNMENT_PHASE,
    FSR_FAILED,
    CLOSED,
    MAX_STATE_INDEX
} auction_state_e;

#define INITIAL_CLOCK_ROUND_NUMBER 1

namespace ClockAuction
{
class StaticAuctionState
{
private:
    Principal owner_;
    std::string name_;
    std::vector<Territory> territories_;
    std::vector<Bidder> bidders_;
    std::vector<Eligibility> initialEligibilities_;
    uint32_t activityRequirementPercentage_;
    uint32_t clockPriceIncrementPercentage_;

public:
    bool toJsonObject(JSON_Object* root_object) const;
    bool fromJsonObject(const JSON_Object* root_object);
    bool fromExtendedJsonObject(const JSON_Object* root_object);
    bool checkValidity();
    void setOwner(const Principal& p);
    ErrorReport getErrorReport();
    ErrorReport er_;

    bool isTerritoryIdValid(uint32_t checkId);
    const Territory* getTerritory(uint32_t territoryId) const;
    int32_t getTerritoryIndex(uint32_t territoryId) const;
    std::vector<uint32_t> getTerritoryIds() const;
    std::vector<bool> getHighDemandVector() const;
    std::vector<uint32_t> getSupply() const;

    bool isPrincipalOwner(const Principal& p) const;
    bool isPrincipalBidder(const Principal& p);
    uint32_t fromPrincipalToBidderId(const Principal& p) const;
    int32_t fromBidderIdToBidderIndex(uint32_t bidderId) const;
    uint32_t fromBidderIndexToBidderId(uint32_t bidderIndex) const;
    const Principal fromBidderIdToPrincipal(uint32_t bidderId) const;
    uint32_t getEligibilityNumber(uint32_t bidderId) const;

    std::vector<double> getInitialPrices() const;
    std::vector<uint32_t> getInitialEligibilities() const;
    uint32_t getActivityRequirementPercentage() const;
    uint32_t getClockPriceIncrementPercentage() const;
    uint32_t getBiddersN() const;
    uint32_t getTerritoryN() const;
};

class DynamicAuctionState
{
private:
    auction_state_e auctionState_;
    uint32_t clockRound_;
    bool roundActive_;

    Principal submitterPrincipal_;

    std::vector<std::vector<double> > postedPrice_;    // vector: [round][territory-index]  = price
    std::vector<std::vector<double> > clockPrice_;     // vector: [round][territory-index]  = price
    std::vector<std::vector<uint32_t> > eligibility_;  // vector: [round][bidder-index]     = number
    std::vector<std::vector<ClockAuction::Bid> >
        clockBids_;                                    // vector: [round][bidder-index]     = bid
    std::vector<std::vector<int32_t> > excessDemand_;  // vector: [round][territory-index]  = number
    std::vector<std::vector<std::vector<uint32_t> > >
        processedLicenses_;  // vector: [round][bidder-index][territory-index] = number
    std::vector<std::vector<int32_t> >
        winAssign_;  // vector: [territory-index][channel-index] = bidder-index (-1=unassigned)
    std::vector<std::vector<double> >
        assignCost_;  // vector: [territory-index][bidder-index] = price
    std::vector<std::vector<double> >
        channelPrice_;  // vector: [territory-index][channel-index] = price

    void identifySubmitter(shim_ctx_ptr_t ctx);

public:
    ErrorReport er_;
    DynamicAuctionState(shim_ctx_ptr_t ctx);
    void initialize(auction_state_e auctionState,
        uint32_t clockRound,
        bool roundActive,
        StaticAuctionState& staticAuctionState);
    bool toJsonObject(JSON_Object* root_object) const;
    bool fromJsonObject(const JSON_Object* root_object);

    bool toRoundInfoJsonObject(JSON_Object* root_object,
        const ClockAuction::StaticAuctionState& sState,
        uint32_t round) const;
    bool toBidderRoundResultsJsonObject(JSON_Object* root_object,
        const ClockAuction::StaticAuctionState& sState,
        uint32_t round) const;
    bool toOwnerRoundResultsJsonObject(JSON_Object* root_object,
        const ClockAuction::StaticAuctionState& sState,
        uint32_t round) const;
    bool toAssignmentResultsJsonObject(
        JSON_Object* root_object, const ClockAuction::StaticAuctionState& sState) const;

    bool isRoundActive() const;
    void startRound(const StaticAuctionState& sState);
    void endRound();
    void endRoundAndAdvance();
    bool isStateClockPhase() const;
    bool isStateAssignmentPhase() const;
    bool isStateClosedPhase() const;
    uint32_t getRound() const;
    bool isLastClockRound(uint32_t round) const;
    void closeAuctionState();

    const Principal getSubmitter() const;

    bool isValidBid(const StaticAuctionState& sState, const Bid& bid);
    bool isValidBidder(const StaticAuctionState& sState);
    bool isValidOwner(const StaticAuctionState& sState);
    void storeBid(const StaticAuctionState& sState, const Bid& bid);
    void fillMissingBids(const StaticAuctionState& sState, uint32_t auctionId);

    void processBidsPreamble();
    void processInitialRoundBids(StaticAuctionState& sState, uint32_t auctionId);
    void processRegularRoundBids(StaticAuctionState& sState, uint32_t auctionId);
    void processBidsPostamble(StaticAuctionState& sState);
    void processAssignmentRound(StaticAuctionState& sState);
};
}  // namespace ClockAuction
