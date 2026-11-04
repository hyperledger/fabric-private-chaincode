/*
 * Copyright 2021 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "id.h"

Contract::Id::Id() {}

Contract::Id::Id(std::string& uuid, ByteArray& publicKey, ByteArray& publicEncryptionKey)
    : uuid_(uuid), publicKey_(publicKey), publicEncryptionKey_(publicEncryptionKey)
{
}
