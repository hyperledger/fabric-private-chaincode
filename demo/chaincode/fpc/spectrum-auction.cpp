/*
 * Copyright 2019 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "spectrum-auction.h"
#include <string>
#include "bid.h"
#include "error-codes.h"
#include "spectrum-auction-message.h"
#include "storage.h"
#include "utils.h"

#define AUCTION_ID_COUNTER_STRING "AuctionIdCounter"
#define AUCTION_ID_FIRST_COUNTER_VALUE 1

ClockAuction::SpectrumAuction::SpectrumAuction(shim_ctx_ptr_t ctx)
    : dynamicAuctionState_(ctx), auctionStorage_(ctx)
{
}

void ClockAuction::SpectrumAuction::InitializeAuctionIdCounter()
{
    uint32_t bytesRead = 0;
    const uint8_t* key = (const uint8_t*)AUCTION_ID_COUNTER_STRING;
    uint32_t keyLength = strlen(AUCTION_ID_COUNTER_STRING);
    uint8_t* value = (uint8_t*)&auctionIdCounter_;
    uint32_t valueLength = sizeof(auctionIdCounter_);
    auctionStorage_.ledgerPrivateGetBinary(key, keyLength, value, valueLength, &bytesRead);
    if (bytesRead == 0)
        auctionIdCounter_ = AUCTION_ID_FIRST_COUNTER_VALUE;
}

void ClockAuction::SpectrumAuction::IncrementAndStoreAuctionIdCounter()
{
    uint32_t c = auctionIdCounter_ + 1;
    const uint8_t* key = (const uint8_t*)AUCTION_ID_COUNTER_STRING;
    uint32_t keyLength = strlen(AUCTION_ID_COUNTER_STRING);
    uint8_t* value = (uint8_t*)&c;
    uint32_t valueLength = sizeof(auctionIdCounter_);
    auctionStorage_.ledgerPrivatePutBinary(key, keyLength, value, valueLength);
}

void ClockAuction::SpectrumAuction::storeAuctionState()
{
    {  // store static state
        ClockAuction::SpectrumAuctionMessage outStateMsg;
        outStateMsg.toStaticAuctionStateJson(staticAuctionState_);
        std::string stateKey(
            "Auction." + std::to_string(auctionIdCounter_) + ".staticAuctionState");
        auctionStorage_.ledgerPrivatePutString(stateKey, outStateMsg.getJsonString());
        LOG_DEBUG("Stored static state: %s", (outStateMsg.getJsonString()).c_str());
    }
    {  // store dynamic state
        ClockAuction::SpectrumAuctionMessage outStateMsg;
        outStateMsg.toDynamicAuctionStateJson(dynamicAuctionState_);
        std::string stateKey(
            "Auction." + std::to_string(auctionIdCounter_) + ".dynamicAuctionState");
        auctionStorage_.ledgerPrivatePutString(stateKey, outStateMsg.getJsonString());
        LOG_DEBUG("Stored dynamic state: %s", (outStateMsg.getJsonString()).c_str());
    }
}

bool ClockAuction::SpectrumAuction::loadAuctionState()
{
    {  // load static state
        std::string stateKey(
            "Auction." + std::to_string(auctionIdCounter_) + ".staticAuctionState");
        std::string stateValue;
        auctionStorage_.ledgerPrivateGetString(stateKey, stateValue);
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, stateValue.length() == 0);
        // next is same as for createAuction
        JSON_Object* root_object = ClockAuction::JsonUtils::openJsonObject(stateValue.c_str());
        FAST_FAIL_CHECK_EX(er_, &staticAuctionState_.er_, EC_INVALID_INPUT,
            !staticAuctionState_.fromExtendedJsonObject(root_object));
        ClockAuction::JsonUtils::closeJsonObject(root_object, NULL);
    }
    {  // load dynamic state
        std::string stateKey(
            "Auction." + std::to_string(auctionIdCounter_) + ".dynamicAuctionState");
        std::string stateValue;
        auctionStorage_.ledgerPrivateGetString(stateKey, stateValue);
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, stateValue.length() == 0);
        JSON_Object* root_object = ClockAuction::JsonUtils::openJsonObject(stateValue.c_str());
        FAST_FAIL_CHECK_EX(er_, &dynamicAuctionState_.er_, EC_INVALID_INPUT,
            !dynamicAuctionState_.fromJsonObject(root_object));
        ClockAuction::JsonUtils::closeJsonObject(root_object, NULL);
    }
    return true;
}

AUCTION_API_PROTOTYPE(ClockAuction::SpectrumAuction::createAuction)
{
    // parse and validate input string
    ClockAuction::SpectrumAuctionMessage inMsg(inputString);
    FAST_FAIL_CHECK_EX(
        er, &inMsg.er_, EC_INVALID_INPUT, !inMsg.fromCreateAuctionJson(staticAuctionState_));
    FAST_FAIL_CHECK_EX(
        er, &staticAuctionState_.er_, EC_INVALID_INPUT, !staticAuctionState_.checkValidity());

    // no authentication, anybody can create

    // all check passed, install auction

    // get auction id
    InitializeAuctionIdCounter();
    IncrementAndStoreAuctionIdCounter();

    // initialize dynamic state
    dynamicAuctionState_.initialize(
        CLOCK_PHASE, INITIAL_CLOCK_ROUND_NUMBER, false, staticAuctionState_);

    // store static and dynamic state
    storeAuctionState();

    er.set(EC_SUCCESS, "");

    // prepare response message
    ClockAuction::SpectrumAuctionMessage msg;
    msg.toCreateAuctionJson(0, "Auction created", auctionIdCounter_);
    outputString = msg.getJsonString();
    return true;
}

AUCTION_API_PROTOTYPE(ClockAuction::SpectrumAuction::getAuctionDetails)
{
    // parse and validate input string
    ClockAuction::SpectrumAuctionMessage inMsg(inputString);
    FAST_FAIL_CHECK(er, EC_INVALID_INPUT, !inMsg.fromGetAuctionDetailsJson(auctionIdCounter_));
    FAST_FAIL_CHECK_EX(er, &er_, EC_INVALID_INPUT, !loadAuctionState());

    // no authentication, anybody can execute

    // all check passed, return details
    er.set(EC_SUCCESS, "");
    ClockAuction::SpectrumAuctionMessage msg;
    msg.toGetAuctionDetailsJson(0, "Auction details", staticAuctionState_);
    outputString = msg.getJsonString();
    return true;
}

AUCTION_API_PROTOTYPE(ClockAuction::SpectrumAuction::getAuctionStatus)
{
    // parse and validate input string
    ClockAuction::SpectrumAuctionMessage inMsg(inputString);
    FAST_FAIL_CHECK(er, EC_INVALID_INPUT, !inMsg.fromGetAuctionStatusJson(auctionIdCounter_));
    FAST_FAIL_CHECK_EX(er, &er_, EC_INVALID_INPUT, !loadAuctionState());

    // no authentication, anybody can execute

    // all check passed, return status
    er.set(EC_SUCCESS, "");
    ClockAuction::SpectrumAuctionMessage msg;
    msg.toGetAuctionStatusJson(0, "Auction status", dynamicAuctionState_);
    outputString = msg.getJsonString();
    return true;
}

AUCTION_API_PROTOTYPE(ClockAuction::SpectrumAuction::startNextRound)
{
    // parse and validate input string
    ClockAuction::SpectrumAuctionMessage inMsg(inputString);
    FAST_FAIL_CHECK(er, EC_INVALID_INPUT, !inMsg.fromStartNextRoundJson(auctionIdCounter_));
    FAST_FAIL_CHECK_EX(er, &er_, EC_INVALID_INPUT, !loadAuctionState());
    FAST_FAIL_CHECK(er, EC_ROUND_ACTIVE, dynamicAuctionState_.isRoundActive());
    FAST_FAIL_CHECK(er, EC_RESTRICTED_AUCTION_STATE,
        !dynamicAuctionState_.isStateClockPhase() &&
            !dynamicAuctionState_.isStateAssignmentPhase());

    // authenticate auction owner
    FAST_FAIL_CHECK(
        er, EC_UNRECOGNIZED_SUBMITTER, !dynamicAuctionState_.isValidOwner(staticAuctionState_));

    // all check passed

    dynamicAuctionState_.startRound(staticAuctionState_);

    storeAuctionState();

    er.set(EC_SUCCESS, "");
    ClockAuction::SpectrumAuctionMessage msg;
    msg.toStartNextRoundJson(0, "Start next round");
    outputString = msg.getJsonString();
    return true;
}

AUCTION_API_PROTOTYPE(ClockAuction::SpectrumAuction::endRound)
{
    // parse and validate input string
    ClockAuction::SpectrumAuctionMessage inMsg(inputString);
    FAST_FAIL_CHECK(er, EC_INVALID_INPUT, !inMsg.fromEndRoundJson(auctionIdCounter_));
    FAST_FAIL_CHECK_EX(er, &er_, EC_INVALID_INPUT, !loadAuctionState());
    FAST_FAIL_CHECK(er, EC_ROUND_NOT_ACTIVE, !dynamicAuctionState_.isRoundActive());
    FAST_FAIL_CHECK(er, EC_RESTRICTED_AUCTION_STATE,
        !dynamicAuctionState_.isStateClockPhase() &&
            !dynamicAuctionState_.isStateAssignmentPhase());

    // authenticate auction owner
    FAST_FAIL_CHECK(
        er, EC_UNRECOGNIZED_SUBMITTER, !dynamicAuctionState_.isValidOwner(staticAuctionState_));

    // all check passed

    dynamicAuctionState_.endRound();
    evaluateClockRound();
    FAST_FAIL_CHECK_EX(
        er, &dynamicAuctionState_.er_, EC_INVALID_INPUT, !dynamicAuctionState_.er_.isSuccess());

    if (dynamicAuctionState_.isStateAssignmentPhase())
    {
        // *********
        // IMPORTANT: we short-circuit the assignment phase, by starting it immediately with
        // the end-round invocation (though only when the clock phase terminates)
        // *********
        LOG_INFO(
            "Assignment Phase short-circuit: start immediately, assign randomly, terminate, "
            "evaluate");
        dynamicAuctionState_.startRound(staticAuctionState_);
        dynamicAuctionState_.endRound();
        evaluateAssignmentRound();

        FAST_FAIL_CHECK(er, EC_EVALUATION_ERROR, !dynamicAuctionState_.isStateClosedPhase());
        LOG_INFO("Assignment phase complete --> Auction closed");
    }

    storeAuctionState();

    er.set(EC_SUCCESS, "");
    ClockAuction::SpectrumAuctionMessage msg;
    msg.toEndRoundJson(0, "End round");
    outputString = msg.getJsonString();
    return true;
}

AUCTION_API_PROTOTYPE(ClockAuction::SpectrumAuction::submitClockBid)
{
    // parse and validate input string
    ClockAuction::SpectrumAuctionMessage inMsg(inputString);
    Bid submittedBid;
    FAST_FAIL_CHECK_EX(
        er, &inMsg.er_, EC_INVALID_INPUT, !inMsg.fromSubmitClockBidJson(submittedBid));
    auctionIdCounter_ = submittedBid.auctionId_;
    // retrieve auction data (also checks the auction id)
    FAST_FAIL_CHECK_EX(er, &er_, EC_INVALID_INPUT, !loadAuctionState());

    // check clock phase
    FAST_FAIL_CHECK(er, EC_NOT_IN_CLOCK_PHASE, !dynamicAuctionState_.isStateClockPhase());

    // authenticate bidder
    FAST_FAIL_CHECK(
        er, EC_UNRECOGNIZED_SUBMITTER, !dynamicAuctionState_.isValidBidder(staticAuctionState_));

    // validate bid
    FAST_FAIL_CHECK_EX(er, &dynamicAuctionState_.er_, EC_INVALID_INPUT,
        !dynamicAuctionState_.isValidBid(staticAuctionState_, submittedBid));

    // all check passed

    dynamicAuctionState_.storeBid(staticAuctionState_, submittedBid);

    storeAuctionState();

    er.set(EC_SUCCESS, "");
    ClockAuction::SpectrumAuctionMessage msg;
    msg.toSubmitClockBidJson(0, "Submit clock bid");
    outputString = msg.getJsonString();
    return true;
}

AUCTION_API_PROTOTYPE(ClockAuction::SpectrumAuction::getRoundInfo)
{
    // parse and validate input string
    ClockAuction::SpectrumAuctionMessage inMsg(inputString);
    uint32_t requestedRound;
    FAST_FAIL_CHECK(
        er, EC_INVALID_INPUT, !inMsg.fromGetRoundInfoJson(auctionIdCounter_, requestedRound));
    FAST_FAIL_CHECK_EX(er, &er_, EC_INVALID_INPUT, !loadAuctionState());
    FAST_FAIL_CHECK(er, EC_INVALID_INPUT,
        !dynamicAuctionState_.isStateClockPhase() &&
            !dynamicAuctionState_.isStateAssignmentPhase() &&
            !dynamicAuctionState_.isStateClosedPhase());
    FAST_FAIL_CHECK(er, EC_INVALID_INPUT, requestedRound > dynamicAuctionState_.getRound());

    // no authentication, anybody can execute

    // all check passed, return details

    // nothing to store, state not updated

    er.set(EC_SUCCESS, "");
    ClockAuction::SpectrumAuctionMessage msg;
    msg.toGetRoundInfoJson(
        0, "Get round info", staticAuctionState_, dynamicAuctionState_, requestedRound);
    outputString = msg.getJsonString();
    return true;
}

AUCTION_API_PROTOTYPE(ClockAuction::SpectrumAuction::getBidderRoundResults)
{
    // parse and validate input string
    ClockAuction::SpectrumAuctionMessage inMsg(inputString);
    uint32_t requestedRound;
    FAST_FAIL_CHECK(er, EC_INVALID_INPUT,
        !inMsg.fromGetBidderRoundResultsJson(auctionIdCounter_, requestedRound));
    FAST_FAIL_CHECK_EX(er, &er_, EC_INVALID_INPUT, !loadAuctionState());
    FAST_FAIL_CHECK(er, EC_INVALID_INPUT,
        !dynamicAuctionState_.isStateClockPhase() &&
            !dynamicAuctionState_.isStateAssignmentPhase() &&
            !dynamicAuctionState_.isStateClosedPhase());
    FAST_FAIL_CHECK(er, EC_INVALID_INPUT,
        dynamicAuctionState_.isStateClockPhase() &&
            requestedRound >= dynamicAuctionState_.getRound());

    // authenticate bidder
    FAST_FAIL_CHECK(
        er, EC_UNRECOGNIZED_SUBMITTER, !dynamicAuctionState_.isValidBidder(staticAuctionState_));

    // all check passed, return details

    // nothing to store, state not updated

    er.set(EC_SUCCESS, "");
    ClockAuction::SpectrumAuctionMessage msg;
    msg.toGetBidderRoundResultsJson(
        0, "Get bidder round results", staticAuctionState_, dynamicAuctionState_, requestedRound);
    outputString = msg.getJsonString();
    return true;
}

AUCTION_API_PROTOTYPE(ClockAuction::SpectrumAuction::getOwnerRoundResults)
{
    // parse and validate input string
    ClockAuction::SpectrumAuctionMessage inMsg(inputString);
    uint32_t requestedRound;
    FAST_FAIL_CHECK(er, EC_INVALID_INPUT,
        !inMsg.fromGetOwnerRoundResultsJson(auctionIdCounter_, requestedRound));
    FAST_FAIL_CHECK_EX(er, &er_, EC_INVALID_INPUT, !loadAuctionState());

    // authenticate auction owner
    FAST_FAIL_CHECK(
        er, EC_UNRECOGNIZED_SUBMITTER, !dynamicAuctionState_.isValidOwner(staticAuctionState_));

    FAST_FAIL_CHECK(er, EC_INVALID_INPUT,
        !dynamicAuctionState_.isStateClockPhase() &&
            !dynamicAuctionState_.isStateAssignmentPhase() &&
            !dynamicAuctionState_.isStateClosedPhase());
    FAST_FAIL_CHECK(er, EC_INVALID_INPUT,
        dynamicAuctionState_.isStateClockPhase() &&
            requestedRound >= dynamicAuctionState_.getRound());

    // all check passed, return details

    // nothing to store, state not updated

    er.set(EC_SUCCESS, "");
    ClockAuction::SpectrumAuctionMessage msg;
    msg.toGetOwnerRoundResultsJson(
        0, "Get owner round results", staticAuctionState_, dynamicAuctionState_, requestedRound);
    outputString = msg.getJsonString();
    return true;
}

AUCTION_API_PROTOTYPE(ClockAuction::SpectrumAuction::submitAssignmentBid)
{
    FAST_FAIL_CHECK(er, EC_UNIMPLEMENTED_API, true);
}

AUCTION_API_PROTOTYPE(ClockAuction::SpectrumAuction::getAssignmentResults)
{
    // parse and validate input string
    ClockAuction::SpectrumAuctionMessage inMsg(inputString);
    FAST_FAIL_CHECK(er, EC_INVALID_INPUT, !inMsg.fromGetAssignmentResultsJson(auctionIdCounter_));
    FAST_FAIL_CHECK_EX(er, &er_, EC_INVALID_INPUT, !loadAuctionState());

    FAST_FAIL_CHECK(er, EC_INVALID_INPUT, !dynamicAuctionState_.isStateClosedPhase());

    // no authentication, anybody can execute

    // all check passed, return details

    er.set(EC_SUCCESS, "");
    ClockAuction::SpectrumAuctionMessage msg;
    msg.toGetAssignmentResultsJson(
        0, "Get assignment results", staticAuctionState_, dynamicAuctionState_);
    outputString = msg.getJsonString();
    return true;
}

AUCTION_API_PROTOTYPE(ClockAuction::SpectrumAuction::publishAssignmentResults)
{
    // parse and validate input string
    ClockAuction::SpectrumAuctionMessage inMsg(inputString);
    FAST_FAIL_CHECK(
        er, EC_INVALID_INPUT, !inMsg.fromPublishAssignmentResultsJson(auctionIdCounter_));
    FAST_FAIL_CHECK_EX(er, &er_, EC_INVALID_INPUT, !loadAuctionState());

    FAST_FAIL_CHECK(er, EC_INVALID_INPUT, !dynamicAuctionState_.isStateClosedPhase());

    // no authentication, anybody can execute

    // all check passed, return details

    {
        // create string of public results
        std::string publicResultsString;
        JSON_Object* r = ClockAuction::JsonUtils::openJsonObject(NULL);
        dynamicAuctionState_.toAssignmentResultsJsonObject(r, staticAuctionState_);
        ClockAuction::JsonUtils::closeJsonObject(r, &publicResultsString);

        // publish results on the ledger publicly (no encryption)
        std::string publicResultsKeyString(
            "Auction." + std::to_string(auctionIdCounter_) + ".publicResults");
        auctionStorage_.ledgerPublicPutString(publicResultsKeyString, publicResultsString);
        LOG_DEBUG("Published Results: %s", publicResultsString.c_str());
    }

    er.set(EC_SUCCESS, "");
    ClockAuction::SpectrumAuctionMessage msg;
    msg.toPublishAssignmentResultsJson(0, "Publish assignment results");
    outputString = msg.getJsonString();
    return true;
}
