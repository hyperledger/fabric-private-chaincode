/*
 * Copyright 2019 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#pragma once

#include "auction-state.h"
#include "common.h"
#include "error-codes.h"
#include "storage.h"

#define AUCTION_API_PROTOTYPE(api_name) \
    bool api_name(                      \
        const std::string& inputString, std::string& outputString, ClockAuction::ErrorReport& er)

namespace ClockAuction
{
class SpectrumAuction
{
private:
    uint32_t auctionIdCounter_;
    void InitializeAuctionIdCounter();
    void IncrementAndStoreAuctionIdCounter();

    StaticAuctionState staticAuctionState_;
    DynamicAuctionState dynamicAuctionState_;
    ClockAuction::Storage auctionStorage_;

    void storeAuctionState();
    bool loadAuctionState();

    void evaluateClockRound();
    void evaluateAssignmentRound();

public:
    SpectrumAuction(shim_ctx_ptr_t ctx);
    ErrorReport er_;

    AUCTION_API_PROTOTYPE(createAuction);
    AUCTION_API_PROTOTYPE(getAuctionDetails);
    AUCTION_API_PROTOTYPE(getAuctionStatus);
    AUCTION_API_PROTOTYPE(startNextRound);
    AUCTION_API_PROTOTYPE(endRound);
    AUCTION_API_PROTOTYPE(submitClockBid);
    AUCTION_API_PROTOTYPE(getRoundInfo);
    AUCTION_API_PROTOTYPE(getBidderRoundResults);
    AUCTION_API_PROTOTYPE(getOwnerRoundResults);
    AUCTION_API_PROTOTYPE(submitAssignmentBid);
    AUCTION_API_PROTOTYPE(getAssignmentResults);
    AUCTION_API_PROTOTYPE(publishAssignmentResults);
};
}  // namespace ClockAuction

typedef AUCTION_API_PROTOTYPE((ClockAuction::SpectrumAuction::*spectrumAuctionFunctionP));
