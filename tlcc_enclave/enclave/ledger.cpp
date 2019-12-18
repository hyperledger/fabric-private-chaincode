/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "ledger.h"

// proto
#include <pb_decode.h>
#include <pb_encode.h>
#include "common/common.pb.h"
#include "common/configtx.pb.h"
#include "ledger/rwset/kvrwset/kv_rwset.pb.h"
#include "ledger/rwset/rwset.pb.h"
#include "msp/identities.pb.h"
#include "msp/msp_config.pb.h"
#include "peer/proposal.pb.h"
#include "peer/proposal_response.pb.h"
#include "peer/transaction.pb.h"

// openssl
#include <openssl/x509.h>
#include "openssl/sha.h"

#include "asn1_utils.h"
#include "base64.h"
#include "crypto.h"
#include "ecc_json.h"
#include "ias.h"
#include "logging.h"
#include "utils.h"

// root cert store for orderer org
static X509_STORE* root_certs_orderer = NULL;
// root cert store for application orgs
static X509_STORE* root_certs_apps = NULL;

static kvs_t state;      // blockchain state
static spinlock_t lock;  // state lock

static uint32_t sequence_number = -1;  // sequence number counter

int init_ledger()
{
    LOG_DEBUG("Ledger: ########## init ledger  ##########");

    root_certs_orderer = X509_STORE_new();
    if (root_certs_orderer == NULL)
    {
        LOG_ERROR("Ledger: Can not create root cert store for orderer");
        return LEDGER_ERROR_CRYPTO;
    }

    root_certs_apps = X509_STORE_new();
    if (root_certs_apps == NULL)
    {
        LOG_ERROR("Ledger: Can not create root cert store for apps");
        return LEDGER_ERROR_CRYPTO;
    }

    return LEDGER_SUCCESS;
}

int parse_block(uint8_t* block_data, uint32_t block_data_len)
{
    LOG_DEBUG("Ledger: ########## Parse block data; len =  %d ##########", block_data_len);

    common_Block block = common_Block_init_zero;
    decode_pb(block, common_Block_fields, block_data, block_data_len);

    const uint32_t block_sequence_number = block.header.number;
    // output block header
    LOG_DEBUG("Ledger: Block number: %d", block_sequence_number);

    // create DER representation of block header
    uint8_t* header_DER;
    uint32_t header_DER_len = block_header2DER(&block.header, &header_DER);

    if (block_sequence_number != sequence_number + 1)
    {
        LOG_ERROR("Ledger: Last known seqNo = %d but receiving block no %d", sequence_number,
            block.header.number);
        pb_release(common_Block_fields, &block);
        free(header_DER);
        return LEDGER_INVALID_BLOCK_NUMBER;
    }

    // verify block - get block signature from metadata[0]
    pb_bytes_array_t* metadata_bytes =
        block.metadata.metadata[common_BlockMetadataIndex_SIGNATURES];

    // decode metadata
    common_Metadata metadata = common_Metadata_init_zero;
    decode_pb(metadata, common_Metadata_fields, metadata_bytes->bytes, metadata_bytes->size);
    {
        // metadata value ... seems to be empty all the time ???
        if (metadata.value == NULL)
        {
            metadata.value = (pb_bytes_array_t*)malloc(PB_BYTES_ARRAY_T_ALLOCSIZE(0));
            metadata.value->size = 0;
        }

        // go through all block signatues
        for (int i = 0; i < metadata.signatures_count; i++)
        {
            LOG_DEBUG("Ledger: Verify block signature[%d/%d]", i + 1, metadata.signatures_count);
            common_MetadataSignature* metadata_signature = &metadata.signatures[i];

            // unwrap signature header
            common_SignatureHeader header = common_SignatureHeader_init_zero;
            decode_pb(header, common_SignatureHeader_fields,
                metadata_signature->signature_header->bytes,
                metadata_signature->signature_header->size);

            // unwrap identity
            msp_SerializedIdentity identity = msp_SerializedIdentity_init_zero;
            decode_pb(identity, msp_SerializedIdentity_fields, header.creator->bytes,
                header.creator->size);
            LOG_DEBUG("Ledger: \t\\-> MSPID: %s", identity.mspid);

            // compute signature hash
            unsigned char sig_hash[HASH_SIZE];
            SHA256_CTX sha256;
            SHA256_Init(&sha256);
            SHA256_Update(&sha256, (const uint8_t*)&metadata.value->bytes, metadata.value->size);
            SHA256_Update(&sha256, (const uint8_t*)&metadata_signature->signature_header->bytes,
                metadata_signature->signature_header->size);
            SHA256_Update(&sha256, (const uint8_t*)header_DER, header_DER_len);
            SHA256_Final(sig_hash, &sha256);

            const unsigned char* ptr = metadata_signature->signature->bytes;
            if (verify_signature(&ptr, metadata_signature->signature->size, sig_hash, HASH_SIZE,
                    identity.id_bytes->bytes, identity.id_bytes->size, root_certs_orderer) != 1)
            {
                LOG_ERROR("Ledger: Block signature valudation failed");
            }
            else
            {
                LOG_DEBUG("Ledger: \t\\-> Valid block cert");
            }

            pb_release(msp_SerializedIdentity_fields, &identity);
            pb_release(common_SignatureHeader_fields, &header);
        }  // block signatures
    }      // metadata
    pb_release(common_Metadata_fields, &metadata);
    free(header_DER);

    // prepare tx filter
    pb_bytes_array_t* tx_filter_pb =
        block.metadata.metadata[common_BlockMetadataIndex_TRANSACTIONS_FILTER];
    uint8_t tx_filter[block.data.data_count] = {0};
    for (int i = 0; i < tx_filter_pb->size; i++)
    {
        tx_filter[i] = tx_filter_pb->bytes[i];
    }

    // prepare updates/write set for this block
    kvs_t updates;

    // go through all envelopes/transactions (block.data)
    for (uint64_t i = 0; i < block.data.data_count; i++)
    {
        LOG_DEBUG("Ledger: Process Envelope[%d/%d]", i + 1, block.data.data_count);

        // check if tx was invalided in pre consensus; in that case skip tx
        if (tx_filter[i] == 1)
        {
            LOG_DEBUG("Ledger: Transaction [%d] marked as invalid. Continue", i);
            continue;
        }

        // set version of this tx
        version_t version = {block.header.number, i};

        // unwrapp common envelope
        common_Envelope envelope = common_Envelope_init_zero;
        decode_pb(
            envelope, common_Envelope_fields, block.data.data[i]->bytes, block.data.data[i]->size);

        // unwrapp envelope payload
        common_Payload payload = common_Payload_init_zero;
        decode_pb(payload, common_Payload_fields, envelope.payload->bytes, envelope.payload->size);

        // verify envelope signature
        LOG_DEBUG("Ledger: \t\\-> Verify envelope signature");

        // unwrap envelope signature header
        common_SignatureHeader header = common_SignatureHeader_init_zero;
        decode_pb(header, common_SignatureHeader_fields, payload.header.signature_header->bytes,
            payload.header.signature_header->size);

        // unwrap identity
        msp_SerializedIdentity identity = msp_SerializedIdentity_init_zero;
        decode_pb(
            identity, msp_SerializedIdentity_fields, header.creator->bytes, header.creator->size);
        LOG_DEBUG("Ledger: \t\t\\-> MSPID: %s", identity.mspid);

        // compute signature hash
        unsigned char sig_hash[HASH_SIZE];
        SHA256_CTX sha256;
        SHA256_Init(&sha256);
        SHA256_Update(&sha256, (const uint8_t*)&envelope.payload->bytes, envelope.payload->size);
        SHA256_Final(sig_hash, &sha256);

        // verify validate cert. note that if this is a gensis block, we skip
        // the validation
        if (block.header.number > 0)
        {
            const unsigned char* ptr = envelope.signature->bytes;
            if (verify_signature(&ptr, envelope.signature->size, sig_hash, HASH_SIZE,
                    identity.id_bytes->bytes, identity.id_bytes->size, root_certs_apps) != 1)
            {
                LOG_ERROR("Ledger: Envelope signature validation failed");
            }
            else
            {
                LOG_DEBUG("Ledger: \t\t\\-> Valid envelope signature");
            }
        }
        else
        {
            LOG_DEBUG("Ledger: Skip signature validation; genesis block");
        }

        // parse channel header
        common_ChannelHeader chdr = common_ChannelHeader_init_zero;
        decode_pb(chdr, common_ChannelHeader_fields, payload.header.channel_header->bytes,
            payload.header.channel_header->size);

        // the following checks are not needed for genesis block
        if (block.header.number > 0)
        {
            // create tx_id from nonce and creator
            unsigned char tx_id_bytes[HASH_SIZE];
            SHA256_CTX sha256;
            SHA256_Init(&sha256);
            SHA256_Update(&sha256, header.nonce->bytes, header.nonce->size);
            SHA256_Update(&sha256, header.creator->bytes, header.creator->size);
            SHA256_Final(tx_id_bytes, &sha256);

            // transform to string and verify Tx (note that genesis block has no
            // tx ID)
            char* tx_id = bytes_to_hexstring(tx_id_bytes, HASH_SIZE);
            if (strlen(tx_id) == strlen(chdr.tx_id) &&
                memcmp(tx_id, chdr.tx_id, strlen(tx_id)) != 0)
            {
                LOG_ERROR("Ledger: Incorrect TxID");
                // TODO abord
            }
            LOG_DEBUG("Ledger: \t\t\\-> Valid tx id: %s", chdr.tx_id);
            free(tx_id);
        }

        spin_lock(&lock);
        switch (chdr.type)
        {
            case common_HeaderType_CONFIG:
                if (block.header.number == 0)
                {
                    // note that we currently do not support config updates
                    // (genesis only)
                    parse_config(payload.data->bytes, payload.data->size);
                }
                break;
            case common_HeaderType_ENDORSER_TRANSACTION:
                parse_endorser_transaction(
                    payload.data->bytes, payload.data->size, &updates, &version);
                break;
            default:
                LOG_ERROR("Ledger: Invalid envelope type");
                break;
        }
        spin_unlock(&lock);

        pb_release(common_ChannelHeader_fields, &chdr);
        pb_release(msp_SerializedIdentity_fields, &identity);
        pb_release(common_SignatureHeader_fields, &header);
        pb_release(common_Payload_fields, &payload);
        pb_release(common_Envelope_fields, &envelope);
    }  // envelops

    pb_release(common_Block_fields, &block);

    // commit updates/writeset
    commit_state_updates(&updates, block_sequence_number);

    return LEDGER_SUCCESS;
}

int commit_state_updates(kvs_t* updates, const uint32_t block_sequence_number)
{
    LOG_DEBUG("Ledger: ### Apply updates ###");
    spin_lock(&lock);
    assert(block_sequence_number == sequence_number + 1);
    sequence_number = block_sequence_number;
    for (auto pair : *updates)
    {
        LOG_DEBUG("Ledger: \\-> Key: \"%s\" => version: (%d,%d)", pair.first.c_str(),
            pair.second.second.block_num, pair.second.second.tx_num);
        state[pair.first] = pair.second;
    }
    spin_unlock(&lock);
    return LEDGER_SUCCESS;
}

int parse_config(uint8_t* config_data, uint32_t config_data_len)
{
    LOG_DEBUG("Ledger: ### Parse Config ###");

    common_ConfigEnvelope config_envelope = common_ConfigEnvelope_init_zero;
    decode_pb(config_envelope, common_ConfigEnvelope_fields, config_data, config_data_len);

    LOG_DEBUG("Ledger: ConfigEnv.config.ChannelGroup.Groups:");
    for (int i = 0; i < config_envelope.config.channel_group.groups_count; i++)
    {
        char* group = config_envelope.config.channel_group.groups[i].key;
        common_ConfigGroup* groups = &config_envelope.config.channel_group.groups[i].value;
        LOG_DEBUG("Ledger: \tGroup [%d/%d]: %s", i + 1,
            config_envelope.config.channel_group.groups_count, group);

        // select correct root cert store
        X509_STORE* root_certs = NULL;
        if (strcmp(group, "Orderer") == 0)
        {
            root_certs = root_certs_orderer;
        }
        else if (strcmp(group, "Application") == 0)
        {
            root_certs = root_certs_apps;
        }
        else
        {
            LOG_ERROR("Ledger: Unknown channel group: %s", group);
        }

        // go through
        for (int j = 0; j < groups->groups_count; j++)
        {
            common_ConfigGroup* orgs = &groups->groups[j].value;
            LOG_DEBUG(
                "Ledger: \t\tOrg [%d/%d]: %s", j + 1, groups->groups_count, groups->groups[j].key);

            for (int h = 0; h < orgs->values_count; h++)
            {
                if (strcmp(orgs->values[h].key, "MSP") != 0)
                {
                    // skip everything except of MSP config updates
                    LOG_DEBUG("Ledger: \t\t\t>> Skip %s config update", orgs->values[h].key);
                    continue;
                }
                common_ConfigValue* msp = &orgs->values[h].value;
                msp_MSPConfig msp_config = msp_MSPConfig_init_zero;
                decode_pb(msp_config, msp_MSPConfig_fields, msp->value->bytes, msp->value->size);

                // ignore type here and just assume fabric msp config
                msp_FabricMSPConfig fabric_msp_config = msp_FabricMSPConfig_init_zero;
                decode_pb(fabric_msp_config, msp_FabricMSPConfig_fields, msp_config.config->bytes,
                    msp_config.config->size);
                LOG_DEBUG("Ledger: \t\t\tMSP Config: %s", fabric_msp_config.name);

                LOG_DEBUG("Ledger: \t\t\t\\-> Root certs: %d", fabric_msp_config.root_certs_count);
                for (int r = 0; r < fabric_msp_config.root_certs_count; r++)
                {
                    if (store_root_cert(fabric_msp_config.root_certs[r]->bytes,
                            fabric_msp_config.root_certs[r]->size, root_certs) != 1)
                    {
                        LOG_ERROR("Ledger: Can not store root cert");
                    }
                }

                LOG_DEBUG("Ledger: \t\t\t\\-> Admin certs: %d", fabric_msp_config.admins_count);
                for (int r = 0; r < fabric_msp_config.admins_count; r++)
                {
                    if (validate_cert(fabric_msp_config.admins[r]->bytes,
                            fabric_msp_config.admins[r]->size, root_certs) != 1)
                    {
                        LOG_ERROR("Ledger: Invalid admin cert");
                    }
                }

                pb_release(msp_FabricMSPConfig_fields, &fabric_msp_config);
                pb_release(msp_MSPConfig_fields, &msp_config);
            }
        }
    }
    pb_release(common_ConfigEnvelope_fields, &config_envelope);

    return LEDGER_SUCCESS;
}

int parse_endorser_transaction(
    uint8_t* tx_data, uint32_t tx_data_len, kvs_t* updates, version_t* tx_version)
{
    LOG_DEBUG("Ledger: ### Parse Endorser Transaction");

    protos_Transaction transaction = protos_Transaction_init_zero;
    decode_pb(transaction, protos_Transaction_fields, tx_data, tx_data_len);

    // go through all actions in transaction
    for (int i = 0; i < transaction.actions_count; i++)
    {
        /* LOG_DEBUG("### Action[%d] ###", i); */

        // get action payload
        protos_ChaincodeActionPayload action_payload = protos_ChaincodeActionPayload_init_zero;
        decode_pb(action_payload, protos_ChaincodeActionPayload_fields,
            transaction.actions[i].payload->bytes, transaction.actions[i].payload->size);

        // proposal (note that this is not needed here in the ledger)
        protos_ChaincodeProposalPayload proposal_payload =
            protos_ChaincodeProposalPayload_init_zero;
        decode_pb(proposal_payload, protos_ChaincodeProposalPayload_fields,
            action_payload.chaincode_proposal_payload->bytes,
            action_payload.chaincode_proposal_payload->size);
        protos_ChaincodeInvocationSpec cis = protos_ChaincodeInvocationSpec_init_zero;
        decode_pb(cis, protos_ChaincodeInvocationSpec_fields, proposal_payload.input->bytes,
            proposal_payload.input->size);

        /* LOG_DEBUG("Ledger: \tInput args:"); */
        /* for (int i= 0; i< cis.chaincode_spec.input.args_count; i++) { */
        /*     std::string arg((const
         * char*)cis.chaincode_spec.input.args[i]->bytes,
         * cis.chaincode_spec.input.args[i]->size); */
        /*     LOG_DEBUG("Ledger: \t \\-> [%d] %s", i, arg.c_str()); */
        /* } */

        // get function
        std::string function = "";
        if (cis.chaincode_spec.input.args_count > 0)
        {
            function.append((const char*)cis.chaincode_spec.input.args[0]->bytes,
                cis.chaincode_spec.input.args[0]->size);
        }
        LOG_DEBUG("Ledger: function='%s'", function.c_str());

        // parse proposal repsonse payload
        protos_ProposalResponsePayload p_response_payload =
            protos_ProposalResponsePayload_init_zero;
        decode_pb(p_response_payload, protos_ProposalResponsePayload_fields,
            action_payload.action.proposal_response_payload->bytes,
            action_payload.action.proposal_response_payload->size);

        // TODO check that this hash matches the proposal (not relevant for
        // prototype)

        // parse chaincode action
        protos_ChaincodeAction cc_action = protos_ChaincodeAction_init_zero;
        decode_pb(cc_action, protos_ChaincodeAction_fields, p_response_payload.extension->bytes,
            p_response_payload.extension->size);

        LOG_DEBUG(
            "Ledger: \t \\-> ChaincodeAction.ChaincodeId.Name: %s", cc_action.chaincode_id.name);
        LOG_DEBUG("Ledger: \t \\-> ChaincodeAction.Response:");

        // if there are any results let's parse them
        if (cc_action.results != NULL)
        {
            LOG_DEBUG("Ledger: \t\t \\-> ChaincodeAction.Result:");

            rwset_TxReadWriteSet tx_rw_set = rwset_TxReadWriteSet_init_zero;
            decode_pb(tx_rw_set, rwset_TxReadWriteSet_fields, cc_action.results->bytes,
                cc_action.results->size);

            // validity marker for this tx
            bool valid_tx = true;

            // we need that during writeset validation
            std::string ecc_namespace;
            std::string cc_name(cc_action.chaincode_id.name);

            // currently private chaincodes need to have ecc prefix
            if (cc_name.compare(0, 3, "ecc") == 0)
            {
                ecc_namespace = cc_name;
            }

            // read set
            read_set_t ecc_read_set;
            LOG_DEBUG("Ledger: \t\t \\-> Reads:");
            for (int i = 0; i < tx_rw_set.ns_rwset_count; i++)
            {
                kvrwset_KVRWSet kvrwset = kvrwset_KVRWSet_init_zero;
                decode_pb(kvrwset, kvrwset_KVRWSet_fields, tx_rw_set.ns_rwset[i].rwset->bytes,
                    tx_rw_set.ns_rwset[i].rwset->size);
                std::string ns(tx_rw_set.ns_rwset[i].ns);

                // check range queries
                for (int i = 0; i < kvrwset.range_queries_info_count; i++)
                {
                    kvrwset_RangeQueryInfo query_info = kvrwset.range_queries_info[i];
                    kvrwset_QueryReads raw_reads = query_info.reads_info.raw_reads;
                    for (int j = 0; j < raw_reads.kv_reads_count; j++)
                    {
                        std::string key = ns + ".";
                        // we replace 0x00 with "."
                        char* idx = raw_reads.kv_reads[j].key + 1;
                        while (*idx)
                        {
                            key.append(idx);
                            key.append(".");
                            idx += strlen(idx) + 1;
                        }
                        kvrwset_Version _v = raw_reads.kv_reads[j].version;
                        version_t v = {_v.block_num, _v.tx_num};

                        // add to ecc read_set
                        if (ns.compare(ecc_namespace) == 0)
                        {
                            std::string kkey(key, ns.length() + 1, std::string::npos);
                            ecc_read_set.insert(kkey);
                        }

                        // check if there is already something in the blockchain
                        // state
                        // next check in update/writeset for this block
                        if (has_version_conflict(key, &state, &v) == 1 ||
                            has_version_conflict(key, updates, &v) == 1)
                        {
                            valid_tx = false;
                            continue;
                        }
                        LOG_DEBUG(
                            "Ledger: \t\t\t \\-> key = %s version = blockNum: "
                            "%d; txNum: %d",
                            key.c_str(), v.block_num, v.tx_num);
                    }

                    if (!valid_tx)
                    {
                        continue;
                    }
                }

                // normal reads
                for (int i = 0; i < kvrwset.reads_count; i++)
                {
                    std::string key = ns + "." + kvrwset.reads[i].key;
                    kvrwset_Version _v = kvrwset.reads[i].version;
                    version_t v = {_v.block_num, _v.tx_num};

                    // add to read_set
                    if (ns.compare(ecc_namespace) == 0)
                    {
                        std::string kkey(kvrwset.reads[i].key);
                        ecc_read_set.insert(kkey);
                    }

                    // check if there is already something in the blockchain
                    // state
                    // next check in update/writeset for this block
                    if (has_version_conflict(key, &state, &v) == 1 ||
                        has_version_conflict(key, updates, &v) == 1)
                    {
                        valid_tx = false;
                        continue;
                    }
                    LOG_DEBUG("Ledger: \t\t\t \\-> key = %s version = blockNum: %d; txNum: %d",
                        key.c_str(), v.block_num, v.tx_num);
                }
                pb_release(kvrwset_KVRWSet_fields, &kvrwset);
                if (!valid_tx)
                {
                    continue;
                }
            }  // reads

            // abort if tx is invalid
            if (!valid_tx)
            {
                LOG_DEBUG("Ledger: >> Invalid tx through readset");
                pb_release(rwset_TxReadWriteSet_fields, &tx_rw_set);
                pb_release(protos_ChaincodeAction_fields, &cc_action);
                pb_release(protos_ProposalResponsePayload_fields, &p_response_payload);
                pb_release(protos_ChaincodeActionPayload_fields, &action_payload);
                continue;
            }

            // write set
            write_set_t ecc_write_set;
            LOG_DEBUG("Ledger: \t\t \\-> Writes:");
            for (int i = 0; i < tx_rw_set.ns_rwset_count; i++)
            {
                kvrwset_KVRWSet kvrwset = kvrwset_KVRWSet_init_zero;
                decode_pb(kvrwset, kvrwset_KVRWSet_fields, tx_rw_set.ns_rwset[i].rwset->bytes,
                    tx_rw_set.ns_rwset[i].rwset->size);

                std::string ns(tx_rw_set.ns_rwset[i].ns);
                // this is a hack; normally we should ask lscc to get
                // corresponding vscc for chaincode
                // however, for the prototype this is OK!!!
                if (ns.compare("lscc") == 0)
                {
                    for (int i = 0; i < kvrwset.writes_count; i++)
                    {
                        std::string key = ns + "." + kvrwset.writes[i].key;
                        LOG_DEBUG("Ledger: \t\t\t \\-> key = %s", key.c_str());
                        // note prototype does not support deletes
                        std::string val((const char*)kvrwset.writes[i].value->bytes,
                            kvrwset.writes[i].value->size);
                        version_t version = {tx_version->block_num, tx_version->tx_num};
                        updates->insert(kvs_item_t(key, kvs_value_t(val, version)));
                    }
                }
                else if (ns.compare("ercc") == 0)
                {
                    // ercc does only a single write
                    if (kvrwset.writes_count != 1)
                    {
                        LOG_ERROR("Ledger: ercc expects only a single write ");
                        valid_tx = false;
                    }
                    std::string mrenclave_key = ecc_namespace + ".MRENCLAVE";

                    // get mrenclave from updates
                    kvs_iterator_t it = updates->find(mrenclave_key);
                    if (it == updates->end())
                    {
                        // if not in updates get it from kvs
                        it = state.find(mrenclave_key);
                        if (it == state.end())
                        {
                            // if there is no mrenclave at all we are in trouble
                            LOG_ERROR("Ledger: >>> NO MRENCLAVE found for: %s", mrenclave_key);
                            valid_tx = false;
                        }
                    }

                    mrenclave_t mrenclave;
                    std::string _mrenclave = base64_decode(it->second.first);
                    memcpy(&mrenclave, _mrenclave.c_str(), _mrenclave.size());

                    // TODO enable verification here!!!!
                    /* if
                     * (verify_attestation_report(kvrwset.writes[0].value->bytes,
                     * kvrwset.writes[0].value->size, &mrenclave) != 0) { */
                    /*     LOG_ERROR("Ledger: >>> Attestation report invalid"); */
                    /*     valid_tx = false; */
                    /* } */

                    std::string key = ns + "." + kvrwset.writes[0].key;
                    LOG_DEBUG("Ledger: \t\t\t \\-> key = %s", key.c_str());
                    std::string val(
                        (const char*)kvrwset.writes[0].value->bytes, kvrwset.writes[0].value->size);
                    version_t version = {tx_version->block_num, tx_version->tx_num};
                    updates->insert(kvs_item_t(key, kvs_value_t(val, version)));
                }
                else if (ns.compare(0, 3, "ecc") == 0)
                {
                    for (int i = 0; i < kvrwset.writes_count; i++)
                    {
                        std::string kkey;
                        // check for composite keys
                        // note that they start with 0x00 and also use 0x00 as
                        // separator
                        std::string key = ns + ".";
                        if (kvrwset.writes[i].key[0] == 0x00)
                        {
                            // we replace 0x00 with "."
                            char* idx = kvrwset.writes[i].key + 1;
                            while (*idx)
                            {
                                key.append(idx);
                                key.append(".");
                                idx += strlen(idx) + 1;
                            }
                            kkey.append(std::string(key, ns.length(), std::string::npos));
                        }
                        else
                        {
                            // if this is a normal key
                            key.append(kvrwset.writes[i].key);
                            kkey.append(kvrwset.writes[i].key);
                        }
                        LOG_DEBUG("Ledger: \t\t\t \\-> key = %s", key.c_str());
                        // note prototype does not support deletes
                        std::string val((const char*)kvrwset.writes[i].value->bytes,
                            kvrwset.writes[i].value->size);
                        version_t version = {tx_version->block_num, tx_version->tx_num};
                        updates->insert(kvs_item_t(key, kvs_value_t(val, version)));

                        // add to ecc write set
                        ecc_write_set.insert({kkey, val});
                    }
                }
                pb_release(kvrwset_KVRWSet_fields, &kvrwset);

                if (!valid_tx)
                {
                    LOG_DEBUG("Ledger: >> Invalid tx through readset");
                    continue;
                }
            }  // writes

            // only do if ecc transaction
            if (ecc_namespace != "")
            {
                // skip setup transaction
                if (function.compare("__setup") != 0)
                {
                    LOG_DEBUG("Ledger: Validate ECC tx");
                    uint8_t response_data[cc_action.response.payload
                                              ->size];  // no compression involved, so this should
                                                        // be a safe upperbound ...
                    uint32_t response_len = sizeof(response_data);
                    uint8_t signature[96];  // TODO: replace me with something more robust than a
                                            // simply (unexplained) constant ..
                    uint32_t signature_len = sizeof(signature);
                    uint8_t pk[96];  // TODO: replace me with something more robust than a simply
                                     // (unexplained) constant ..
                    uint32_t pk_len = sizeof(pk);

                    unmarshal_ecc_response((const uint8_t*)cc_action.response.payload->bytes,
                        cc_action.response.payload->size, (uint8_t*)&response_data, &response_len,
                        signature, &signature_len, pk, &pk_len);

                    const char* txType;
                    std::vector<std::string> argss;
                    if (function.compare("__init") == 0)
                    {
                        txType = "init";
                        for (int i = 1; i < cis.chaincode_spec.input.args_count; i++)
                        {  // Start from 1 as we drop __init
                            std::string arg = "";
                            arg.append((const char*)cis.chaincode_spec.input.args[i]->bytes,
                                cis.chaincode_spec.input.args[i]->size);
                            argss.push_back(arg);
                        }
                    }
                    else
                    {
                        txType = "invoke";
                        for (int i = 0; i < cis.chaincode_spec.input.args_count; i++)
                        {
                            std::string arg = "";
                            arg.append((const char*)cis.chaincode_spec.input.args[i]->bytes,
                                cis.chaincode_spec.input.args[i]->size);
                            argss.push_back(arg);
                        }
                    }
                    std::string encoded_args = "";
                    marshal_ecc_args(argss, encoded_args);

                    // Note: below signature was created in
                    // ecc_enclave/enclave/enclave.cpp::gen_response
                    // see also replicated verification in ecc/crypto/ecdsa.go::Verify (for
                    // VSCC)

                    // format H(txType in {"init", "invoke"} || encoded_args || result || read
                    // set || write set)
                    uint8_t hash[HASH_SIZE];
                    SHA256_CTX sha256;
                    SHA256_Init(&sha256);
                    LOG_DEBUG("Ledger: txType: %s", txType);
                    SHA256_Update(&sha256, (const uint8_t*)txType, strlen(txType));
                    LOG_DEBUG("Ledger: encoded_args: %s", encoded_args.c_str());
                    SHA256_Update(
                        &sha256, (const uint8_t*)encoded_args.c_str(), encoded_args.size());
                    LOG_DEBUG("Ledger: response_data len: %d", response_len);
                    SHA256_Update(&sha256, response_data, response_len);

                    // hash read and write set
                    LOG_DEBUG("Ledger: read_set:");
                    for (auto& it : ecc_read_set)
                    {
                        LOG_DEBUG("\\-> %s", it.c_str());
                        SHA256_Update(&sha256, (const uint8_t*)it.c_str(), it.size());
                    }

                    LOG_DEBUG("Ledger: write_set:");
                    for (auto& it : ecc_write_set)
                    {
                        LOG_DEBUG("\\-> %s - %s", it.first.c_str(), it.second.c_str());
                        SHA256_Update(&sha256, (const uint8_t*)it.first.c_str(), it.first.size());
                        SHA256_Update(&sha256, (const uint8_t*)it.second.c_str(), it.second.size());
                    }
                    SHA256_Final(hash, &sha256);

                    std::string base64_hash = base64_encode((const unsigned char*)hash, 32);
                    LOG_DEBUG("Ledger: ecc sig hash (base64): %s", base64_hash.c_str());

                    // hash again!!!! Note that sgx_ecdsa_sign hashes the data
                    // input
                    uint8_t hash2[HASH_SIZE];
                    SHA256_Init(&sha256);
                    SHA256_Update(&sha256, hash, HASH_SIZE);
                    SHA256_Final(hash2, &sha256);

                    std::string base64_hashhash = base64_encode((const unsigned char*)hash2, 32);
                    LOG_DEBUG("Ledger: ecc sig hash-hash (base64): %s", base64_hashhash.c_str());

                    std::string base64_pk = base64_encode((const unsigned char*)pk, pk_len);
                    LOG_DEBUG("Ledger: ecc sig pk (base64): %s", base64_pk.c_str());

                    const unsigned char* sig_ptr = signature;
                    const unsigned char* pk_ptr = pk;
                    int err = verify_enclave_signature(
                        &sig_ptr, signature_len, hash2, 32, &pk_ptr, pk_len);
                    if (err != 1)
                    {
                        LOG_ERROR("Ledger: ecc signature validation failed (err=%d)", err);
                        // TODO mark as invalid
                    }
                    else
                    {
                        LOG_DEBUG("Ledger: ecc signature is valid :)");
                    }

                    // TODO check pk is valid (registered at ercc)
                }
            }
            pb_release(rwset_TxReadWriteSet_fields, &tx_rw_set);
        }

        // if there are any events let's parse them
        if (cc_action.events != NULL)
        {
            LOG_DEBUG("Ledger: \t\t \\-> ChaincodeAction.Events:");
        }

        pb_release(protos_ChaincodeInvocationSpec_fields, &cis);
        pb_release(protos_ChaincodeProposalPayload_fields, &proposal_payload);
        pb_release(protos_ChaincodeAction_fields, &cc_action);
        pb_release(protos_ProposalResponsePayload_fields, &p_response_payload);
        pb_release(protos_ChaincodeActionPayload_fields, &action_payload);

    }  // all actions
    pb_release(protos_Transaction_fields, &transaction);

    return LEDGER_SUCCESS;
}

// return true for a conflict
int has_version_conflict(const std::string& key, kvs_t* state, version_t* v)
{
    kvs_iterator_t it = state->find(key);
    if (it != state->end())
    {
        version_t version = it->second.second;
        // check if read.version is less than the state
        if (cmp_version(v, &version) == -1)
        {
            return 1;
        }
    }
    return 0;
}

int print_state()
{
    LOG_DEBUG("Ledger: ### Print state ###");
    for (auto& pair : state)
    {
        LOG_DEBUG("Ledger: \\-> Key: \"%s\" => version: (%d,%d)", pair.first.c_str(),
            pair.second.second.block_num, pair.second.second.tx_num);
    }
    return LEDGER_SUCCESS;
}

int ledger_get_state_hash(const char* key, uint8_t* out_hash)
{
    std::string k(key);
    LOG_DEBUG("Ledger: Search %s", k.c_str());

    spin_lock(&lock);
    auto iter = state.find(key);
    if (iter != state.end())
    {
        const kvs_value_t& value = iter->second;
        LOG_DEBUG("Ledger: \\-> Found Key: \"%s\" => version: (%d,%d)", iter->first.c_str(),
            value.second.block_num, value.second.tx_num);

        // hash item
        SHA256_CTX sha256;
        SHA256_Init(&sha256);
        SHA256_Update(&sha256, (const uint8_t*)value.first.c_str(), value.first.size());
        SHA256_Final(out_hash, &sha256);
        spin_unlock(&lock);
        return LEDGER_SUCCESS;
    }
    spin_unlock(&lock);

    LOG_DEBUG("Ledger: %s not found!", k.c_str());
    return LEDGER_NOT_FOUND;
}

int ledger_get_multi_state_hash(const char* comp_key, uint8_t* out_hash)
{
    const std::string k(comp_key);
    LOG_DEBUG("Ledger: Search %s", k.c_str());

    SHA256_CTX sha256;
    SHA256_Init(&sha256);

    spin_lock(&lock);
    auto p = state.lower_bound(k);
    for (auto iter = p; iter != state.end(); ++iter)
    {
        std::string key = iter->first;
        const kvs_value_t& value = iter->second;

        // if has no prefix anymore abort
        if (key.compare(0, k.size(), k) != 0)
        {
            break;
        }

        // remove channel name prefix from key if exists
        size_t found = key.find_first_of(".");
        if (found != std::string::npos)
        {
            key.erase(0, found);
        }

        SHA256_Update(&sha256, (const uint8_t*)key.c_str(), key.size());
        SHA256_Update(&sha256, (const uint8_t*)value.first.c_str(), value.first.size());
    }
    spin_unlock(&lock);

    SHA256_Final(out_hash, &sha256);
    return LEDGER_SUCCESS;
}
