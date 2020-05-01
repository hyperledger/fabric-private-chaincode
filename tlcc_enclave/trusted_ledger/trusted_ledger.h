/*
 * Copyright 2019 Intel Corporation
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#ifndef _TRUSTED_LEDGER_H_
#define _TRUSTED_LEDGER_H_

#include "fpc-types.h"

#ifdef __cplusplus
extern "C" {
#endif

int tlcc_init_with_genesis(enclave_id_t eid, uint8_t* genesis, uint32_t genesis_size);

int tlcc_send_block(enclave_id_t eid, uint8_t* block, uint32_t block_size);

// this is for debugging
int tlcc_print_state(enclave_id_t eid);

int tlcc_get_state_metadata(enclave_id_t eid, const char* key, uint8_t* nonce, cmac_t* cmac);

int tlcc_get_multi_state_metadata(
    enclave_id_t eid, const char* comp_key, uint8_t* nonce, cmac_t* cmac);

#ifdef __cplusplus
}
#endif /* __cplusplus */

#endif /* !_TRUSTED_LEDGER_H_ */
