/*
 * Copyright 2019 Intel Corporation
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

// TODO: This should go at compile time
#define PB_ENABLE_MALLOC

#include "messages.h"
#include <mbusafecrt.h> /* for memcpy_s etc */
#include <pb.h>
#include <pb_decode.h>
#include <pb_encode.h>
#include <string>
#include "_protos/irb.pb.h"

Contract::EASMessage::EASMessage() {}

Contract::EASMessage::EASMessage(const std::string& message) : inputString_(message)
{
    // base64-decode every message at the beginning
    inputMessageBytes_ = Base64EncodedStringToByteArray(inputString_);
}

bool Contract::EASMessage::toStatus(const std::string& message, int rc, std::string& outputMessage)
{
    Status status;
    int ret;
    bool b;

    if (message.empty())
    {
        status.msg = NULL;
    }
    else
    {
        status.msg = (char*)pb_realloc(status.msg, message.length() + 1);
        FAST_FAIL_CHECK(er_, EC_ERROR, status.msg == NULL);
        ret = memcpy_s(status.msg, message.length(), message.c_str(), message.length());
        FAST_FAIL_CHECK(er_, EC_ERROR, ret != 0);
        status.msg[message.length()] = '\0';
    }

    status.return_code = (Status_ReturnCode)rc;

    // encode in protobuf
    pb_ostream_t ostream;
    uint32_t response_len = 1024;
    uint8_t response[response_len];
    ostream = pb_ostream_from_buffer(response, response_len);
    b = pb_encode(&ostream, Status_fields, &status);
    FAST_FAIL_CHECK(er_, EC_ERROR, !b);

    // once encoded on buffer, release dynamic fields
    pb_release(Status_fields, &status);

    // base64-encode
    outputMessage =
        ByteArrayToBase64EncodedString(ByteArray(response, response + ostream.bytes_written));

    return true;
}

bool Contract::EASMessage::fromRegisterDataRequest(
    std::string& uuid, ByteArray& publicKey, ByteArray& decryptionKey)
{
    pb_istream_t istream;
    bool b;
    RegisterDataRequest registerDataRequest;

    istream = pb_istream_from_buffer(
        (const unsigned char*)inputMessageBytes_.data(), inputMessageBytes_.size());
    b = pb_decode(&istream, RegisterDataRequest_fields, &registerDataRequest);
    FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, !b);
    FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, !registerDataRequest.has_participant);
    FAST_FAIL_CHECK(er_, EC_INVALID_INPUT, registerDataRequest.participant.uuid == NULL);

    uuid = std::string(registerDataRequest.participant.uuid);
    publicKey = ByteArray(registerDataRequest.participant.public_key->bytes,
        registerDataRequest.participant.public_key->bytes +
            registerDataRequest.participant.public_key->size);
    decryptionKey = ByteArray(registerDataRequest.decryption_key->bytes,
        registerDataRequest.decryption_key->bytes + registerDataRequest.decryption_key->size);

    return true;
}
