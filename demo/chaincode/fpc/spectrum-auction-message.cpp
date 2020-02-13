/*
 * Copyright 2019 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "spectrum-auction-message.h"
#include "common.h"
#include "error-codes.h"
#include "utils.h"

ClockAuction::SpectrumAuctionMessage::SpectrumAuctionMessage() {}

ClockAuction::SpectrumAuctionMessage::SpectrumAuctionMessage(const std::string& message)
    : inputJsonString_(message)
{
}

ClockAuction::ErrorReport ClockAuction::SpectrumAuctionMessage::getErrorReport()
{
    return er_;
}

void ClockAuction::SpectrumAuctionMessage::toStatusJsonObject(
    JSON_Object* root_object, int rc, const std::string& message)
{
    json_object_set_number(root_object, "rc", rc);
    json_object_set_string(root_object, "message", message.c_str());
}

void ClockAuction::SpectrumAuctionMessage::toWrappedStatusJsonObject(
    JSON_Object* root_object, int rc, const std::string& message)
{
    JSON_Object* r = ClockAuction::JsonUtils::openJsonObject(NULL);
    toStatusJsonObject(r, rc, message);
    json_object_set_value(root_object, "status", json_object_get_wrapping_value(r));
}

void ClockAuction::SpectrumAuctionMessage::toStatusJsonString(
    int rc, std::string& message, std::string& jsonString)
{
    JSON_Object* root_object = ClockAuction::JsonUtils::openJsonObject(NULL);
    toStatusJsonObject(root_object, rc, message);
    ClockAuction::JsonUtils::closeJsonObject(root_object, &jsonString);
}

std::string ClockAuction::SpectrumAuctionMessage::getJsonString()
{
    return jsonString_;
}

void ClockAuction::SpectrumAuctionMessage::toCreateAuctionJson(
    int rc, const std::string& message, unsigned int auctionId)
{
    JSON_Object* root_object = ClockAuction::JsonUtils::openJsonObject(NULL);
    {
        toWrappedStatusJsonObject(root_object, rc, message);
    }
    {
        JSON_Object* r = ClockAuction::JsonUtils::openJsonObject(NULL);
        json_object_set_number(r, "auctionId", auctionId);
        json_object_set_value(root_object, "response", json_object_get_wrapping_value(r));
    }
    ClockAuction::JsonUtils::closeJsonObject(root_object, &jsonString_);
}

bool ClockAuction::SpectrumAuctionMessage::fromCreateAuctionJson(
    StaticAuctionState& staticAuctionState)
{
    JSON_Object* root_object = ClockAuction::JsonUtils::openJsonObject(inputJsonString_.c_str());
    bool b = staticAuctionState.fromJsonObject(root_object);
    ClockAuction::JsonUtils::closeJsonObject(root_object, NULL);
    FAST_FAIL_CHECK_EX(er_, &staticAuctionState.er_, EC_INVALID_INPUT, !b);
    return true;
}

bool ClockAuction::SpectrumAuctionMessage::fromGetAuctionDetailsJson(uint32_t& auctionId)
{
    JSON_Object* root_object = ClockAuction::JsonUtils::openJsonObject(inputJsonString_.c_str());
    auctionId = json_object_get_number(root_object, "auctionId");
    ClockAuction::JsonUtils::closeJsonObject(root_object, NULL);
    return auctionId != 0;
}

void ClockAuction::SpectrumAuctionMessage::toGetAuctionDetailsJson(
    int rc, const std::string& message, const StaticAuctionState& staticAuctionState)
{
    JSON_Object* root_object = ClockAuction::JsonUtils::openJsonObject(NULL);
    {
        toWrappedStatusJsonObject(root_object, rc, message);
    }
    {
        JSON_Object* r = ClockAuction::JsonUtils::openJsonObject(NULL);
        staticAuctionState.toJsonObject(r);
        json_object_set_value(root_object, "response", json_object_get_wrapping_value(r));
    }
    ClockAuction::JsonUtils::closeJsonObject(root_object, &jsonString_);
}

bool ClockAuction::SpectrumAuctionMessage::fromGetAuctionStatusJson(uint32_t& auctionId)
{
    JSON_Object* root_object = ClockAuction::JsonUtils::openJsonObject(inputJsonString_.c_str());
    auctionId = json_object_get_number(root_object, "auctionId");
    ClockAuction::JsonUtils::closeJsonObject(root_object, NULL);
    return auctionId != 0;
}

void ClockAuction::SpectrumAuctionMessage::toGetAuctionStatusJson(
    int rc, const std::string& message, const DynamicAuctionState& dynamicAuctionState)
{
    JSON_Object* root_object = ClockAuction::JsonUtils::openJsonObject(NULL);
    {
        toWrappedStatusJsonObject(root_object, rc, message);
    }
    {
        JSON_Object* r = ClockAuction::JsonUtils::openJsonObject(NULL);
        dynamicAuctionState.toJsonObject(r);
        json_object_set_value(root_object, "response", json_object_get_wrapping_value(r));
    }
    ClockAuction::JsonUtils::closeJsonObject(root_object, &jsonString_);
}

bool ClockAuction::SpectrumAuctionMessage::fromStartNextRoundJson(uint32_t& auctionId)
{
    JSON_Object* root_object = ClockAuction::JsonUtils::openJsonObject(inputJsonString_.c_str());
    auctionId = json_object_get_number(root_object, "auctionId");
    ClockAuction::JsonUtils::closeJsonObject(root_object, NULL);
    return auctionId != 0;
}

void ClockAuction::SpectrumAuctionMessage::toStartNextRoundJson(int rc, const std::string& message)
{
    JSON_Object* root_object = ClockAuction::JsonUtils::openJsonObject(NULL);
    {
        toWrappedStatusJsonObject(root_object, rc, message);
    }
    ClockAuction::JsonUtils::closeJsonObject(root_object, &jsonString_);
}

bool ClockAuction::SpectrumAuctionMessage::fromEndRoundJson(uint32_t& auctionId)
{
    JSON_Object* root_object = ClockAuction::JsonUtils::openJsonObject(inputJsonString_.c_str());
    auctionId = json_object_get_number(root_object, "auctionId");
    ClockAuction::JsonUtils::closeJsonObject(root_object, NULL);
    return auctionId != 0;
}

void ClockAuction::SpectrumAuctionMessage::toEndRoundJson(int rc, const std::string& message)
{
    JSON_Object* root_object = ClockAuction::JsonUtils::openJsonObject(NULL);
    {
        toWrappedStatusJsonObject(root_object, rc, message);
    }
    ClockAuction::JsonUtils::closeJsonObject(root_object, &jsonString_);
}

bool ClockAuction::SpectrumAuctionMessage::fromSubmitClockBidJson(Bid& bid)
{
    JSON_Object* root_object = ClockAuction::JsonUtils::openJsonObject(inputJsonString_.c_str());
    FAST_FAIL_CHECK_EX(er_, &bid.er_, EC_INVALID_INPUT, !bid.fromJsonObject(root_object));
    ClockAuction::JsonUtils::closeJsonObject(root_object, NULL);
    return true;
}

void ClockAuction::SpectrumAuctionMessage::toSubmitClockBidJson(int rc, const std::string& message)
{
    JSON_Object* root_object = ClockAuction::JsonUtils::openJsonObject(NULL);
    {
        toWrappedStatusJsonObject(root_object, rc, message);
    }
    ClockAuction::JsonUtils::closeJsonObject(root_object, &jsonString_);
}

void ClockAuction::SpectrumAuctionMessage::toStaticAuctionStateJson(
    const StaticAuctionState& staticAuctionState)
{
    JSON_Object* root_object = ClockAuction::JsonUtils::openJsonObject(NULL);
    staticAuctionState.toJsonObject(root_object);
    ClockAuction::JsonUtils::closeJsonObject(root_object, &jsonString_);
}

bool ClockAuction::SpectrumAuctionMessage::fromStaticAuctionStateJson(
    StaticAuctionState& staticAuctionState)
{
    JSON_Object* root_object = ClockAuction::JsonUtils::openJsonObject(jsonString_.c_str());
    bool b = staticAuctionState.fromJsonObject(root_object);
    ClockAuction::JsonUtils::closeJsonObject(root_object, NULL);
    return b;
}

void ClockAuction::SpectrumAuctionMessage::toDynamicAuctionStateJson(
    const DynamicAuctionState& dynamicAuctionState)
{
    JSON_Object* root_object = ClockAuction::JsonUtils::openJsonObject(NULL);
    dynamicAuctionState.toJsonObject(root_object);
    ClockAuction::JsonUtils::closeJsonObject(root_object, &jsonString_);
}

bool ClockAuction::SpectrumAuctionMessage::fromDynamicAuctionStateJson(
    DynamicAuctionState& dynamicAuctionState)
{
    JSON_Object* root_object = ClockAuction::JsonUtils::openJsonObject(jsonString_.c_str());
    bool b = dynamicAuctionState.fromJsonObject(root_object);
    ClockAuction::JsonUtils::closeJsonObject(root_object, NULL);
    return b;
}

bool ClockAuction::SpectrumAuctionMessage::fromGetRoundInfoJson(
    uint32_t& auctionId, uint32_t& requestedRound)
{
    JSON_Object* root_object = ClockAuction::JsonUtils::openJsonObject(inputJsonString_.c_str());
    auctionId = json_object_get_number(root_object, "auctionId");
    requestedRound = json_object_get_number(root_object, "round");
    ClockAuction::JsonUtils::closeJsonObject(root_object, NULL);
    // returns false is any of the parameters does not exist or is 0
    return auctionId != 0 && requestedRound != 0;
}

void ClockAuction::SpectrumAuctionMessage::toGetRoundInfoJson(int rc,
    const std::string& message,
    const StaticAuctionState& sState,
    const DynamicAuctionState& dState,
    uint32_t requestedRound)
{
    JSON_Object* root_object = ClockAuction::JsonUtils::openJsonObject(NULL);
    {
        toWrappedStatusJsonObject(root_object, rc, message);
    }
    {
        JSON_Object* r = ClockAuction::JsonUtils::openJsonObject(NULL);
        dState.toRoundInfoJsonObject(r, sState, requestedRound);
        json_object_set_value(root_object, "response", json_object_get_wrapping_value(r));
    }
    ClockAuction::JsonUtils::closeJsonObject(root_object, &jsonString_);
}

bool ClockAuction::SpectrumAuctionMessage::fromGetBidderRoundResultsJson(
    uint32_t& auctionId, uint32_t& requestedRound)
{
    JSON_Object* root_object = ClockAuction::JsonUtils::openJsonObject(inputJsonString_.c_str());
    auctionId = json_object_get_number(root_object, "auctionId");
    requestedRound = json_object_get_number(root_object, "round");
    ClockAuction::JsonUtils::closeJsonObject(root_object, NULL);
    // returns false is any of the parameters does not exist or is 0
    return auctionId != 0 && requestedRound != 0;
}

void ClockAuction::SpectrumAuctionMessage::toGetBidderRoundResultsJson(int rc,
    const std::string& message,
    const StaticAuctionState& sState,
    const DynamicAuctionState& dState,
    uint32_t requestedRound)
{
    JSON_Object* root_object = ClockAuction::JsonUtils::openJsonObject(NULL);
    {
        toWrappedStatusJsonObject(root_object, rc, message);
    }
    {
        JSON_Object* r = ClockAuction::JsonUtils::openJsonObject(NULL);
        dState.toBidderRoundResultsJsonObject(r, sState, requestedRound);
        json_object_set_value(root_object, "response", json_object_get_wrapping_value(r));
    }
    ClockAuction::JsonUtils::closeJsonObject(root_object, &jsonString_);
}

bool ClockAuction::SpectrumAuctionMessage::fromGetOwnerRoundResultsJson(
    uint32_t& auctionId, uint32_t& requestedRound)
{
    JSON_Object* root_object = ClockAuction::JsonUtils::openJsonObject(inputJsonString_.c_str());
    auctionId = json_object_get_number(root_object, "auctionId");
    requestedRound = json_object_get_number(root_object, "round");
    ClockAuction::JsonUtils::closeJsonObject(root_object, NULL);
    // returns false is any of the parameters does not exist or is 0
    return auctionId != 0 && requestedRound != 0;
}

void ClockAuction::SpectrumAuctionMessage::toGetOwnerRoundResultsJson(int rc,
    const std::string& message,
    const StaticAuctionState& sState,
    const DynamicAuctionState& dState,
    uint32_t requestedRound)
{
    JSON_Object* root_object = ClockAuction::JsonUtils::openJsonObject(NULL);
    {
        toWrappedStatusJsonObject(root_object, rc, message);
    }
    {
        JSON_Object* r = ClockAuction::JsonUtils::openJsonObject(NULL);
        dState.toOwnerRoundResultsJsonObject(r, sState, requestedRound);
        json_object_set_value(root_object, "response", json_object_get_wrapping_value(r));
    }
    ClockAuction::JsonUtils::closeJsonObject(root_object, &jsonString_);
}

bool ClockAuction::SpectrumAuctionMessage::fromGetAssignmentResultsJson(uint32_t& auctionId)
{
    JSON_Object* root_object = ClockAuction::JsonUtils::openJsonObject(inputJsonString_.c_str());
    auctionId = json_object_get_number(root_object, "auctionId");
    ClockAuction::JsonUtils::closeJsonObject(root_object, NULL);
    // returns false is any of the parameters does not exist or is 0
    return auctionId != 0;
}

void ClockAuction::SpectrumAuctionMessage::toGetAssignmentResultsJson(int rc,
    const std::string& message,
    const StaticAuctionState& sState,
    const DynamicAuctionState& dState)
{
    JSON_Object* root_object = ClockAuction::JsonUtils::openJsonObject(NULL);
    {
        toWrappedStatusJsonObject(root_object, rc, message);
    }
    {
        JSON_Object* r = ClockAuction::JsonUtils::openJsonObject(NULL);
        dState.toAssignmentResultsJsonObject(r, sState);
        json_object_set_value(root_object, "response", json_object_get_wrapping_value(r));
    }
    ClockAuction::JsonUtils::closeJsonObject(root_object, &jsonString_);
}

bool ClockAuction::SpectrumAuctionMessage::fromPublishAssignmentResultsJson(uint32_t& auctionId)
{
    JSON_Object* root_object = ClockAuction::JsonUtils::openJsonObject(inputJsonString_.c_str());
    auctionId = json_object_get_number(root_object, "auctionId");
    ClockAuction::JsonUtils::closeJsonObject(root_object, NULL);
    // returns false is any of the parameters does not exist or is 0
    return auctionId != 0;
}

void ClockAuction::SpectrumAuctionMessage::toPublishAssignmentResultsJson(
    int rc, const std::string& message)
{
    JSON_Object* root_object = ClockAuction::JsonUtils::openJsonObject(NULL);
    {
        toWrappedStatusJsonObject(root_object, rc, message);
    }
    ClockAuction::JsonUtils::closeJsonObject(root_object, &jsonString_);
}
