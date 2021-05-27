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

Contract::ExperimentApprovalService::ExperimentApprovalService(shim_ctx_ptr_t ctx) : storage_(ctx)
{
    // set Study Approval Service verification key

    // set initial approver (investigator)

    // set Experimenter IDs
}

CONTRACT_API_PROTOTYPE(Contract::ExperimentApprovalService::registerData)
{
    std::string uuid, dataHandler;
    ByteArray pk, dk;
    bool b;
    Contract::EASMessage icm(inputString);

    // get relevant fields
    b = icm.fromRegisterDataRequest(uuid, pk, dk, dataHandler);
    FAST_FAIL_CHECK_EX(er, &icm.er_, EC_INVALID_INPUT, !b);

    LOG_DEBUG("Participant uuid: %s", uuid);
    LOG_DEBUG("Participant pk: %s", std::string((char*)pk.data(), pk.size()));
    LOG_DEBUG("Data handler: %s", dataHandler);

    // Store all fields on the ledger
    storage_.ledgerPrivatePutString("user." + uuid + ".uuid", uuid);
    storage_.ledgerPrivatePutString(
        "user." + uuid + ".pk", std::string((char*)pk.data(), pk.size()));
    storage_.ledgerPrivatePutString("user." + uuid + ".data.handler", dataHandler);
    storage_.ledgerPrivatePutString(
        "user." + uuid + ".data.dk", ByteArrayToBase64EncodedString(dk));

    // Prepare status message
    Contract::EASMessage ocm;
    b = ocm.toStatus("", 1, outputString);
    FAST_FAIL_CHECK_EX(er, &ocm.er_, EC_INVALID_INPUT, !b);

    er.set(EC_SUCCESS, "");
    return true;
}

CONTRACT_API_PROTOTYPE(Contract::ExperimentApprovalService::registerStudy)
{
    Contract::Study study;
    bool b;
    Contract::EASMessage icm(inputString);
    std::string storedStudy;
    std::string returnString("");
    int status = 1;

    // get relevant fields
    b = icm.fromStudyDetails(study);
    FAST_FAIL_CHECK_EX(er, &icm.er_, EC_INVALID_INPUT, !b);

    LOG_DEBUG("Study id: %s", study.studyId_);
    LOG_DEBUG("Metadata: %s", study.metadata_);
    LOG_DEBUG("Users: %d", study.userIds_.size());

    b = study.retrieve(storage_, storedStudy);
    if (!b)
    {
        // TODO check study approval service signature (request.details.signature)
        b = study.store(storage_, inputString);
        FAST_FAIL_CHECK(er, EC_ERROR, !b);
    }
    else
    {
        // if study already exists abort
        returnString = "Study already registered";
        status = 0;
    }

    // prepare status message
    Contract::EASMessage ocm;
    b = ocm.toStatus(returnString, status, outputString);
    FAST_FAIL_CHECK_EX(er, &ocm.er_, EC_INVALID_INPUT, !b);

    er.set(EC_SUCCESS, "");

    return true;
}

CONTRACT_API_PROTOTYPE(Contract::ExperimentApprovalService::newExperiment)
{
    Contract::Experiment experiment;
    bool b;
    Contract::EASMessage icm(inputString);
    std::string storedExperiment;
    std::string returnString("");
    int status = 1;

    // get relevant fields
    b = icm.fromNewExperiment(experiment);
    FAST_FAIL_CHECK_EX(er, &icm.er_, EC_INVALID_INPUT, !b);

    LOG_DEBUG("Experiment id: %s", experiment.experimentId_);

    // If not already registered, store experiment
    b = experiment.retrieve(storage_, storedExperiment);
    if (!b)
    {
        // TODO(optional) check experimenter identity
        // TODO sanity check that experiment details are complete
        // TODO check worker attestation (using workerPK, MRENCLAVE)
        b = experiment.store(storage_, inputString);
        FAST_FAIL_CHECK(er, EC_ERROR, !b);
    }
    else
    {
        // check that experiment does not exist already
        returnString = "Experiment already registered";
        status = 0;
    }

    // prepare status message
    Contract::EASMessage ocm;
    b = ocm.toStatus(returnString, status, outputString);
    FAST_FAIL_CHECK_EX(er, &ocm.er_, EC_INVALID_INPUT, !b);

    er.set(EC_SUCCESS, "");
    return true;
}

CONTRACT_API_PROTOTYPE(Contract::ExperimentApprovalService::getExperimentProposal)
{
    Contract::Experiment experiment;
    bool b;
    Contract::EASMessage icm(inputString);
    std::string storedExperiment;
    std::string returnString("");
    int status = 1;

    b = icm.fromGetExperimentRequest(experiment);
    FAST_FAIL_CHECK_EX(er, &icm.er_, EC_INVALID_INPUT, !b);

    LOG_DEBUG("Request for experiment id: %s", experiment.experimentId_);

    // get experiment from state and return
    b = experiment.retrieve(storage_, storedExperiment);
    if (!b)
    {
        returnString = "Experiment not found";
        status = 0;
        // prepare status message
        Contract::EASMessage ocm;
        b = ocm.toStatus(returnString, status, outputString);
        FAST_FAIL_CHECK_EX(er, &ocm.er_, EC_INVALID_INPUT, !b);
    }
    else
    {
        // return the retrieve experiment (b64 experiment proposal)
        outputString = storedExperiment;
    }

    er.set(EC_SUCCESS, "");
    return true;
}

CONTRACT_API_PROTOTYPE(Contract::ExperimentApprovalService::approveExperiment)
{
    Contract::SignedApproval signedApproval;
    Contract::Experiment experiment;
    bool b;
    Contract::EASMessage icm(inputString);
    std::string storedExperiment;
    std::string returnString("");
    int status = 1;

    b = icm.fromSignedApproval(signedApproval);
    FAST_FAIL_CHECK_EX(er, &icm.er_, EC_INVALID_INPUT, !b);

    LOG_DEBUG("Approval for experiment id: %s", signedApproval.experimentId_);
    LOG_DEBUG(
        "Approval decision: %s", signedApproval.approved_ ? "approved" : "rejected/undefined");

    // check experiment exists for approval
    experiment.experimentId_ = signedApproval.experimentId_;
    b = experiment.retrieve(storage_, storedExperiment);
    if (!b)
    {
        returnString = "Experiment not found";
        status = 0;
    }
    else
    {
        // if approval criteria reached; mark experiment as approved
        b = signedApproval.store(storage_, inputString);
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, !b);
    }

    // prepare status message
    Contract::EASMessage ocm;
    b = ocm.toStatus(returnString, status, outputString);
    FAST_FAIL_CHECK_EX(er, &ocm.er_, EC_INVALID_INPUT, !b);

    er.set(EC_SUCCESS, "");
    return true;
}

CONTRACT_API_PROTOTYPE(Contract::ExperimentApprovalService::requestEvaluationPack)
{
    bool b;
    Contract::EASMessage icm(inputString);
    std::string storedExperiment;
    Contract::Experiment experiment;
    Contract::EvaluationPack evaluationPack;
    std::string returnString("");
    int status = 1;

    b = icm.fromEvaluationPackRequest(evaluationPack);
    FAST_FAIL_CHECK_EX(er, &icm.er_, EC_INVALID_INPUT, !b);

    LOG_DEBUG("Request for evaluation pack for experimend id: %s", evaluationPack.experimentId_);

    // check that study and experiment exist
    experiment.experimentId_ = evaluationPack.experimentId_;
    b = experiment.retrieve(storage_, storedExperiment);
    if (!b)
    {
        returnString = "Experiment not found";
        status = 0;
        // prepare status message
        Contract::EASMessage ocm;
        b = ocm.toStatus(returnString, status, outputString);
        FAST_FAIL_CHECK_EX(er, &ocm.er_, EC_INVALID_INPUT, !b);
    }
    else
    {
        b = evaluationPack.build(storage_, outputString);
        FAST_FAIL_CHECK_EX(er, &evaluationPack.er_, EC_ERROR, !b)
    }

    er.set(EC_SUCCESS, "");
    return true;
}
