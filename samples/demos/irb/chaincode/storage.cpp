/*
 * Copyright 2021 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "storage.h"
#include "common.h"

#define MAX_LENGTH_STRING_SIZE 10  // order of gigabytes

Contract::Storage::Storage(shim_ctx_ptr_t ctx) : ctx_(ctx) {}

void Contract::Storage::ledgerPrivatePutString(const std::string& key, const std::string& value)
{
    // first write length -- string with null char
    unsigned int valueLength = value.length() + 1;
    std::string valueLengthString = std::to_string(valueLength);
    if (valueLengthString.length() == 0 || valueLengthString.length() > MAX_LENGTH_STRING_SIZE)
    {
        LOG_ERROR("value length is 0 or too large. No data will be written to ledger");
        return;
    }
    std::string valueLengthKey = key + "L";
    ledgerPrivatePutBinary((const uint8_t*)valueLengthKey.c_str(), valueLengthKey.length(),
        (const uint8_t*)valueLengthString.c_str(), valueLengthString.length() + 1);

    // second, write actual value
    std::string valueKey = key + "K";
    ledgerPrivatePutBinary(
        (uint8_t*)valueKey.c_str(), valueKey.length(), (const uint8_t*)value.c_str(), valueLength);
}

void Contract::Storage::ledgerPrivateGetString(const std::string& key, std::string& value)
{
    // first, get the value length
    uint8_t valueLengthArray[MAX_LENGTH_STRING_SIZE + 1];
    uint32_t actualValueLengthLength = 0;
    std::string valueLengthKey = key + "L";
    ledgerPrivateGetBinary((uint8_t*)valueLengthKey.c_str(), valueLengthKey.length(),
        valueLengthArray, MAX_LENGTH_STRING_SIZE + 1, &actualValueLengthLength);
    if (actualValueLengthLength == 0)
    {
        LOG_DEBUG("Key not found -- length not stored");
        return;
    }
    if (actualValueLengthLength > MAX_LENGTH_STRING_SIZE)
    {
        LOG_ERROR("Value length returned is too large");
        return;
    }
    if (valueLengthArray[actualValueLengthLength - 1] != '\0')
    {
        LOG_ERROR("Value length returned is not null terminated");
        return;
    }
    uint32_t storedValueLength = std::stoul(std::string((char*)valueLengthArray), NULL, 10);

    // second, get the value
    std::string valueKey = key + "K";
    uint8_t valueBinary[storedValueLength];
    uint32_t actualValueLength = 0;
    ledgerPrivateGetBinary((uint8_t*)valueKey.c_str(), valueKey.length(), valueBinary,
        storedValueLength, &actualValueLength);
    if (actualValueLength != storedValueLength)
    {
        LOG_ERROR("Unexpected length of retrieved value -- no value returned");
        return;
    }
    if (valueBinary[storedValueLength - 1] != '\0')
    {
        LOG_ERROR("Retrieved value is not null terminated -- no value returned");
        return;
    }
    value.assign((char*)valueBinary);
}

void Contract::Storage::ledgerPublicPutString(const std::string& key, const std::string& value)
{
    // first write length -- string with null char
    unsigned int valueLength = value.length() + 1;
    std::string valueLengthString = std::to_string(valueLength);
    if (valueLengthString.length() == 0 || valueLengthString.length() > MAX_LENGTH_STRING_SIZE)
    {
        LOG_ERROR("value length is 0 or too large. No data will be written to ledger");
        return;
    }
    std::string valueLengthKey = key + "L";
    ledgerPublicPutBinary((const uint8_t*)valueLengthKey.c_str(), valueLengthKey.length(),
        (const uint8_t*)valueLengthString.c_str(), valueLengthString.length() + 1);

    // second, write actual value
    std::string valueKey = key + "K";
    ledgerPublicPutBinary(
        (uint8_t*)valueKey.c_str(), valueKey.length(), (const uint8_t*)value.c_str(), valueLength);
}

void Contract::Storage::ledgerPublicPutBinary(
    const uint8_t* key, const uint32_t keyLength, const uint8_t* value, const uint32_t valueLength)
{
    put_public_state((const char*)key, (uint8_t*)value, valueLength, ctx_);
}

void Contract::Storage::ledgerPrivatePutBinary(
    const uint8_t* key, const uint32_t keyLength, const uint8_t* value, const uint32_t valueLength)
{
    put_state((const char*)key, (uint8_t*)value, valueLength, ctx_);
}

void Contract::Storage::ledgerPrivateGetBinary(const uint8_t* key,
    const uint32_t keyLength,
    uint8_t* value,
    const uint32_t valueLength,
    uint32_t* actualValueLength)
{
    get_state((const char*)key, value, valueLength, actualValueLength, ctx_);
}
