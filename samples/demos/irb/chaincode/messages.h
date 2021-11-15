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
#include "evaluationpack.h"
#include "experiment.h"
#include "id.h"
#include "signedapproval.h"
#include "study.h"

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
    EASMessage(const ByteArray& messageBytes);
    ErrorReport getErrorReport();
    ByteArray getInputMessageBytes();

    bool toStatus(const std::string& message, int rc, std::string& outputMessage);

    bool fromIdentity(Contract::Id& contractId);
    bool fromRegisterDataRequest(std::string& uuid,
        ByteArray& publicKey,
        ByteArray& decryptionKey,
        std::string& dataHandler);
    bool fromNewExperiment(Contract::Experiment& experiment);
    bool fromGetExperimentRequest(Contract::Experiment& experiment);
    bool fromStudyDetails(Contract::Study& study);
    bool fromSignedApproval(Contract::SignedApproval& signedApproval);
    bool fromEvaluationPackRequest(Contract::EvaluationPack& evaluationPack);
};
}  // namespace Contract
