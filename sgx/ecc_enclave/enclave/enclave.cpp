/*
 * Copyright IBM Corp. All Rights Reserved.
 * Copyright 2020 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "base64.h"
#include "cc_data.h"
#include "crypto.h"
#include "enclave_t.h"
#include "error.h"
#include "fabric/common/common.pb.h"
#include "fabric/msp/identities.pb.h"
#include "fabric/peer/chaincode.pb.h"
#include "fabric/peer/proposal.pb.h"
#include "fabric/peer/proposal_response.pb.h"
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
    fpc_KeyTransportMessage key_transport_message = {};
    t_shim_ctx_t ctx;
    int ret;
    int invoke_ret;
    // estimate max response len (take into account other fields and b64 encoding)
    uint32_t response_len = signed_cc_response_message_bytes_len_in / 4 * 3 - 1024;
    uint8_t response[signed_cc_response_message_bytes_len_in / 4 * 3];
    uint32_t response_len_out = 0;
    ByteArray proto_response_bytes;
    ByteArray cc_response_message;
    size_t cc_response_message_estimated_size;
    ByteArray response_encryption_key;

    ctx.u_shim_ctx = u_shim_ctx;
    ctx.signed_proposal = ByteArray(
        signed_proposal_proto_bytes, signed_proposal_proto_bytes + signed_proposal_proto_bytes_len);

    // unmarshall proposal
    {
        pb_istream_t istream;

        protos_SignedProposal signed_proposal = protos_SignedProposal_init_zero;
        istream = pb_istream_from_buffer(
            (const unsigned char*)signed_proposal_proto_bytes, signed_proposal_proto_bytes_len);
        b = pb_decode(&istream, protos_SignedProposal_fields, &signed_proposal);
        COND2LOGERR(!b, PB_GET_ERROR(&istream));

        protos_Proposal proposal = protos_Proposal_init_zero;
        istream =
            pb_istream_from_buffer((const unsigned char*)signed_proposal.proposal_bytes->bytes,
                signed_proposal.proposal_bytes->size);
        b = pb_decode(&istream, protos_Proposal_fields, &proposal);
        COND2LOGERR(!b, PB_GET_ERROR(&istream));

        common_Header header = common_Header_init_zero;
        istream = pb_istream_from_buffer(
            (const unsigned char*)proposal.header->bytes, proposal.header->size);
        b = pb_decode(&istream, common_Header_fields, &header);
        COND2LOGERR(!b, PB_GET_ERROR(&istream));

        // set tx_id and channel id
        common_ChannelHeader channel_header = common_ChannelHeader_init_zero;
        istream = pb_istream_from_buffer(
            (const unsigned char*)header.channel_header->bytes, header.channel_header->size);
        b = pb_decode(&istream, common_ChannelHeader_fields, &channel_header);
        COND2LOGERR(!b, PB_GET_ERROR(&istream));

        ctx.tx_id = std::string(channel_header.tx_id);
        ctx.channel_id = std::string(channel_header.channel_id);
        COND2LOGERR(ctx.channel_id != g_cc_data->get_channel_id(),
            "channel id of the tx proposal does not match as initialized with cc_parameters");

        // TODO implement me
        // transient data
        // protos_ChaincodeProposalPayload cc_proposal_payload =
        //     protos_ChaincodeProposalPayload_init_zero;
        // istream = pb_istream_from_buffer(
        //     (const unsigned char*)proposal.payload->bytes, proposal.payload->size);
        // b = pb_decode(&istream, protos_ChaincodeProposalPayload_fields, &cc_proposal_payload);
        // COND2LOGERR(!b, PB_GET_ERROR(&istream));

        // transform _protos_ChaincodeProposalPayload_TransientMapEntry to std::map
        // ctx.transient_data = ;

        // set creator
        common_SignatureHeader signature_header = common_SignatureHeader_init_zero;
        istream = pb_istream_from_buffer(
            (const unsigned char*)header.signature_header->bytes, header.signature_header->size);
        b = pb_decode(&istream, common_SignatureHeader_fields, &signature_header);
        COND2LOGERR(!b, PB_GET_ERROR(&istream));

        ByteArray signature = ByteArray(signed_proposal.signature->bytes,
            signed_proposal.signature->bytes + signed_proposal.signature->size);
        ByteArray message = ByteArray(signed_proposal.proposal_bytes->bytes,
            signed_proposal.proposal_bytes->bytes + signed_proposal.proposal_bytes->size);

        ctx.creator = ByteArray(signature_header.creator->bytes,
            signature_header.creator->bytes + signature_header.creator->size);

        msp_SerializedIdentity identity = msp_SerializedIdentity_init_zero;
        istream = pb_istream_from_buffer((const unsigned char*)(signature_header.creator->bytes),
            signature_header.creator->size);
        b = pb_decode(&istream, msp_SerializedIdentity_fields, &identity);
        COND2LOGERR(!b, PB_GET_ERROR(&istream));

        ByteArray encoded_signer_cert =
            ByteArray(identity.id_bytes->bytes, identity.id_bytes->bytes + identity.id_bytes->size);
        b = validate_message_signature(signature, message, encoded_signer_cert);
        COND2LOGERR(!b, "signature validation failed");

        ctx.creator_msp_id = std::string(identity.mspid);
        ctx.creator_name = extract_subject_from_cert(encoded_signer_cert);

        // TODO unwrap ChaincodeRequestMessage from signed proposal
        // once this is in place we can remove ChaincodeRequestMessage from ecall

        // TODO implement me
        // ByteArray binding_data;
        // binding_data.insert(binding_data.end(), signature_header.nonce->bytes,
        // signature_header.nonce->size); binding_data.insert(binding_data.end(),
        // signature_header.creator->bytes, signature_header.creator->size);
        // TODO add channel_header->epoch (as ByteArray) to binding (enforce little endian)
        // b = compute_message_hash(binding_data, ctx->binding);
        // COND2LOGERR(!b, "cannot compute binding");

        pb_release(msp_SerializedIdentity_fields, &identity);
        pb_release(common_SignatureHeader_fields, &signature_header);
        pb_release(common_ChannelHeader_fields, &channel_header);
        pb_release(common_Header_fields, &header);
        pb_release(protos_Proposal_fields, &proposal);
        pb_release(protos_SignedProposal_fields, &signed_proposal);
    }

    {
        pb_istream_t istream;
        ByteArray clear_request;
        ByteArray key_transport;

        // set stream for ChaincodeRequestMessage
        istream = pb_istream_from_buffer(
            (const unsigned char*)cc_request_message_bytes, cc_request_message_bytes_len);

        b = pb_decode(&istream, fpc_ChaincodeRequestMessage_fields, &cc_request_message);
        COND2LOGERR(!b, PB_GET_ERROR(&istream));
        COND2LOGERR(cc_request_message.encrypted_request->size == 0, "zero size request");
        COND2LOGERR(cc_request_message.encrypted_key_transport_message->size == 0,
            "zero size key transport message");

        {  // decrypt key transport
            ByteArray encrypted_key_transport_message =
                ByteArray(cc_request_message.encrypted_key_transport_message->bytes,
                    cc_request_message.encrypted_key_transport_message->bytes +
                        cc_request_message.encrypted_key_transport_message->size);
            b = g_cc_data->decrypt_key_transport_message(
                encrypted_key_transport_message, key_transport);
            COND2LOGERR(!b, "cannot decrypt key transport message");
        }

        // set stream for KeyTransportMessage
        istream = pb_istream_from_buffer(
            (const unsigned char*)key_transport.data(), key_transport.size());
        b = pb_decode(&istream, fpc_KeyTransportMessage_fields, &key_transport_message);
        COND2LOGERR(!b, PB_GET_ERROR(&istream));

        // get and set response encryption key
        response_encryption_key = ByteArray(key_transport_message.response_encryption_key->bytes,
            key_transport_message.response_encryption_key->bytes +
                key_transport_message.response_encryption_key->size);

        {  // decrypt request
            ByteArray request_encryption_key =
                ByteArray(key_transport_message.request_encryption_key->bytes,
                    key_transport_message.request_encryption_key->bytes +
                        key_transport_message.request_encryption_key->size);
            ByteArray encrypted_request = ByteArray(cc_request_message.encrypted_request->bytes,
                cc_request_message.encrypted_request->bytes +
                    cc_request_message.encrypted_request->size);
            b = decrypt_message(request_encryption_key, encrypted_request, clear_request);
            COND2LOGERR(!b, "message decryption failed");
        }

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

        // the dynamic memory in the message is released at the end
    }

    try
    {
        invoke_ret = invoke(response, response_len, &response_len_out, &ctx);
        // invoke_ret is not checked

        // TODO double check or rethink if it is appropriate for a chaincode
        // to return an error and still forward the response
        // in particular: should the enclave sign a response? and the rwset? could the tx be
        // committed though it failed?
    }
    catch (std::exception& e)
    {
        COND2LOGERR(true, e.what());
    }

    // wrap invocation response in a peer.Response message
    {
        protos_Response protoResponse = {};

        if (invoke_ret == 0)
        {
            protoResponse.status = 200;

            // fill in response message
            protoResponse.payload =
                (pb_bytes_array_t*)pb_realloc(NULL, PB_BYTES_ARRAY_T_ALLOCSIZE(response_len_out));
            COND2LOGERR(protoResponse.payload == NULL, "cannot allocate response payload");
            protoResponse.payload->size = response_len_out;

            ret = memcpy_s(protoResponse.payload->bytes, protoResponse.payload->size, response,
                response_len_out);
            COND2LOGERR(ret != 0, "cannot encode field");
        }
        else
        {
            // set error
            protoResponse.status = 500;
            // TODO set response.message with proper error message as reported by user
        }

        // estimate response message size and resize buffer
        size_t response_estimated_size;
        b = pb_get_encoded_size(&response_estimated_size, protos_Response_fields, &protoResponse);
        COND2LOGERR(!b, "cannot estimate response message size");

        CATCH(b, proto_response_bytes.resize(response_estimated_size));

        // encode proto
        pb_ostream_t ostream =
            pb_ostream_from_buffer(proto_response_bytes.data(), proto_response_bytes.size());
        b = pb_encode(&ostream, protos_Response_fields, &protoResponse);
        COND2LOGERR(!b, "error encoding response proto");
        COND2LOGERR(ostream.bytes_written != response_estimated_size,
            "encoding size different than estimated");

        pb_release(protos_Response_fields, &protoResponse);
    }

    {
        ByteArray encrypted_response;
        fpc_ChaincodeResponseMessage crm;
        pb_ostream_t ostream;
        std::string enclave_id;

        // create proto struct to encode
        crm = {};

        {  // encrypt response
            b = encrypt_message(response_encryption_key, proto_response_bytes, encrypted_response);
            COND2LOGERR(!b, "cannot encrypt response message");
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

        {  // fill chaincode request message hash
            // hash request
            ByteArray ba_cc_request_message(
                cc_request_message_bytes, cc_request_message_bytes + cc_request_message_bytes_len);
            ByteArray ba_cc_request_message_hash;
            b = compute_message_hash(ba_cc_request_message, ba_cc_request_message_hash);
            COND2LOGERR(!b, "cannot compute request message hash");

            // encode field
            LOG_DEBUG("adding request hash: %s",
                (ByteArrayToHexEncodedString(ba_cc_request_message_hash)).c_str());
            crm.chaincode_request_message_hash = (pb_bytes_array_t*)pb_realloc(
                NULL, PB_BYTES_ARRAY_T_ALLOCSIZE(ba_cc_request_message_hash.size()));
            COND2LOGERR(crm.chaincode_request_message_hash == NULL, "cannot allocate request hash");
            crm.chaincode_request_message_hash->size = ba_cc_request_message_hash.size();
            ret = memcpy_s(crm.chaincode_request_message_hash->bytes,
                crm.chaincode_request_message_hash->size, ba_cc_request_message_hash.data(),
                ba_cc_request_message_hash.size());
            COND2LOGERR(ret != 0, "cannot encode request hash");
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
    pb_release(fpc_CleartextChaincodeRequest_fields, &cleartext_cc_request);
    pb_release(fpc_KeyTransportMessage_fields, &key_transport_message);

    // TODO: generate signature (as short-cut for now over proposal _and_ args with consistency of
    // proposal and args verified in "__endorse" rather than enclave)

    return 0;

err:
    *signed_cc_response_message_bytes_len_out = 0;
    return 1;
}
