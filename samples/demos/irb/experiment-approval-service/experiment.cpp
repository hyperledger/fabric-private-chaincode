/*
 * Copyright 2021 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "experiment.h"

Contract::Experiment::Experiment() {}

Contract::Experiment::Experiment(
    std::string studyId, std::string experimentId, Contract::Id& workerId)
    : studyId_(studyId), experimentId_(experimentId), workerId_(workerId)
{
}

bool Contract::Experiment::store(
    Contract::Storage& storage, const std::string& experimentProposalB64)
{
    FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, experimentId_.empty());

    std::string experimentKey("experiment." + experimentId_);

    storage.ledgerPrivatePutString(experimentKey, experimentProposalB64);

    return true;
}

bool Contract::Experiment::retrieve(Contract::Storage& storage, std::string& experimentProposalB64)
{
    FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, experimentId_.empty());

    std::string experimentKey("experiment." + experimentId_);
    experimentProposalB64.empty();
    storage.ledgerPrivateGetString(experimentKey, experimentProposalB64);

    if (experimentProposalB64.length() > 0)
    {
        return true;
    }
    return false;
}
