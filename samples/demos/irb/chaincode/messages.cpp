/*
 * Copyright 2019 Intel Corporation
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

// TODO: This should go at compile time
#define PB_ENABLE_MALLOC

#include "messages.h"
#include <mbusafecrt.h> /* for memcpy_s etc */
#include <pb.h>
#include <pb_decode.h>
#include <pb_encode.h>
#include <string>
#include "_protos/irb.pb.h"

Contract::EASMessage::EASMessage() {}

Contract::EASMessage::EASMessage(const std::string& message) : inputString_(message)
{
    // base64-decode every message at the beginning
    inputMessageBytes_ = Base64EncodedStringToByteArray(inputString_);
}

Contract::EASMessage::EASMessage(const ByteArray& messageBytes) : inputMessageBytes_(messageBytes)
{
}

ByteArray Contract::EASMessage::getInputMessageBytes()
{
    return inputMessageBytes_;
}

bool Contract::EASMessage::toStatus(const std::string& message, int rc, std::string& outputMessage)
{
    Status status;
    int ret;
    bool b;

    if (message.empty())
    {
        status.msg = NULL;
    }
    else
    {
        status.msg = (char*)pb_realloc(status.msg, message.length() + 1);
        FAST_FAIL_CHECK(er_, EC_ERROR, status.msg == NULL);
        ret = memcpy_s(status.msg, message.length(), message.c_str(), message.length());
        FAST_FAIL_CHECK(er_, EC_ERROR, ret != 0);
        status.msg[message.length()] = '\0';
    }

    status.return_code = (Status_ReturnCode)rc;

    // encode in protobuf
    pb_ostream_t ostream;
    uint32_t response_len = 1024;
    uint8_t response[response_len];
    ostream = pb_ostream_from_buffer(response, response_len);
    b = pb_encode(&ostream, Status_fields, &status);
    FAST_FAIL_CHECK(er_, EC_ERROR, !b);

    // once encoded on buffer, release dynamic fields
    pb_release(Status_fields, &status);

    // base64-encode
    outputMessage =
        ByteArrayToBase64EncodedString(ByteArray(response, response + ostream.bytes_written));

    return true;
}

bool Contract::EASMessage::fromRegisterDataRequest(
    std::string& uuid, ByteArray& publicKey, ByteArray& decryptionKey, std::string& dataHandler)
{
    pb_istream_t istream;
    bool b;
    RegisterDataRequest registerDataRequest;

    istream = pb_istream_from_buffer(
        (const unsigned char*)inputMessageBytes_.data(), inputMessageBytes_.size());
    b = pb_decode(&istream, RegisterDataRequest_fields, &registerDataRequest);
    FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, !b);
    FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, !registerDataRequest.has_participant);
    FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, registerDataRequest.participant.uuid == NULL);
    FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, registerDataRequest.data_handler == NULL);
    FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, registerDataRequest.decryption_key == NULL);
    FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, registerDataRequest.participant.public_key == NULL);

    uuid = std::string(registerDataRequest.participant.uuid);
    dataHandler = std::string(registerDataRequest.data_handler);

    publicKey = ByteArray(registerDataRequest.participant.public_key->bytes,
        registerDataRequest.participant.public_key->bytes +
            registerDataRequest.participant.public_key->size);

    decryptionKey = ByteArray(registerDataRequest.decryption_key->bytes,
        registerDataRequest.decryption_key->bytes + registerDataRequest.decryption_key->size);

    return true;
}

void _fromIdentityToContractId(Identity& identity, Contract::Id& contractId)
{
    contractId.uuid_ = std::string((identity.uuid == NULL ? "" : identity.uuid));

    if (identity.public_key != NULL)
    {
        contractId.publicKey_ = ByteArray(
            identity.public_key->bytes, identity.public_key->bytes + identity.public_key->size);
    }

    if (identity.public_encryption_key != NULL)
    {
        contractId.publicEncryptionKey_ = ByteArray(identity.public_encryption_key->bytes,
            identity.public_encryption_key->bytes + identity.public_encryption_key->size);
    }
}

bool Contract::EASMessage::fromIdentity(Contract::Id& contractId)
{
    pb_istream_t istream;
    bool b;
    Identity identity;

    istream = pb_istream_from_buffer(
        (const unsigned char*)inputMessageBytes_.data(), inputMessageBytes_.size());
    b = pb_decode(&istream, Identity_fields, &identity);
    FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, !b);

    _fromIdentityToContractId(identity, contractId);

    return true;
}

bool Contract::EASMessage::fromNewExperiment(Contract::Experiment& experiment)
{
    pb_istream_t istream;
    bool b;
    ExperimentProposal experimentProposal;

    istream = pb_istream_from_buffer(
        (const unsigned char*)inputMessageBytes_.data(), inputMessageBytes_.size());
    b = pb_decode(&istream, ExperimentProposal_fields, &experimentProposal);
    FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, !b);
    FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, experimentProposal.study_id == NULL);
    FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, experimentProposal.experiment_id == NULL);
    FAST_FAIL_CHECK(
        er_, EC_INVALID_INPUT, experimentProposal.worker_credentials.identity_bytes == NULL);

    experiment.studyId_ = std::string(experimentProposal.study_id);
    experiment.experimentId_ = std::string(experimentProposal.experiment_id);

    ByteArray identityBytes(experimentProposal.worker_credentials.identity_bytes->bytes,
        experimentProposal.worker_credentials.identity_bytes->bytes +
            experimentProposal.worker_credentials.identity_bytes->size);
    Contract::EASMessage m(identityBytes);
    b = m.fromIdentity(experiment.workerId_);
    FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, !b);

    return true;
}

bool Contract::EASMessage::fromGetExperimentRequest(Contract::Experiment& experiment)
{
    pb_istream_t istream;
    bool b;
    GetExperimentRequest getExperimentRequest;

    istream = pb_istream_from_buffer(
        (const unsigned char*)inputMessageBytes_.data(), inputMessageBytes_.size());
    b = pb_decode(&istream, GetExperimentRequest_fields, &getExperimentRequest);
    FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, !b);
    FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, getExperimentRequest.experiment_id == NULL);

    experiment.experimentId_ = std::string(getExperimentRequest.experiment_id);
    return true;
}

bool Contract::EASMessage::fromStudyDetails(Contract::Study& study)
{
    pb_istream_t istream;
    bool b;
    StudyDetailsMessage studyDetailsMessage;

    istream = pb_istream_from_buffer(
        (const unsigned char*)inputMessageBytes_.data(), inputMessageBytes_.size());
    b = pb_decode(&istream, StudyDetailsMessage_fields, &studyDetailsMessage);
    FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, !b);
    FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, studyDetailsMessage.study_id == NULL);

    study.studyId_ = std::string(studyDetailsMessage.study_id);
    study.metadata_ =
        std::string(studyDetailsMessage.metadata == NULL ? "" : studyDetailsMessage.metadata);

    for (int i = 0; i < studyDetailsMessage.user_identities_count; i++)
    {
        Contract::Id id;
        _fromIdentityToContractId(studyDetailsMessage.user_identities[i], id);
        study.userIds_.push_back(id);
    }

    return true;
}

bool Contract::EASMessage::fromSignedApproval(Contract::SignedApproval& signedApproval)
{
    pb_istream_t istream;
    bool b;
    SignedApprovalMessage sa;

    istream = pb_istream_from_buffer(
        (const unsigned char*)inputMessageBytes_.data(), inputMessageBytes_.size());
    b = pb_decode(&istream, SignedApprovalMessage_fields, &sa);
    FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, !b);
    FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, sa.approval == NULL);

    // get the approval bytes
    signedApproval.approvalBytes_ =
        ByteArray(sa.approval->bytes, sa.approval->bytes + sa.approval->size);
    // get signature (if any)
    if (sa.signature != NULL)
    {
        signedApproval.signature_ =
            ByteArray(sa.signature->bytes, sa.signature->bytes + sa.signature->size);
    }

    {
        // unmashal approval
        pb_istream_t istream;
        Approval a;

        istream = pb_istream_from_buffer(
            signedApproval.approvalBytes_.data(), signedApproval.approvalBytes_.size());
        b = pb_decode(&istream, Approval_fields, &a);
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, !b);
        FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, a.experiment_id == NULL);

        signedApproval.experimentId_ = std::string(a.experiment_id);

        if (a.decision == Approval_Decision_APPROVED)
        {
            // approved == false, means rejected or undefined
            signedApproval.approved_ = true;
        }

        if (a.has_approver)
        {
            _fromIdentityToContractId(a.approver, signedApproval.approverId_);
        }
    }

    return true;
}

bool Contract::EASMessage::fromEvaluationPackRequest(Contract::EvaluationPack& evaluationPack)
{
    pb_istream_t istream;
    bool b;
    EvaluationPackRequest evaluationPackRequest;

    istream = pb_istream_from_buffer(
        (const unsigned char*)inputMessageBytes_.data(), inputMessageBytes_.size());
    b = pb_decode(&istream, EvaluationPackRequest_fields, &evaluationPackRequest);
    FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, !b);
    FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, evaluationPackRequest.experiment_id == NULL);

    evaluationPack.experimentId_ = std::string(evaluationPackRequest.experiment_id);
    return true;
}
