/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#pragma once

#include <string>

std::string auction_create(std::string auction_name, void* ctx);
std::string auction_submit(std::string auction_name, std::string bidder_name, int value, void* ctx);
std::string auction_eval(std::string auction_name, void* ctx);
std::string auction_close(std::string auction_name, void* ctx);
