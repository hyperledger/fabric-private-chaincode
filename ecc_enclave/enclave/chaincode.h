/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#pragma once

int invoke_enc(const char* args,
    const char* pk,
    uint8_t* response,
    uint32_t max_response_len,
    uint32_t* actual_response_len,
    void* ctx);
int invoke(const char* args,
    uint8_t* response,
    uint32_t max_response_len,
    uint32_t* actual_response_len,
    void* ctx);
