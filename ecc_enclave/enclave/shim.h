/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#pragma once

#include <map>
#include <string>
#include <vector>
#include <stdbool.h>

typedef void* fpc_ctx_t;


// Function which FPC chaincode has to implement
// ==================================================
int invoke(const char* args,
           uint8_t* response,
           uint32_t max_response_len,
           uint32_t* actual_response_len,
           fpc_ctx_t ctx);


// Shim Function which FPC chaincode can use
// ==================================================

// put/get state
//-------------------------------------------------
// TODO: documention, e.g., how are error handled?
void get_state(const char* key, uint8_t* val, uint32_t max_val_len, uint32_t* val_len, fpc_ctx_t ctx);
void put_state(const char* key, uint8_t* val, uint32_t val_len, fpc_ctx_t ctx);
void get_state_by_partial_composite_key(
    const char* comp_key, std::map<std::string, std::string>& values, fpc_ctx_t ctx);

// (un)marshalling
//-------------------------------------------------
int unmarshal_args(std::vector<std::string>& argss, const char* json_string);
int unmarshal_values(
    std::map<std::string, std::string>& values, const char* json_bytes, uint32_t json_len);

// transaction creator
//-------------------------------------------------

// GetCreator returns `SignatureHeader.Creator` (e.g. an identity)
// of the `SignedProposal`. This is the identity of the agent (or user)
// submitting the transaction.
// Note: caller will have to free memory after use ..
char* get_creator();


// logging
//-------------------------------------------------
// TODO: add ....


