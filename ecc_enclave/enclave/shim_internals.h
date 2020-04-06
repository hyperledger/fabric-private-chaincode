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
