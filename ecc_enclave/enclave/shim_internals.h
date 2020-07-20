/*
 * Copyright IBM Corp. All Rights Reserved.
 * Copyright 2020 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#pragma once

#include <map>
#include <set>
#include <string>
#include "shim.h"  // for shim_ctx_ptr_t

#define MAX_CHANNEL_ID_LENGTH 1024
#define MAX_MSP_ID_LENGTH 1024

// read/writeset
typedef std::map<std::string, std::string> write_set_t;
typedef std::set<std::string> read_set_t;

// shim context
typedef struct t_shim_ctx
{
    void* u_shim_ctx;
    read_set_t read_set;
    write_set_t write_set;
    const char* encoded_args;  // args as passed from client-side shim, potentially encrypted
    const char* json_args;     // clear-text args from client-side shim
} t_shim_ctx_t;

bool internal_set_channel_id(char* channel_id, uint32_t channel_id_length);
bool internal_get_channel_id(char* channel_id, uint32_t max_channel_id_len, shim_ctx_ptr_t ctx);

bool internal_set_msp_id(char* msp_id, uint32_t msp_id_length);
bool internal_get_msp_id(char* msp_id, uint32_t max_msp_id_len, shim_ctx_ptr_t ctx);
