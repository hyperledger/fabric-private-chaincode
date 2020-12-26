/*
 * Copyright IBM Corp. All Rights Reserved.
 * Copyright 2020 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "base64.h"
#include "cc_data.h"
#include "enclave_t.h"
#include "error.h"
#include "fpc/fpc.pb.h"
#include "logging.h"
#include "pb_decode.h"
#include "pb_encode.h"
#include "shim.h"
#include "shim_internals.h"

#include <mbusafecrt.h> /* for memcpy_s etc */

int ecall_cc_invoke(const uint8_t* signed_proposal_proto_bytes,
    uint32_t signed_proposal_proto_bytes_len,
    const uint8_t* cc_request_message_bytes,
    uint32_t cc_request_message_bytes_len,
    uint8_t* signed_cc_response_message_bytes,
    uint32_t signed_cc_response_message_bytes_len_in,
    uint32_t* signed_cc_response_message_bytes_len_out,
    void* u_shim_ctx)
{
    LOG_DEBUG("ecall_cc_invoke");
    LOG_DEBUG("signed proposal length %u", signed_proposal_proto_bytes_len);

    bool b;
    fpc_ChaincodeRequestMessage cc_request_message = {};
    fpc_CleartextChaincodeRequest cleartext_cc_request = {};
    t_shim_ctx_t ctx;
    int ret;
    int invoke_ret;
    // estimate max response len (take into account other fields and b64 encoding)
    uint32_t response_len = signed_cc_response_message_bytes_len_in / 4 * 3 - 1024;
    uint8_t response[signed_cc_response_message_bytes_len_in / 4 * 3];
    uint32_t response_len_out = 0;
    std::string b64_response;
    ByteArray cc_response_message;
    size_t cc_response_message_estimated_size;
    ByteArray return_encryption_key;

    ctx.u_shim_ctx = u_shim_ctx;

    {
        pb_istream_t istream;
        ByteArray clear_request;

        // set stream for ChaincodeRequestMessage
        istream = pb_istream_from_buffer(
            (const unsigned char*)cc_request_message_bytes, cc_request_message_bytes_len);

        b = pb_decode(&istream, fpc_ChaincodeRequestMessage_fields, &cc_request_message);
        COND2LOGERR(!b, PB_GET_ERROR(&istream));
        COND2LOGERR(cc_request_message.encrypted_request->size == 0, "zero size request");

        // decrypt request
        b = g_cc_data->decrypt_cc_message(ByteArray(cc_request_message.encrypted_request->bytes,
                                              cc_request_message.encrypted_request->bytes +
                                                  cc_request_message.encrypted_request->size),
            clear_request);
        COND2ERR(!b);

        // set stream for CleartextChaincodeRequestMessage
        istream = pb_istream_from_buffer(
            (const unsigned char*)clear_request.data(), clear_request.size());
        b = pb_decode(&istream, fpc_CleartextChaincodeRequest_fields, &cleartext_cc_request);
        COND2LOGERR(!b, PB_GET_ERROR(&istream));
        COND2LOGERR(!cleartext_cc_request.has_input, "no input in cleartext request");

        // prepare input arguments
        for (int i = 0; i < cleartext_cc_request.input.args_count; i++)
        {
            ctx.string_args.push_back(
                std::string((const char*)cleartext_cc_request.input.args[i]->bytes,
                    cleartext_cc_request.input.args[i]->size));
        }

        // get and set return encryption key
        return_encryption_key = ByteArray(cleartext_cc_request.return_encryption_key->bytes,
            cleartext_cc_request.return_encryption_key->bytes +
                cleartext_cc_request.return_encryption_key->size);
        COND2LOGERR(return_encryption_key.size() != pdo::crypto::constants::SYM_KEY_LEN,
            "invalid return encryption key length");

        // the dynamic memory in the message is released at the end
    }

    invoke_ret = invoke(response, response_len, &response_len_out, &ctx);
    // invoke_ret is not checked

    // TODO double check or rethink if it is appropriate for a chaincode
    // to return an error and still forward the response
    // in particular: should the enclave sign a response? and the rwset? could the tx be committed
    // though it failed?

    b64_response = base64_encode((const unsigned char*)response, response_len_out);

    {
        // TODO put response in protobuf and encode it

        ByteArray encrypted_response;
        fpc_ChaincodeResponseMessage crm;
        pb_ostream_t ostream;
        std::string enclave_id;

        // create proto struct to encode
        // TODO: create fabric Response object
        // TODO: encrypt fabric Response object
        crm = {};

        {  // encrypt response
            encrypted_response = pdo::crypto::skenc::EncryptMessage(return_encryption_key,
                ByteArray(b64_response.c_str(), b64_response.c_str() + b64_response.length()));
        }

        {  // fill encrypted response
            crm.encrypted_response = (pb_bytes_array_t*)pb_realloc(
                crm.encrypted_response, PB_BYTES_ARRAY_T_ALLOCSIZE(encrypted_response.size()));
            COND2LOGERR(crm.encrypted_response == NULL, "cannot allocate encrypted message");
            crm.encrypted_response->size = encrypted_response.size();
            ret = memcpy_s(crm.encrypted_response->bytes, crm.encrypted_response->size,
                encrypted_response.data(), encrypted_response.size());
            COND2LOGERR(ret != 0, "cannot encode field");
        }

        {  // fill enclave id
            enclave_id = g_cc_data->get_enclave_id();
            crm.enclave_id = (char*)pb_realloc(crm.enclave_id, enclave_id.length() + 1);
            ret = memcpy_s(
                crm.enclave_id, enclave_id.length(), enclave_id.c_str(), enclave_id.length());
            crm.enclave_id[enclave_id.length()] = '\0';
            COND2LOGERR(ret != 0, "cannot encode enclave id");
        }

        {  // fill proposal
            pb_istream_t istream;

            // set stream for ChaincodeRequestMessage
            istream = pb_istream_from_buffer(
                (const unsigned char*)signed_proposal_proto_bytes, signed_proposal_proto_bytes_len);

            b = pb_decode(&istream, protos_SignedProposal_fields, &crm.proposal);
            COND2LOGERR(!b, PB_GET_ERROR(&istream));
            COND2LOGERR(
                crm.proposal.proposal_bytes == NULL || crm.proposal.proposal_bytes->size == 0,
                "zero size proposal");

            crm.has_proposal = true;
        }

        {  // fill rwset
            crm.has_fpc_rw_set = true;
            rwset_to_proto(&ctx, &crm.fpc_rw_set);
        }

        // estimate response message size
        b = pb_get_encoded_size(
            &cc_response_message_estimated_size, fpc_ChaincodeResponseMessage_fields, &crm);
        COND2LOGERR(!b, "cannot estimate response message size");

        // encode proto
        CATCH(b, cc_response_message.resize(cc_response_message_estimated_size));
        COND2LOGERR(!b, "cannot allocate response buffer");
        ostream = pb_ostream_from_buffer(cc_response_message.data(), cc_response_message.size());
        b = pb_encode(&ostream, fpc_ChaincodeResponseMessage_fields, &crm);
        COND2LOGERR(!b, "error encoding proto");
        COND2LOGERR(ostream.bytes_written != cc_response_message_estimated_size,
            "encoding size different than estimated");

        pb_release(fpc_ChaincodeResponseMessage_fields, &crm);
    }

    {
        // create signed response message
        pb_ostream_t ostream;

        // compute signature
        ByteArray signature;
        b = g_cc_data->sign_message(cc_response_message, signature);
        COND2ERR(!b);

        // fill in protobuf structure
        fpc_SignedChaincodeResponseMessage signed_crm = {};

        // fill in response message
        signed_crm.chaincode_response_message = (pb_bytes_array_t*)pb_realloc(
            NULL, PB_BYTES_ARRAY_T_ALLOCSIZE(cc_response_message.size()));
        COND2LOGERR(
            signed_crm.chaincode_response_message == NULL, "cannot allocate response message");
        signed_crm.chaincode_response_message->size = cc_response_message.size();
        ret = memcpy_s(signed_crm.chaincode_response_message->bytes,
            signed_crm.chaincode_response_message->size, cc_response_message.data(),
            cc_response_message.size());
        COND2LOGERR(ret != 0, "cannot encode field");

        // fill in signature
        signed_crm.signature =
            (pb_bytes_array_t*)pb_realloc(NULL, PB_BYTES_ARRAY_T_ALLOCSIZE(signature.size()));
        COND2LOGERR(signed_crm.signature == NULL, "cannot allocate signature");
        signed_crm.signature->size = signature.size();
        ret = memcpy_s(signed_crm.signature->bytes, signed_crm.signature->size, signature.data(),
            signature.size());
        COND2LOGERR(ret != 0, "cannot encode field");

        // encode proto
        ostream = pb_ostream_from_buffer(
            signed_cc_response_message_bytes, signed_cc_response_message_bytes_len_in);
        b = pb_encode(&ostream, fpc_SignedChaincodeResponseMessage_fields, &signed_crm);
        COND2LOGERR(!b, "error encoding proto");

        pb_release(fpc_SignedChaincodeResponseMessage_fields, &signed_crm);

        *signed_cc_response_message_bytes_len_out = ostream.bytes_written;
    }

    // release dynamic allocations (TODO:release in case of error)
    pb_release(fpc_ChaincodeRequestMessage_fields, &cc_request_message);

    // TODO: generate signature (as short-cut for now over proposal _and_ args with consistency of
    // proposal and args verified in "__endorse" rather than enclave)

    return 0;

err:
    *signed_cc_response_message_bytes_len_out = 0;
    return 1;
}
