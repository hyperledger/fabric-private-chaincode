/*
 * Copyright 2019 Intel Corporation
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#pragma once

#include <string>
#include "common.h"
#include "errors.h"

namespace Contract
{
class EASMessage
{
private:
    const std::string inputString_;
    ByteArray inputMessageBytes_;

public:
    ErrorReport er_;

    EASMessage();
    EASMessage(const std::string& message);
    ErrorReport getErrorReport();

    bool toStatus(const std::string& message, int rc, std::string& outputMessage);

    bool fromRegisterDataRequest(std::string& uuid, ByteArray& publicKey, ByteArray& decryptionKey);
};
}  // namespace Contract
