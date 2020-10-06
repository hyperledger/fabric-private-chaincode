/*
 * Copyright 2020 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "test.h"
#include <string>
#include "error.h"
#include "logging.h"
#include "test-defines.h"
#include "test-utils.h"

#include "attestation_tags.h"
#include "verify-evidence.h"

bool test()
{
    uint32_t buffer_length = 1 << 20;
    char buffer[buffer_length];
    uint32_t filled_size;
    std::string jsonevidence;
    std::string expected_statement;
    std::string expected_code_id;
    std::string wrong_expected_statement;
    std::string wrong_expected_code_id;

    // set the callback to dump log on stdout
    logging_set_callback(puts);

    LOG_DEBUG("verify_evidence_app test -- logging enabled");

    COND2LOGERR(!load_file(EVIDENCE_FILE, buffer, buffer_length, &filled_size),
        "can't read input evidence " EVIDENCE_FILE);
    jsonevidence = std::string(buffer, filled_size);

    COND2LOGERR(!load_file(STATEMENT_FILE, buffer, buffer_length, &filled_size),
        "can't read input statement " STATEMENT_FILE);
    expected_statement = std::string(buffer, filled_size);

    COND2LOGERR(!load_file(CODE_ID_FILE, buffer, buffer_length, &filled_size),
        "can't read input code id " CODE_ID_FILE);
    expected_code_id = std::string(buffer, filled_size);

    wrong_expected_statement = std::string("wrong statement");
    wrong_expected_code_id =
        std::string("BADBADBADBAD9E317C4F7312A0D644FFC052F7645350564D43586D8102663358");

    bool b, expected_b;
    // test normal situation
    b = verify_evidence((uint8_t*)jsonevidence.c_str(), jsonevidence.length(),
        (uint8_t*)expected_statement.c_str(), expected_statement.length(),
        (uint8_t*)expected_code_id.c_str(), expected_code_id.length());
    COND2LOGERR(!b, "correct evidence failed");

    // this test succeeds for simulated attestations, and fails for real ones
    // test with wrong statement
    expected_b = (jsonevidence.find(SIMULATED_TYPE_TAG) == std::string::npos ? false : true);
    if (expected_b == false)
    {
        LOG_INFO("next test expected to fail");
    }
    b = verify_evidence((uint8_t*)jsonevidence.c_str(), jsonevidence.length(),
        (uint8_t*)wrong_expected_statement.c_str(), wrong_expected_statement.length(),
        (uint8_t*)expected_code_id.c_str(), expected_code_id.length());
    COND2LOGERR(b != expected_b, "evidence with bad statement succeeded");

    // this test succeeds for simulated attestations, and fails for real ones
    // test with wrong code id
    expected_b = (jsonevidence.find(SIMULATED_TYPE_TAG) == std::string::npos ? false : true);
    if (expected_b == false)
    {
        LOG_INFO("next test expected to fail");
    }
    b = verify_evidence((uint8_t*)jsonevidence.c_str(), jsonevidence.length(),
        (uint8_t*)expected_statement.c_str(), expected_statement.length(),
        (uint8_t*)wrong_expected_code_id.c_str(), wrong_expected_code_id.length());
    COND2LOGERR(b != expected_b, "evidence with bad code id succeeded");

    LOG_INFO("Test Successful\n");
    return true;

err:
    return false;
}
