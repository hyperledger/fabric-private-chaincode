/*
 * Copyright 2021 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

// TODO: This should go at compile time
#define PB_ENABLE_MALLOC

#include "evaluationpack.h"
#include <pb.h>
#include <pb_decode.h>
#include <pb_encode.h>
#include "_protos/irb.pb.h"
#include "experiment.h"
#include "messages.h"
#include "study.h"

bool Contract::EvaluationPack::build(
    Contract::Storage& storage, std::string& encryptedEvaluationPackB64)
{
    Contract::Experiment experiment;
    Contract::Study study;
    std::string storedExperiment, storedStudy;
    EvaluationPackMessage epm;
    ByteArray evaluationPackMessageBytes;
    ByteArray encryptedEvaluationPackBytes;
    EncryptedEvaluationPack eep;
    bool b;

    // check that experiment exist
    experiment.experimentId_ = experimentId_;
    b = experiment.retrieve(storage, storedExperiment);
    FAST_FAIL_CHECK(er_, EC_ERROR, !b);
    {
        Contract::EASMessage icm(storedExperiment);
        b = icm.fromNewExperiment(experiment);
        FAST_FAIL_CHECK(er_, EC_ERROR, !b);
    }

    // check that study exist
    study.studyId_ = experiment.studyId_;
    b = study.retrieve(storage, storedStudy);
    FAST_FAIL_CHECK(er_, EC_ERROR, !b);
    {
        Contract::EASMessage icm(storedStudy);
        b = icm.fromStudyDetails(study);
        FAST_FAIL_CHECK(er_, EC_ERROR, !b);
    }

    // check if experiment is approved; otherwise abort
    // TODO

    // create evaluation pack
    // get all participant IDs from study
    epm.registered_data_count = study.userIds_.size();
    epm.registered_data = (RegisterDataRequest*)pb_realloc(
        NULL, epm.registered_data_count * sizeof(RegisterDataRequest));
    FAST_FAIL_CHECK(er_, EC_ERROR, epm.registered_data == NULL);

    // encode registered data fields (only the decryption key!!)
    for (int i = 0; i < epm.registered_data_count; i++)
    {
        epm.registered_data[i] = RegisterDataRequest_init_zero;

        // retrieve decryption key
        ByteArray decryptionKey;
        std::string decryptionKeyB64;
        storage.ledgerPrivateGetString(
            "user." + study.userIds_[i].uuid_ + ".data.dk", decryptionKeyB64);
        FAST_FAIL_CHECK(er_, EC_ERROR, decryptionKeyB64.empty());
        decryptionKey = Base64EncodedStringToByteArray(decryptionKeyB64);

        // encode key in proto struct
        epm.registered_data[i].decryption_key =
            (pb_bytes_array_t*)pb_realloc(NULL, PB_BYTES_ARRAY_T_ALLOCSIZE(decryptionKey.size()));
        FAST_FAIL_CHECK(er_, EC_ERROR, epm.registered_data[i].decryption_key == NULL);
        epm.registered_data[i].decryption_key->size = decryptionKey.size();
        memcpy(epm.registered_data[i].decryption_key->bytes, decryptionKey.data(),
            decryptionKey.size());

        // retrieve data handler
        std::string dataHandler;
        storage.ledgerPrivateGetString(
            "user." + study.userIds_[i].uuid_ + ".data.handler", dataHandler);
        FAST_FAIL_CHECK(er_, EC_ERROR, dataHandler.empty());

        // encode data handler in proto struct
        epm.registered_data[i].data_handler = (char*)pb_realloc(NULL, dataHandler.length() + 1);
        FAST_FAIL_CHECK(er_, EC_ERROR, epm.registered_data[i].data_handler == NULL);
        memcpy(epm.registered_data[i].data_handler, dataHandler.c_str(), dataHandler.length());
        epm.registered_data[i].data_handler[dataHandler.length()] = '\0';
    }

    {
        // get encoding size
        size_t estimated_size;
        b = pb_get_encoded_size(&estimated_size, EvaluationPackMessage_fields, &epm);
        FAST_FAIL_CHECK(er_, EC_ERROR, !b);

        // encode evaluation pack
        evaluationPackMessageBytes.resize(estimated_size);

        pb_ostream_t ostream;
        ostream = pb_ostream_from_buffer(
            evaluationPackMessageBytes.data(), evaluationPackMessageBytes.size());
        b = pb_encode(&ostream, EvaluationPackMessage_fields, &epm);
        FAST_FAIL_CHECK(er_, EC_ERROR, !b);
        FAST_FAIL_CHECK(er_, EC_ERROR, evaluationPackMessageBytes.size() != ostream.bytes_written);

        pb_release(EvaluationPackMessage_fields, &epm);
    }

    // encrypt and authenticate using workerPK
    // TODO

    {
        eep = EncryptedEvaluationPack_init_zero;

        // encode encrypted pack
        eep.encrypted_evaluationpack = (pb_bytes_array_t*)pb_realloc(
            NULL, PB_BYTES_ARRAY_T_ALLOCSIZE(evaluationPackMessageBytes.size()));
        FAST_FAIL_CHECK(er_, EC_ERROR, eep.encrypted_evaluationpack == NULL);
        eep.encrypted_evaluationpack->size = evaluationPackMessageBytes.size();
        memcpy(eep.encrypted_evaluationpack->bytes, evaluationPackMessageBytes.data(),
            evaluationPackMessageBytes.size());

        // get encoding size
        size_t estimated_size;
        b = pb_get_encoded_size(&estimated_size, EncryptedEvaluationPack_fields, &eep);
        FAST_FAIL_CHECK(er_, EC_ERROR, !b);

        encryptedEvaluationPackBytes.resize(estimated_size);

        pb_ostream_t ostream;
        ostream = pb_ostream_from_buffer(
            encryptedEvaluationPackBytes.data(), encryptedEvaluationPackBytes.size());
        b = pb_encode(&ostream, EncryptedEvaluationPack_fields, &eep);
        FAST_FAIL_CHECK(er_, EC_ERROR, !b);
        FAST_FAIL_CHECK(
            er_, EC_ERROR, encryptedEvaluationPackBytes.size() != ostream.bytes_written);

        pb_release(EncryptedEvaluationPack_fields, &eep);
    }

    encryptedEvaluationPackB64 = ByteArrayToBase64EncodedString(encryptedEvaluationPackBytes);

    // return evaluation pack
    return true;
}
