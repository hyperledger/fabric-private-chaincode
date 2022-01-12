/*
 * Copyright 2019 Intel Corporation
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "sgxcclib.h"
#include "check-sgx-error.h"  //CHECK_SGX_ERROR_AND_RETURN_ON_ERROR macro
#include "enclave_u.h"

#include <stdbool.h>
#include <string.h>

// - for accessing ledger kvs
extern void get_state(
    const char* key, uint8_t* val, uint32_t max_val_len, uint32_t* val_len, void* ctx);
extern void get_state_by_partial_composite_key(
    const char* comp_key, uint8_t* values, uint32_t max_len, uint32_t* values_len, void* ctx);
extern void put_state(const char* key, uint8_t* val, uint32_t val_len, void* ctx);
extern void del_state(const char* key, void* ctx);

int sgxcc_invoke(enclave_id_t eid,
    const uint8_t* signed_proposal_proto_bytes,
    uint32_t signed_proposal_proto_bytes_len,
    const uint8_t* b64_chaincode_request_message,
    uint32_t b64_chaincode_request_message_len,
    uint8_t* b64_chaincode_response_message,
    uint32_t b64_chaincode_response_message_len_in,
    uint32_t* b64_chaincode_response_message_len_out,
    void* ctx)
{
    int enclave_ret = SGX_ERROR_UNEXPECTED;
    int ret = ecall_cc_invoke(eid, &enclave_ret, signed_proposal_proto_bytes,
        signed_proposal_proto_bytes_len, b64_chaincode_request_message,
        b64_chaincode_request_message_len, b64_chaincode_response_message,
        b64_chaincode_response_message_len_in, b64_chaincode_response_message_len_out,
        ctx);  // context for callback
    CHECK_SGX_ERROR_AND_RETURN_ON_ERROR(ret)
    CHECK_SGX_ERROR_AND_RETURN_ON_ERROR(enclave_ret)
    return SGX_SUCCESS;
}

void ocall_get_state(
    const char* key, uint8_t* val, uint32_t max_val_len, uint32_t* val_len, void* ctx)
{
    get_state(key, val, max_val_len, val_len, ctx);
}

void ocall_put_state(const char* key, uint8_t* val, uint32_t val_len, void* ctx)
{
    put_state(key, val, val_len, ctx);
}

void ocall_get_state_by_partial_composite_key(
    const char* key, uint8_t* bids_bytes, uint32_t max_len, uint32_t* bids_bytes_len, void* ctx)
{
    get_state_by_partial_composite_key(key, bids_bytes, max_len, bids_bytes_len, ctx);
}

void ocall_del_state(const char* key, void* ctx)
{
    del_state(key, ctx);
}
