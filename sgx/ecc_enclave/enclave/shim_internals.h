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
#include "types.h"

// read/writeset
typedef std::map<std::string, ByteArray> write_set_t;
typedef std::map<std::string, ByteArray> read_set_t;
typedef std::set<std::string> del_set_t;

// shim context
typedef struct t_shim_ctx
{
    void* u_shim_ctx;
    read_set_t read_set;
    write_set_t write_set;
    del_set_t del_set;
    std::vector<std::string> string_args;
    ByteArray signed_proposal;
    std::string tx_id;
    std::string channel_id;
    ByteArray creator;
    std::string creator_msp_id;
    std::string creator_name;
    // TODO to be implemented
    // ByteArray binding;
    // std::map<std::string, ByteArray> transient_data;
} t_shim_ctx_t;

#include "fpc.pb.h"

bool rwset_to_proto(t_shim_ctx_t* ctx, fpc_FPCKVSet* fpc_rwset_proto);
