/*
 * Copyright 2020 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "shim_internals.h"
#include <mbusafecrt.h> /* for memcpy_s etc */
#include "error.h"
#include "logging.h"

/*
 * rwset_to_proto encodes the rwset in a nanopb/protobuf data structure.
 * Strings and binary values are allocated dynamically here.
 * The caller owns the structure and is responsible to free the memory through pb_release.
 */
bool rwset_to_proto(t_shim_ctx_t* ctx, fpc_FPCKVSet* fpc_rwset_proto)
{
    int ret;
    unsigned int i;

    COND2LOGERR(ctx == NULL || fpc_rwset_proto == NULL, "invalid input parameters");

    LOG_DEBUG("Prepare Fabric RWSet construction");

    // reset structure
    *fpc_rwset_proto = fpc_FPCKVSet_init_default;

    // initialize read sets (i.e., the arrays; later we serialize single items)
    fpc_rwset_proto->has_rw_set = true;
    fpc_rwset_proto->read_value_hashes_count = ctx->read_set.size();
    fpc_rwset_proto->read_value_hashes = (pb_bytes_array_t**)pb_realloc(
        NULL, fpc_rwset_proto->read_value_hashes_count * sizeof(pb_bytes_array_t*));
    COND2ERR(fpc_rwset_proto->read_value_hashes == NULL);

    fpc_rwset_proto->rw_set.reads_count = ctx->read_set.size();
    fpc_rwset_proto->rw_set.reads = (kvrwset_KVRead*)pb_realloc(
        NULL, fpc_rwset_proto->rw_set.reads_count * sizeof(kvrwset_KVRead));
    COND2ERR(fpc_rwset_proto->rw_set.reads == NULL);

    // initialiaze write sets (i.e., the arrays; later we serialize single items)
    fpc_rwset_proto->rw_set.writes_count = ctx->write_set.size() + ctx->del_set.size();
    fpc_rwset_proto->rw_set.writes = (kvrwset_KVWrite*)pb_realloc(
        NULL, fpc_rwset_proto->rw_set.writes_count * sizeof(kvrwset_KVWrite));
    COND2ERR(fpc_rwset_proto->rw_set.writes == NULL);

    LOG_DEBUG("Add read_set items");
    i = 0;
    for (auto it = ctx->read_set.begin(); it != ctx->read_set.end(); it++, i++)
    {
        LOG_DEBUG("k=%s , v(hex)=%s , v(len)=%d", it->first.c_str(),
            ByteArrayToHexEncodedString(it->second).c_str(), it->second.size());

        // serialize hash
        fpc_rwset_proto->read_value_hashes[i] =
            (pb_bytes_array_t*)pb_realloc(NULL, PB_BYTES_ARRAY_T_ALLOCSIZE(it->second.size()));
        COND2ERR(fpc_rwset_proto->read_value_hashes[i] == NULL);
        fpc_rwset_proto->read_value_hashes[i]->size = it->second.size();
        ret = memcpy_s(fpc_rwset_proto->read_value_hashes[i]->bytes,
            fpc_rwset_proto->read_value_hashes[i]->size, it->second.data(), it->second.size());
        COND2ERR(ret != 0);

        // serialize read
        fpc_rwset_proto->rw_set.reads[i].has_version = false;
        fpc_rwset_proto->rw_set.reads[i].key = (char*)pb_realloc(NULL, it->first.length() + 1);
        COND2ERR(fpc_rwset_proto->rw_set.reads[i].key == NULL);
        ret = memcpy_s(fpc_rwset_proto->rw_set.reads[i].key, it->first.length(), it->first.c_str(),
            it->first.length());
        fpc_rwset_proto->rw_set.reads[i].key[it->first.length()] = '\0';
        COND2ERR(ret != 0);
    }

    LOG_DEBUG("Add write_set items");
    i = 0;
    for (auto it = ctx->write_set.begin(); it != ctx->write_set.end(); it++, i++)
    {
        LOG_DEBUG(
            "k=%s , v(hex)=%s", it->first.c_str(), ByteArrayToHexEncodedString(it->second).c_str());

        // serialize write
        fpc_rwset_proto->rw_set.writes[i].is_delete = false;

        // serialize key
        fpc_rwset_proto->rw_set.writes[i].key = (char*)pb_realloc(NULL, it->first.length() + 1);
        COND2ERR(fpc_rwset_proto->rw_set.writes[i].key == NULL);
        ret = memcpy_s(fpc_rwset_proto->rw_set.writes[i].key, it->first.length(), it->first.c_str(),
            it->first.length());
        fpc_rwset_proto->rw_set.writes[i].key[it->first.length()] = '\0';
        COND2ERR(ret != 0);

        // serialize value
        fpc_rwset_proto->rw_set.writes[i].value =
            (pb_bytes_array_t*)pb_realloc(NULL, PB_BYTES_ARRAY_T_ALLOCSIZE(it->second.size()));
        COND2ERR(fpc_rwset_proto->rw_set.writes[i].value == NULL);
        fpc_rwset_proto->rw_set.writes[i].value->size = it->second.size();
        ret = memcpy_s(fpc_rwset_proto->rw_set.writes[i].value->bytes,
            fpc_rwset_proto->rw_set.writes[i].value->size, it->second.data(), it->second.size());
        COND2ERR(ret != 0);
    }

    LOG_DEBUG("Add del_set items");
    for (auto it = ctx->del_set.begin(); it != ctx->del_set.end(); it++, i++)
    {
        // note that we continue to use i without resetting since we are appending to the
        // rw_set.writes
        LOG_DEBUG("k=%s", it->c_str());

        // serialize write
        fpc_rwset_proto->rw_set.writes[i].is_delete = true;

        // serialize key
        fpc_rwset_proto->rw_set.writes[i].key = (char*)pb_realloc(NULL, it->length() + 1);
        COND2ERR(fpc_rwset_proto->rw_set.writes[i].key == NULL);
        ret = memcpy_s(
            fpc_rwset_proto->rw_set.writes[i].key, it->length(), it->c_str(), it->length());
        fpc_rwset_proto->rw_set.writes[i].key[it->length()] = '\0';
        COND2ERR(ret != 0);

        fpc_rwset_proto->rw_set.writes[i].value =
            (pb_bytes_array_t*)pb_realloc(NULL, PB_BYTES_ARRAY_T_ALLOCSIZE(0));
        COND2ERR(fpc_rwset_proto->rw_set.writes[i].value == NULL);
    }

    LOG_DEBUG("Fabric RWSet construction successful");
    return true;

err:
    return false;
}
