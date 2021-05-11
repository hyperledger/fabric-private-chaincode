/*
 * Copyright 2019 Intel Corporation
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "chaincode.h"
#include <string>
#include "errors.h"
#include "messages.h"

Contract::ExperimentApprovalService::ExperimentApprovalService(shim_ctx_ptr_t ctx)
{
    // set Study Approval Service verification key

    // set initial approver (investigator)

    // set Experimenter IDs
}

CONTRACT_API_PROTOTYPE(Contract::ExperimentApprovalService::registerData)
{
    // store participant_uuid <- <dec_key, data_handler>

    return true;
}

CONTRACT_API_PROTOTYPE(Contract::ExperimentApprovalService::registerStudy)
{
    // check study approval service signature (request.details.signature)

    // if study already exists abort

    // create study object with participants uuids

    // store study under studyID

    return true;
}

CONTRACT_API_PROTOTYPE(Contract::ExperimentApprovalService::newExperiment)
{
    // check that study ID exists;

    // (optional) check experimenter identity

    // check that experiment does not exist already

    // sanity check that experiment details are complete

    // check worker attestation (using workerPK, MRENCLAVE)

    // create experiment object, set status to pending

    // store experiment inside study object

    return true;
}

CONTRACT_API_PROTOTYPE(Contract::ExperimentApprovalService::getExperimentProposal)
{
    // check that experiment exists

    // get experiment from state and return

    return true;
}

CONTRACT_API_PROTOTYPE(Contract::ExperimentApprovalService::approveExperiment)
{
    // check experiment exists for approval
    // check check signature on approval
    // check approver is valid approver

    // if approval criteria reached; mark experiment as approved

    return true;
}

CONTRACT_API_PROTOTYPE(Contract::ExperimentApprovalService::requestEvaluationPack)
{
    // check that study and experiment exist

    // check if experiment is approved; otherwise abort

    // get all participant IDs from study

    // collect all <dec_key, data_handler> paris for each participant

    // create evaluation pack

    // encrypt and authenticate using workerPK

    // return evaluation pack

    return true;
}
