/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#pragma once

#include <string>
#include "shim.h"

std::string init_auction_house(std::string auction_house_name, shim_ctx_ptr_t ctx);
std::string auction_create(std::string auction_name, shim_ctx_ptr_t ctx);
std::string auction_submit(
    std::string auction_name, std::string bidder_name, int value, shim_ctx_ptr_t ctx);
std::string auction_eval(std::string auction_name, shim_ctx_ptr_t ctx);
std::string auction_close(std::string auction_name, shim_ctx_ptr_t ctx);
