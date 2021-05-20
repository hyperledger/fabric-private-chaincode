/*
 * Copyright IBM Corp. All Rights Reserved.
 * Copyright 2020 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#pragma once

#include <vector>
#include "fpc-types.h"

typedef std::vector<uint8_t> ByteArray;

int get_random_bytes(uint8_t* buffer, size_t length);

bool validate_key_length(const ByteArray key);

bool decrypt_message(const ByteArray key, const ByteArray& encrypted_message, ByteArray& message);

bool encrypt_message(const ByteArray key, const ByteArray& message, ByteArray& encrypted_message);

bool compute_message_hash(const ByteArray message, ByteArray& message_hash);
