/*
* Copyright IBM Corp. 2018 All Rights Reserved.
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

#include "enclave_t.h"

#include "base64.h"
#include "logging.h"
#include "utils.h"

#include <assert.h>
#include <string>

#include "sgx_utils.h"

// enclave sk and pk (both are little endian) used for out signatures
sgx_ec256_private_t enclave_sk = {0};
sgx_ec256_public_t enclave_pk = {0};

// creates new identity if not exists
int ecall_init(void)
{
    // create new pub/prv key pair
    sgx_ecc_state_handle_t ecc_handle = NULL;
    sgx_status_t sgx_ret = sgx_ecc256_open_context(&ecc_handle);
    if (sgx_ret != SGX_SUCCESS) {
        LOG_DEBUG("Enclave: sgx_ecc256_open_context: %d\n", sgx_ret);
        return sgx_ret;
    }

    // create pub and private signature key
    sgx_ret = sgx_ecc256_create_key_pair(&enclave_sk, &enclave_pk, ecc_handle);
    if (sgx_ret != SGX_SUCCESS) {
        LOG_DEBUG("Enclave: sgx_ecc256_create_key_pair: %d\n", sgx_ret);
        return sgx_ret;
    }
    sgx_ecc256_close_context(ecc_handle);

    std::string base64_pk =
        base64_encode((const unsigned char *)&enclave_pk, sizeof(sgx_ec256_public_t));
    LOG_DEBUG("Enc: Enclave pk (little endian): %s", base64_pk.c_str());

    LOG_DEBUG("Enc: Identity generated!");
    return SGX_SUCCESS;
}

// returns report (containing enclave pk hash) and enclave pk in big endian format
int ecall_create_report(
    const sgx_target_info_t *target, sgx_report_t *report_out, uint8_t *pubkey_out)
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
    sgx_sha256_msg(enclave_pk_be, sizeof(sgx_ec256_public_t), (sgx_sha256_hash_t *)&report_data);

    // copy enclave_pk_be outside
    memcpy(pubkey_out, enclave_pk_be, sizeof(sgx_ec256_public_t));

    // create the report
    sgx_status_t ret = sgx_create_report(target, &report_data, &report);
    if (ret != SGX_SUCCESS) {
        LOG_ERROR("Enclave: Error while creating report");
        return ret;
    }
    memcpy(report_out, &report, sizeof(sgx_report_t));

    LOG_DEBUG("Enc: Report generated!");
    return SGX_SUCCESS;
}

int ecall_get_target_info(sgx_target_info_t *target)
{
    if (!target) {
        LOG_ERROR("Enclave: Invalid input");
        return SGX_ERROR_INVALID_PARAMETER;
    }

    // create a report and extract target_info
    sgx_report_t temp_report;
    memset(&temp_report, 0, sizeof(temp_report));
    memset(target, 0, sizeof(sgx_target_info_t));

    sgx_status_t ret = sgx_create_report(NULL, NULL, &temp_report);
    if (ret != SGX_SUCCESS) {
        LOG_ERROR("Enclave: Error while creating report");
        return ret;
    }

    // target info size is 512
    memcpy(&target->mr_enclave, &temp_report.body.mr_enclave, sizeof(sgx_measurement_t));
    memcpy(&target->attributes, &temp_report.body.attributes, sizeof(sgx_attributes_t));
    memcpy(&target->misc_select, &temp_report.body.misc_select, sizeof(sgx_misc_select_t));

    LOG_DEBUG("Enc: TargetInfo created!");
    return SGX_SUCCESS;
}

// returns enclave pk in Big Endian format
int ecall_get_pk(uint8_t *pubkey)
{
    // transform enclave_pk to Big Endian before hashing
    uint8_t enclave_pk_be[sizeof(sgx_ec256_public_t)];
    memcpy(enclave_pk_be, &enclave_pk, sizeof(sgx_ec256_public_t));
    bytes_swap(enclave_pk_be, 32);
    bytes_swap(enclave_pk_be + 32, 32);

    memcpy(pubkey, &enclave_pk_be, sizeof(sgx_ec256_public_t));

    LOG_DEBUG("Enc: Return enclave pk as Big Endian");
    return SGX_SUCCESS;
}
