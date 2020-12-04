/*
 * Copyright IBM Corp. All Rights Reserved.
 * Copyright 2020 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "crypto.h"

#include "sgx_trts.h"

int get_random_bytes(uint8_t* buffer, size_t length)
{
    /* WARNING WARNING WARNING */
    /* WARNING WARNING WARNING */

    // the implementation of this function with SGX rand forces to have a single encalve endorser

    /* WARNING WARNING WARNING */
    /* WARNING WARNING WARNING */
    return sgx_read_rand(buffer, length);
}
