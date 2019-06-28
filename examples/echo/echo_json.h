/*
 * Copyright 2019 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#pragma once

#include <stdint.h>
#include <string>

typedef struct echo
{
    std::string echo_string;
} echo_t;

int unmarshal(echo_t* echo, const char* json_bytes, uint32_t json_len);
std::string marshal(echo_t* echo);
