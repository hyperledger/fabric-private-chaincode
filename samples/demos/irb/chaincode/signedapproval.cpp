/*
 * Copyright 2021 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "signedapproval.h"

bool Contract::SignedApproval::store(
    Contract::Storage& storage, const std::string& signedApprovalB64)
{
    FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, experimentId_.empty());

    std::string approvalKey("approval." + experimentId_);

    storage.ledgerPrivatePutString(approvalKey, signedApprovalB64);

    return true;
}

bool Contract::SignedApproval::retrieve(Contract::Storage& storage, std::string& signedApprovalB64)
{
    FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, experimentId_.empty());

    std::string approvalKey("approval." + experimentId_);
    signedApprovalB64.empty();
    storage.ledgerPrivateGetString(approvalKey, signedApprovalB64);

    if (signedApprovalB64.length() > 0)
    {
        return true;
    }
    return false;
}
