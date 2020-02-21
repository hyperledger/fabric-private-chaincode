/*
 * Copyright 2019 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include <cmath>
#include "auction-state.h"
#include "common.h"
#include "shim.h"  // get_random_bytes
#include "spectrum-auction.h"

void ClockAuction::SpectrumAuction::evaluateClockRound()
{
    dynamicAuctionState_.processBidsPreamble();
    if (dynamicAuctionState_.getRound() == 1)
    {
        LOG_DEBUG("Processing initial bids");
        dynamicAuctionState_.processInitialRoundBids(staticAuctionState_, auctionIdCounter_);
    }
    else
    {
        LOG_DEBUG("Processing regular bids");
        dynamicAuctionState_.processRegularRoundBids(staticAuctionState_, auctionIdCounter_);
    }

    dynamicAuctionState_.processBidsPostamble(staticAuctionState_);
    LOG_DEBUG("Processing complete");
}

void ClockAuction::SpectrumAuction::evaluateAssignmentRound()
{
    dynamicAuctionState_.processAssignmentRound(staticAuctionState_);
}

void ClockAuction::DynamicAuctionState::processBidsPreamble()
{
    er_.set(EC_SUCCESS, "");
    // NOTE: we pre-set to success. If any error occurs, this will be overwritten.
}

void ClockAuction::DynamicAuctionState::processInitialRoundBids(
    StaticAuctionState& sState, uint32_t auctionId)
{
    // missing bids
    fillMissingBids(sState, auctionId);

    // NOTE: missing bids are already set in startStart.
    // Such bids can be recognized by the auctionId = 0

    // set processed licenses
    processedLicenses_.resize(2);  // add the processedLicenses_[round=1] vector
    processedLicenses_[1].resize(sState.getBiddersN());
    for (unsigned int i = 0; i < sState.getBiddersN(); i++)
    {
        processedLicenses_[1][i].resize(sState.getTerritoryN());

        for (unsigned int j = 0; j < clockBids_[1][i].demands_.size(); j++)
        {
            uint32_t demandedTerritoryId = clockBids_[1][i].demands_[j].territoryId_;
            uint32_t demandedQuantity = clockBids_[1][i].demands_[j].quantity_;
            processedLicenses_[1][i][sState.getTerritoryIndex(demandedTerritoryId)] =
                demandedQuantity;
        }
    }

    // set postedPrice to min/initial price
    postedPrice_.resize(2);  // ensures postedPrice_[1] is defined
    postedPrice_[1] = sState.getInitialPrices();
}

void ClockAuction::DynamicAuctionState::processBidsPostamble(StaticAuctionState& sState)
{
    // aggregated demand (per territory)
    std::vector<uint32_t> aggregatedDemand(sState.getTerritoryN());
    for (unsigned int i = 0; i < aggregatedDemand.size(); i++)
    {
        aggregatedDemand[i] = 0;
        for (unsigned int j = 0; j < sState.getBiddersN(); j++)
        {
            aggregatedDemand[i] += processedLicenses_[clockRound_][j][i];
        }
    }

    // excess demand
    std::vector<uint32_t> supply = sState.getSupply();
    std::vector<int32_t> roundExcessDemand(sState.getTerritoryN());
    for (unsigned int i = 0; i < sState.getTerritoryN(); i++)
    {
        roundExcessDemand[i] = aggregatedDemand[i] - supply[i];
    }
    // Note: in regular bid processing, we have already added the excess demand
    //       so we simply double check that the vectors are the same
    if (excessDemand_.size() == clockRound_)
    {
        excessDemand_.push_back(roundExcessDemand);
    }
    else
    {
        if (excessDemand_[clockRound_] != roundExcessDemand)
        {
            std::string s("error excess demand - stop bid processing");
            LOG_ERROR("%s", s.c_str());
            er_.set(EC_EVALUATION_ERROR, s);
            return;
        }
    }

    // final stage rule check
    std::vector<bool> highDemand = sState.getHighDemandVector();
    for (unsigned int i = 0; i < roundExcessDemand.size(); i++)
    {
        if (roundExcessDemand[i] < 0 && highDemand[i])
        {
            LOG_INFO(
                "Final Stage Rule: FAILED - demand below supply for  %u-th high-demand territory",
                i);
            auctionState_ = FSR_FAILED;
        }
    }

    // check clock phase termination
    if (auctionState_ == CLOCK_PHASE)
    {
        if (*max_element(roundExcessDemand.begin(), roundExcessDemand.end()) <= 0)
        {
            LOG_INFO("Clock Phase: COMPLETED!");
            auctionState_ = ASSIGNMENT_PHASE;
        }
        else
        {
            LOG_INFO("Clock Phase: to be continued");

            // processed activity, required activity and new eligibility
            std::vector<uint32_t> processedActivity(sState.getBiddersN());
            std::vector<uint32_t> requiredActivity(sState.getBiddersN());
            std::vector<uint32_t> newRoundEligibility(sState.getBiddersN());
            for (unsigned int i = 0; i < sState.getBiddersN(); i++)
            {
                processedActivity[i] = 0;
                for (unsigned int j = 0; j < sState.getTerritoryN(); j++)
                {
                    processedActivity[i] += processedLicenses_[clockRound_][i][j];
                }
                requiredActivity[i] = eligibility_[clockRound_][i] *
                                      sState.getActivityRequirementPercentage() / 100.0;
                if (processedActivity[i] >= requiredActivity[i])
                {
                    newRoundEligibility[i] = eligibility_[clockRound_][i];
                }
                else
                {
                    newRoundEligibility[i] =
                        processedActivity[i] * 100 / sState.getActivityRequirementPercentage();
                }
            }
            eligibility_.push_back(newRoundEligibility);

            // clock price
            std::vector<double> newRoundClockPrice = postedPrice_[clockRound_];
            for (unsigned int j = 0; j < sState.getTerritoryN(); j++)
            {
                newRoundClockPrice[j] *=
                    (1.0 + ((double)sState.getClockPriceIncrementPercentage() / 100.0));
            }
            clockPrice_.push_back(newRoundClockPrice);

            // new round
            clockRound_++;
        }
    }
}

void ClockAuction::DynamicAuctionState::processRegularRoundBids(
    StaticAuctionState& sState, uint32_t auctionId)
{
    // missing bids
    fillMissingBids(sState, auctionId);

    // pre-set posted price
    std::vector<double> roundPostedPrice(sState.getTerritoryN());
    postedPrice_.push_back(roundPostedPrice);
    for (unsigned int i = 0; i < sState.getTerritoryN(); i++)
    {
        if (excessDemand_[clockRound_ - 1][i] > 0)
        {
            postedPrice_[clockRound_][i] = clockPrice_[clockRound_][i];
        }
        else
        {
            postedPrice_[clockRound_][i] = postedPrice_[clockRound_ - 1][i];
        }
    }

    // prepare bids to be processed
    std::list<enqueuedBid> eBidContainer;
    for (unsigned int i = 0; i < sState.getBiddersN(); i++)
    {
        // remove bids (actually, demands) that maintain current demand
        for (int j = clockBids_[clockRound_][i].demands_.size() - 1; j >= 0; j--)
        {
            // copy the bid
            enqueuedBid eBid;
            eBid.bidderIndex = i;
            eBid.demand = clockBids_[clockRound_][i].demands_[j];
            eBid.territoryIndex = sState.getTerritoryIndex(eBid.demand.territoryId_);

            // skip bid if it maintains demand
            unsigned int processedDemand =
                processedLicenses_[clockRound_ - 1][eBid.bidderIndex][eBid.territoryIndex];
            if (processedDemand == eBid.demand.quantity_)
            {
                continue;
            }

            // compute price point
            double minPrice = postedPrice_[clockRound_ - 1][eBid.territoryIndex];
            double clockPrice = clockPrice_[clockRound_][eBid.territoryIndex];
            double pp = (eBid.demand.price_ - minPrice) / (clockPrice - minPrice) * (100.0);
            eBid.pricePoint = std::lround(pp);

            // compute random value for breaking ties
            if (get_random_bytes((uint8_t*)&eBid.randomValue, sizeof(eBid.randomValue)) != 0)
            {
                std::string s("error getting random value, stop bid processing");
                LOG_ERROR("%s", s.c_str());
                er_.set(EC_EVALUATION_ERROR, s);
                return;
            }

            // enqueue bid
            eBidContainer.push_back(eBid);
        }
    }
    // sort bids by ascending price points
    eBidContainer.sort(compareByIncreasingPricePoint);

    // initialize current round processed licenses with previous round processed licenses
    processedLicenses_.push_back(processedLicenses_[clockRound_ - 1]);

    // compute available eligibility
    std::vector<uint32_t> eligibilityCap = eligibility_[clockRound_];
    for (unsigned int i = 0; i < sState.getBiddersN(); i++)
    {
        uint32_t bidderProcessedLicenses =
            std::accumulate(processedLicenses_[clockRound_ - 1][i].begin(),
                processedLicenses_[clockRound_ - 1][i].end(), 0);
        eligibilityCap[i] -= bidderProcessedLicenses;
    }

    // initialize excess demand for this round with the excess demand of the previous round
    excessDemand_.push_back(excessDemand_[clockRound_ - 1]);

    // process bids
    auto it = eBidContainer.begin();
    while (it != eBidContainer.end())
    {
        enqueuedBid& eBid = *it;

        bool fullyAppliedBid = false;
        unsigned int ownedQuantity =
            processedLicenses_[clockRound_][eBid.bidderIndex][eBid.territoryIndex];
        unsigned int demandedQuantity = eBid.demand.quantity_;

        if (demandedQuantity > ownedQuantity)
        {  // demand strictly increases
            // no eligibility => do not apply bid
            if (eligibilityCap[eBid.bidderIndex] == 0)
            {
                ++it;
                continue;
            }

            // compute processable quantity
            unsigned int deltaQuantity = demandedQuantity - ownedQuantity;
            unsigned int processableDeltaQuantity =
                (deltaQuantity <= eligibilityCap[eBid.bidderIndex]
                        ? deltaQuantity
                        : eligibilityCap[eBid.bidderIndex]);

            // set fully applied bid flag
            fullyAppliedBid = (processableDeltaQuantity == deltaQuantity);

            // update processed licenses, excess demand and eligibility accordingly
            processedLicenses_[clockRound_][eBid.bidderIndex][eBid.territoryIndex] +=
                processableDeltaQuantity;
            excessDemand_[clockRound_][eBid.territoryIndex] += processableDeltaQuantity;
            eligibilityCap[eBid.bidderIndex] -= processableDeltaQuantity;

            if (excessDemand_[clockRound_][eBid.territoryIndex] - processableDeltaQuantity == 0)
            {  // if excess demand was 0, reset posted price to clock price
                postedPrice_[clockRound_][eBid.territoryIndex] =
                    clockPrice_[clockRound_][eBid.territoryIndex];
            }
        }
        else  // demand strictly decreases (it cannot be equal, because we do not process those
              // bids)
        {
            // no excess demand => do not apply bid
            if (excessDemand_[clockRound_][eBid.territoryIndex] <= 0)
            {
                ++it;
                continue;
            }

            // compute processable quantity
            unsigned int deltaQuantity = ownedQuantity - demandedQuantity;
            unsigned int processableDeltaQuantity =
                (deltaQuantity <= excessDemand_[clockRound_][eBid.territoryIndex]
                        ? deltaQuantity
                        : excessDemand_[clockRound_][eBid.territoryIndex]);

            // set fully applied bid flag
            fullyAppliedBid = (processableDeltaQuantity == deltaQuantity);

            // update processed licenses, excess demand and eligibility accordingly
            processedLicenses_[clockRound_][eBid.bidderIndex][eBid.territoryIndex] -=
                processableDeltaQuantity;
            excessDemand_[clockRound_][eBid.territoryIndex] -= processableDeltaQuantity;
            eligibilityCap[eBid.bidderIndex] += processableDeltaQuantity;

            if (excessDemand_[clockRound_][eBid.territoryIndex] == 0)
            {  // if excess demand is now zero, the bid's price is the posted price
                postedPrice_[clockRound_][eBid.territoryIndex] = eBid.demand.price_;
            }
        }

        // remove bid if fully processed
        if (fullyAppliedBid)
        {
            it = eBidContainer.erase(it);
        }

        // restart bid processing to enforce increasing-price-point bid processing
        // (skipped/partially-fulfilled bids might be processable now)
        it = eBidContainer.begin();
    }
}

void ClockAuction::DynamicAuctionState::processAssignmentRound(StaticAuctionState& sState)
{
    {
        // randomly assign channels to winning bidders
        std::vector<uint32_t> terIds = sState.getTerritoryIds();
        winAssign_.resize(terIds.size());
        assignCost_.resize(terIds.size());
        channelPrice_.resize(terIds.size());
        for (unsigned int i = 0; i < terIds.size(); i++)
        {
            const Territory* t = sState.getTerritory(terIds[i]);
            uint32_t channelsN = t->numberOfChannels();

            // initialize vectors
            winAssign_[i].resize(channelsN, -1);
            channelPrice_[i].resize(channelsN, 0);
            assignCost_[i].resize(sState.getBiddersN(), 0);

            // compute a random vector of channel indexes
            std::vector<uint32_t> randomChannelIndexes;
            for (unsigned int j = 0; j < channelsN; j++)
                randomChannelIndexes.push_back(j);
            for (unsigned int j = 0; j < channelsN; j++)
            {
                unsigned int r;
                if (get_random_bytes((uint8_t*)&r, sizeof(unsigned int)) != 0)
                {
                    std::string s("error getting random value, stop assignment processing");
                    LOG_ERROR("%s", s.c_str());
                    er_.set(EC_EVALUATION_ERROR, s);
                    return;
                }
                r = r % channelsN;
                uint32_t a = randomChannelIndexes[j];
                randomChannelIndexes[j] = randomChannelIndexes[r];
                randomChannelIndexes[r] = a;
            }

            // compute channel prices
            std::vector<uint32_t> impairments = t->getChannelImpairments();
            for (unsigned int j = 0; j < channelsN; j++)
            {
                channelPrice_[i][j] =
                    postedPrice_[getRound()][i] * ((100.0 - (double)(impairments[j])) / 100.0);
            }

            // assign the randomized channel indexes to winning bidders
            uint32_t indexToAssign = 0;
            for (unsigned int j = 0; j < sState.getBiddersN(); j++)
            {
                for (unsigned int k = 0; k < processedLicenses_[getRound()][j][i]; k++)
                {
                    winAssign_[i][randomChannelIndexes[indexToAssign]] = j;
                    indexToAssign++;
                }
            }
        }
    }

    closeAuctionState();
}
