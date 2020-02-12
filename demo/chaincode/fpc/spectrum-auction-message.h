/*
 * Copyright 2019 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#pragma once

#include "auction-state.h"
#include "bid.h"
#include "common.h"
#include "error-codes.h"
#include "utils.h"

typedef struct
{
    std::string s;
} create_auction_t;

namespace ClockAuction
{
class SpectrumAuctionMessage
{
private:
    const std::string inputJsonString_;
    std::string jsonString_;

public:
    ErrorReport er_;

    SpectrumAuctionMessage();
    SpectrumAuctionMessage(const std::string& message);
    ErrorReport getErrorReport();

    std::string getJsonString();

    void toStatusJsonObject(JSON_Object* root_object, int rc, const std::string& message);
    void toWrappedStatusJsonObject(JSON_Object* root_object, int rc, const std::string& message);
    void toStatusJsonString(int rc, std::string& message, std::string& jsonString);

    void toCreateAuctionJson(int rc, const std::string& message, unsigned int auctionId);
    bool fromCreateAuctionJson(StaticAuctionState& staticAuctionState);

    void toStaticAuctionStateJson(const StaticAuctionState& staticAuctionState);
    bool fromStaticAuctionStateJson(StaticAuctionState& staticAuctionState);

    void toDynamicAuctionStateJson(const DynamicAuctionState& dynamicAuctionState);
    bool fromDynamicAuctionStateJson(DynamicAuctionState& dynamicAuctionState);

    bool fromGetAuctionDetailsJson(uint32_t& auctionId);
    void toGetAuctionDetailsJson(
        int rc, const std::string& message, const StaticAuctionState& staticAuctionState);

    bool fromGetAuctionStatusJson(uint32_t& auctionId);
    void toGetAuctionStatusJson(
        int rc, const std::string& message, const DynamicAuctionState& dynamicAuctionState);

    bool fromStartNextRoundJson(uint32_t& auctionId);
    void toStartNextRoundJson(int rc, const std::string& message);

    bool fromEndRoundJson(uint32_t& auctionId);
    void toEndRoundJson(int rc, const std::string& message);

    bool fromSubmitClockBidJson(ClockAuction::Bid& bid);
    void toSubmitClockBidJson(int rc, const std::string& message);

    bool fromGetRoundInfoJson(uint32_t& auctionId, uint32_t& requestedRound);
    void toGetRoundInfoJson(int rc,
        const std::string& message,
        const StaticAuctionState& sState,
        const DynamicAuctionState& dState,
        uint32_t requestedRound);

    bool fromGetBidderRoundResultsJson(uint32_t& auctionId, uint32_t& requestedRound);
    void toGetBidderRoundResultsJson(int rc,
        const std::string& message,
        const StaticAuctionState& sState,
        const DynamicAuctionState& dState,
        uint32_t requestedRound);

    bool fromGetOwnerRoundResultsJson(uint32_t& auctionId, uint32_t& requestedRound);
    void toGetOwnerRoundResultsJson(int rc,
        const std::string& message,
        const StaticAuctionState& sState,
        const DynamicAuctionState& dState,
        uint32_t requestedRound);

    bool fromGetAssignmentResultsJson(uint32_t& auctionId);
    void toGetAssignmentResultsJson(int rc,
        const std::string& message,
        const StaticAuctionState& sState,
        const DynamicAuctionState& dState);

    bool fromPublishAssignmentResultsJson(uint32_t& auctionId);
    void toPublishAssignmentResultsJson(int rc, const std::string& message);
};
}  // namespace ClockAuction
