/*
 * Copyright 2019 Intel Corporation
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#pragma once

#include <string>
#include "errors.h"

namespace Contract
{
class EASMessage
{
private:
    const std::string inputString_;

public:
    ErrorReport er_;

    EASMessage();
    EASMessage(const std::string& message);
    ErrorReport getErrorReport();
};
}  // namespace Contract
