/*
 * Copyright 2020 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#pragma once

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

bool verify_evidence(uint8_t* evidence,
    uint32_t evidence_length,
    uint8_t* expected_statement,
    uint32_t expected_statement_length,
    uint8_t* expected_code_id,
    uint32_t expected_code_id_length);

#ifdef __cplusplus
}
#endif
