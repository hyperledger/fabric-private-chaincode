/*
 * Copyright 2020 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include <string>
#include "error.h"
#include "logging.h"
#include "sgx_eid.h"
#include "sgx_error.h"
#include "sgx_urts.h"
#include "test-defines.h"
#include "test-utils.h"
#include "test_enclave_u.h"

int main()
{
    sgx_launch_token_t token = {0};
    int updated = 0;
    sgx_enclave_id_t global_eid = 0;
    sgx_status_t ret = SGX_ERROR_UNEXPECTED;
    const uint32_t buffer_length = 1 << 20;
    uint8_t attestation[buffer_length];
    uint32_t attestation_length = 0;
    uint32_t params_length = 0;
    int b;
    char params_buf[buffer_length];
    std::string params;

    ret = sgx_create_enclave(ENCLAVE_FILENAME, SGX_DEBUG_FLAG, &token, &updated, &global_eid, NULL);
    if (ret != SGX_SUCCESS)
    {
        puts("error creating enclave");
        exit(-1);
    }

    // set logging callback
    logging_set_callback(puts);

    LOG_DEBUG("get_attestation_app test -- logging enabled");

    COND2LOGERR(false == load_file(INIT_DATA_INPUT, params_buf, buffer_length, &params_length),
        "error loading params");
    params = std::string(params_buf, params_length);

    LOG_INFO("Testing init attestation\n");
    init_att(global_eid, &b, (uint8_t*)params.c_str(), params.length());
    COND2LOGERR(!b, "init_attestation failed");

    LOG_INFO("Testing get attestation\n");
    get_att(global_eid, &b, (uint8_t*)STATEMENT, strlen(STATEMENT), attestation, 4096,
        &attestation_length);
    COND2LOGERR(!b, "get_attestation failed");

    COND2LOGERR(false == save_file(GET_ATTESTATION_OUTPUT, (char*)attestation, attestation_length),
        "error saving attestation");
    sgx_destroy_enclave(global_eid);

    LOG_INFO("Test Successful\n");
    return 0;

err:
    sgx_destroy_enclave(global_eid);
    return -1;
}
