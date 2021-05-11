/*
 * Copyright 2019 Intel Corporation
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "messages.h"
#include <pb_decode.h>
#include <pb_encode.h>
#include <string>
#include "_protos/irb.pb.h"

Contract::EASMessage::EASMessage() {}

Contract::EASMessage::EASMessage(const std::string& message) : inputString_(message) {}
