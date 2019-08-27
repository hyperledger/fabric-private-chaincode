/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#pragma once

#include <cstdint>
#include <string>
#include <vector>

int marshal_ecc_args(std::vector<std::string>& argss,
    std::string& encoded_string);  // assumed to be initialized but empty

int unmarshal_ecc_response(const uint8_t* json_bytes,
    uint32_t json_len,
    uint8_t* response_data,
    uint32_t* response_len,
    uint8_t* signature,
    uint32_t* signature_len,
    uint8_t* pk,
    uint32_t* pk_len);
