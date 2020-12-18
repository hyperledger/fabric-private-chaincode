/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "enclave_t.h"

#include "base64.h"
#include "logging.h"
#include "utils.h"

#include <assert.h>
#include <string>

#include "sgx_utils.h"

#include "attestation-api/attestation/attestation.h"
#include "error.h"

#include <pb_decode.h>
#include "protos/fpc/fpc.pb.h"

#include "cc_data.h"

// enclave sk and pk (both are little endian) used for out signatures
sgx_ec256_private_t enclave_sk = {0};
sgx_ec256_public_t enclave_pk = {0};

// creates new identity if not exists
int ecall_init(const uint8_t* attestation_parameters,
    uint32_t ap_size,
    const uint8_t* cc_parameters,
    uint32_t ccp_size,
    const uint8_t* host_parameters,
    uint32_t hp_size,
    uint8_t* credentials,
    uint32_t credentials_max_size,
    uint32_t* credentials_size)
{
    bool b;

    // create new pub/prv key pair
    sgx_ecc_state_handle_t ecc_handle = NULL;
    sgx_status_t sgx_ret = sgx_ecc256_open_context(&ecc_handle);
    if (sgx_ret != SGX_SUCCESS)
    {
        LOG_DEBUG("Enclave: sgx_ecc256_open_context: %d", sgx_ret);
        return sgx_ret;
    }

    // create pub and private signature key
    sgx_ret = sgx_ecc256_create_key_pair(&enclave_sk, &enclave_pk, ecc_handle);
    if (sgx_ret != SGX_SUCCESS)
    {
        LOG_DEBUG("Enclave: sgx_ecc256_create_key_pair: %d", sgx_ret);
        return sgx_ret;
    }
    sgx_ecc256_close_context(ecc_handle);

    std::string base64_pk =
        base64_encode((const unsigned char*)&enclave_pk, sizeof(sgx_ec256_public_t));
    LOG_DEBUG("Enc: Enclave pk (little endian): %s", base64_pk.c_str());

    LOG_DEBUG("Enc: Identity generated!");

    // NOTE:
    // cc_data is a global pointer, meant to reference a global variable of cc_data type.
    // If g_cc_data is implemented as a simple global variable, cgo seems to crash (on enclave
    // destroy). This seems due to constructor/destructor issues -- if the variable is declared
    // in a function, it works. For this reason, we allocate it here dynamically.
    // TODO: free this memory when necessary.
    COND2LOGERR(g_cc_data != NULL, "cc data already created");

    g_cc_data = new cc_data;
    COND2LOGERR(g_cc_data == NULL, "error creating cc data object");

    b = g_cc_data->generate();
    COND2LOGERR(!b, "error generating cc data");

    // if a credential buffer was provided, then get credentials, else ignore
    if (credentials_max_size > 0)
    {
        b = g_cc_data->get_credentials(attestation_parameters, ap_size, cc_parameters, ccp_size,
            host_parameters, hp_size, credentials, credentials_max_size, credentials_size);
        COND2LOGERR(!b, "error getting credentials");
    }

    LOG_DEBUG("init enclave successful");
    return SGX_SUCCESS;

err:
    return SGX_ERROR_UNEXPECTED;
}

// returns report (containing enclave pk hash) and enclave pk in big endian format
int ecall_create_report(
    const sgx_target_info_t* target, sgx_report_t* report_out, uint8_t* pubkey_out)
{
    sgx_report_t report;
    sgx_report_data_t report_data = {{0}};

    memset(&report, 0, sizeof(report));

    // transform enclave_pk to Big Endian before hashing
    uint8_t enclave_pk_be[sizeof(sgx_ec256_public_t)];
    memcpy(enclave_pk_be, &enclave_pk, sizeof(sgx_ec256_public_t));
    bytes_swap(enclave_pk_be, 32);
    bytes_swap(enclave_pk_be + 32, 32);

    // write H(enclave_pk) in report data
    assert(sizeof(report_data) >= sizeof(sgx_sha256_hash_t));
    sgx_sha256_msg(enclave_pk_be, sizeof(sgx_ec256_public_t), (sgx_sha256_hash_t*)&report_data);

    // copy enclave_pk_be outside
    memcpy(pubkey_out, enclave_pk_be, sizeof(sgx_ec256_public_t));

    // create the report
    sgx_status_t ret = sgx_create_report(target, &report_data, &report);
    if (ret != SGX_SUCCESS)
    {
        LOG_ERROR("Enclave: Error while creating report");
        return ret;
    }
    memcpy(report_out, &report, sizeof(sgx_report_t));

    LOG_DEBUG("Enc: Report generated!");
    return SGX_SUCCESS;
}
