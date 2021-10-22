/*
 * Copyright 2021 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#pragma once

#include "common.h"

namespace Contract
{
class Id
{
public:
    std::string uuid_;
    ByteArray publicKey_;
    ByteArray publicEncryptionKey_;

    Id();
    Id(std::string& uuid, ByteArray& publicKey, ByteArray& publicEncryptionKey);
};
}  // namespace Contract
