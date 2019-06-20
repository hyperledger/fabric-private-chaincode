/*
 * Copyright IBM Corp. 2018 All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

#pragma once

#include <map>
#include <set>
#include <string>
#include <vector>

typedef std::map<std::string, std::string> write_set_t;
typedef std::set<std::string> read_set_t;
typedef std::map<void*, std::pair<read_set_t*, write_set_t*>> context_t;

// shim put/get
void get_state(const char* key, uint8_t* val, uint32_t max_val_len, uint32_t* val_len, void* ctx);
void put_state(const char* key, uint8_t* val, uint32_t val_len, void* ctx);
void get_state_by_partial_composite_key(
    const char* comp_key, std::map<std::string, std::string>& values, void* ctx);

int unmarshal_args(std::vector<std::string>& argss, const char* json_string);
int unmarshal_values(
    std::map<std::string, std::string>& values, const char* json_bytes, uint32_t json_len);

// read/writeset
void register_rwset(void* ctx, read_set_t* readset, write_set_t* writeset);
void free_rwset(void* ctx);
read_set_t* get_read_set(context_t* context, void* ctx);
write_set_t* get_write_set(context_t* context, void* ctx);
