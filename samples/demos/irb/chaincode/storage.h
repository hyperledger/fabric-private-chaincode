/*
 * Copyright 2021 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#pragma once

#include "common.h"

namespace Contract
{
class Storage
{
private:
    shim_ctx_ptr_t ctx_;

public:
    Storage(shim_ctx_ptr_t ctx);
    void ledgerPrivatePutString(const std::string& key, const std::string& value);
    void ledgerPrivateGetString(const std::string& key, std::string& value);

    void ledgerPublicPutString(const std::string& key, const std::string& value);

    void ledgerPublicPutBinary(const uint8_t* key,
        const uint32_t keyLength,
        const uint8_t* value,
        const uint32_t valueLength);

    void ledgerPrivatePutBinary(const uint8_t* key,
        const uint32_t keyLength,
        const uint8_t* value,
        const uint32_t valueLength);
    void ledgerPrivateGetBinary(const uint8_t* key,
        const uint32_t keyLength,
        uint8_t* value,
        const uint32_t valueLength,
        uint32_t* actualValueLength);
};
}  // namespace Contract
