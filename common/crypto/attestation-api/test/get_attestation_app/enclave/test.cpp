/*
 * Copyright 2020 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "test.h"
#include "attestation-api/attestation/attestation.h"
#include "fpc-types.h"
#include "test_enclave_t.h"

int init_att(uint8_t* params, uint32_t params_length)
{
    return init_attestation(params, params_length);
}

int get_att(uint8_t* statement,
    uint32_t statement_length,
    uint8_t* attestation,
    uint32_t attestation_max_length,
    uint32_t* attestation_length)
{
    return get_attestation(
        statement, statement_length, attestation, attestation_max_length, attestation_length);
}
