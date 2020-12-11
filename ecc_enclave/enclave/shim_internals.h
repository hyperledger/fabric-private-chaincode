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
#include <vector>

// read/writeset
typedef std::map<std::string, std::string> write_set_t;
typedef std::set<std::string> read_set_t;

// shim context
typedef struct t_shim_ctx
{
    void* u_shim_ctx;
    read_set_t read_set;
    write_set_t write_set;
    std::vector<std::string> string_args;
} t_shim_ctx_t;
