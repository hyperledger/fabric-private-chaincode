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
class Channel
{
private:
    uint32_t id_;
    std::string name_;
    uint32_t impairment_;

public:
    Channel();
    Channel(uint32_t id, std::string& name, uint32_t impairment);
    void toJsonObject(JSON_Object* root_object) const;
    bool fromJsonObject(const JSON_Object* root_object);

    ErrorReport er_;

    uint32_t getImpairment() const;
    uint32_t getId() const;
};

class Territory
{
private:
    uint32_t id_;
    std::string name_;
    bool isHighDemand_;
    double minPrice_;
    std::vector<Channel> channels_;

public:
    Territory();
    Territory(uint32_t id,
        std::string& name,
        bool isHighDemand,
        double minPrice,
        std::vector<Channel>& channels);
    void toJsonObject(JSON_Object* root_object) const;
    bool fromJsonObject(const JSON_Object* root_object);

    ErrorReport er_;

    uint32_t getTerritoryId() const;
    uint32_t numberOfChannels() const;
    double getMinPrice() const;
    bool isHighDemand() const;
    std::vector<uint32_t> getChannelImpairments() const;
    std::vector<uint32_t> getChannelIds() const;
};
}  // namespace ClockAuction
