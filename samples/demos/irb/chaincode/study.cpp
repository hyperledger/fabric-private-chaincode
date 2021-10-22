/*
 * Copyright 2021 Intel Corporation
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "study.h"
#include "storage.h"

bool Contract::Study::store(Contract::Storage& storage, const std::string& studyDetailsB64)
{
    FAST_FAIL_CHECK(er_, EC_ERROR, studyId_.empty());
    std::string studyKey("study." + studyId_);

    storage.ledgerPrivatePutString(studyKey, studyDetailsB64);

    return true;
}

bool Contract::Study::retrieve(Contract::Storage& storage, std::string& studyDetailsB64)
{
    FAST_FAIL_CHECK(er_, EC_ERROR, studyId_.empty());
    std::string studyKey("study." + studyId_);
    studyDetailsB64.empty();
    storage.ledgerPrivateGetString(studyKey, studyDetailsB64);

    if (studyDetailsB64.length() > 0)
    {
        return true;
    }
    return false;
}
