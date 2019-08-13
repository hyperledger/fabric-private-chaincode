/*
 * Copyright IBM Corp. All Rights Reserved.
 * Copyright Intel Corp. All Rights Reserved.
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
typedef std::map<void*, std::pair<read_set_t*, write_set_t*>> context_t;

void register_rwset(void* ctx, read_set_t* readset, write_set_t* writeset);
void free_rwset(void* ctx);
read_set_t* get_read_set(context_t* context, void* ctx);
write_set_t* get_write_set(context_t* context, void* ctx);
