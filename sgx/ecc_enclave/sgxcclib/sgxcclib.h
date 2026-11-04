/*
 * Copyright 2019 Intel Corporation
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#ifndef _SGXCCLIB_H_
#define _SGXCCLIB_H_

#include "fpc-types.h"

#ifdef __cplusplus
extern "C" {
#endif

int sgxcc_invoke(enclave_id_t eid,
    const uint8_t* signed_proposal_proto_bytes,
    uint32_t signed_proposal_proto_bytes_len,
    const uint8_t* b64_chaincode_request_message,
    uint32_t b64_chaincode_request_message_len,
    uint8_t* b64_chaincode_response_message,
    uint32_t b64_chaincode_response_message_len_in,
    uint32_t* b64_chaincode_response_message_len_out,
    void* ctx);

#ifdef __cplusplus
}
#endif /* __cplusplus */

#endif /* !_SGXCCLIB_H_ */
