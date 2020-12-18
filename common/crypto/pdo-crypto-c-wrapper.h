/*
 * Copyright 2020 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#pragma once

#ifdef __cplusplus
extern "C" {
#endif

bool compute_hash(uint8_t* message,
    uint32_t message_len,
    uint8_t* hash,
    uint32_t max_hash_len,
    uint32_t* actual_hash_len);

bool verify_signature(uint8_t* public_key, uint32_t public_key_len, uint8_t* message, uint32_t message_len, uint8_t* signature, uint32_t signature_len);

#ifdef __cplusplus
}
#endif
