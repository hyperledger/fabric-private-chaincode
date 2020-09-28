/*
 * Copyright 2020 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#pragma once

#include <stdint.h>

bool load_file(const char* filename, char* buffer, uint32_t buffer_length, uint32_t* written_bytes);
bool save_file(const char* filename, const char* buffer, uint32_t buffer_length);
