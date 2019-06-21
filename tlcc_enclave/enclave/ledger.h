/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#ifndef ledger_h
#define ledger_h

#include <cstdint>
#include <list>
#include <map>
#include <set>
#include <string>

#include "sgx_spinlock.h"

#define LEDGER_SUCCESS 0
#define LEDGER_ERROR_DECODING -1
#define LEDGER_INVALID_BLOCK_NUMBER -3
#define LEDGER_ERROR_CRYPTO -5
#define LEDGER_NOT_FOUND -9
#define LEDGER_ERROR_OUT_BUFFER_TOO_SMALL -10

#define LEDGER_VERIFICATION_FAILED 0
#define LEDGER_VERIFICATION_SUCCESS 1

#define HASH_SIZE 32

#define decode_pb(obj, obj_fields, data, data_size)                    \
    do                                                                 \
    {                                                                  \
        pb_istream_t stream = pb_istream_from_buffer(data, data_size); \
        bool status = pb_decode(&stream, obj_fields, &obj);            \
        if (!status)                                                   \
        {                                                              \
            LOG_ERROR("Ledger: Can not decode protobuffer");           \
            return LEDGER_ERROR_DECODING;                              \
        }                                                              \
    } while (0)

typedef sgx_spinlock_t spinlock_t;
#define spin_lock(lock) sgx_spin_lock(lock);
#define spin_unlock(lock) sgx_spin_unlock(lock);

typedef struct version
{
    uint64_t block_num;
    uint64_t tx_num;
} version_t;

typedef struct hash_value
{
    uint8_t h[HASH_SIZE];
} hash_value_t;

typedef std::pair<std::string, version_t> kvs_value_t;
typedef std::pair<std::string, kvs_value_t> kvs_item_t;
typedef std::map<std::string, kvs_value_t> kvs_t;
typedef std::map<std::string, kvs_value_t>::const_iterator kvs_iterator_t;

typedef std::map<std::string, std::string> write_set_t;
typedef std::set<std::string> read_set_t;

// v1 == v2 return 0
// v1 < v2 return -1
// v1 > v2 return 1
static int cmp_version(version_t* v1, version_t* v2)
{
    if (v1->block_num > v2->block_num)
    {
        return 1;
    }
    else if (v1->block_num < v2->block_num)
    {
        return -1;
    }
    else
    {
        if (v1->tx_num > v2->tx_num)
        {
            return 1;
        }
        else if (v1->tx_num < v2->tx_num)
        {
            return -1;
        }
        else
        {
            return 0;
        }
    }
}

int has_version_conflict(const std::string& key, kvs_t* state, version_t* v);

int parse_block(uint8_t* block_data, uint32_t block_data_len);
int parse_config(uint8_t* config_data, uint32_t config_data_len);
int parse_endorser_transaction(
    uint8_t* tx_data, uint32_t tx_data_len, kvs_t* updates, version_t* version);
int commit_state_updates(kvs_t* updates, const uint32_t block_sequence_number);

int print_state();

int init_ledger();
int free_ledger();

int ledger_get_state_hash(const char* key, uint8_t* hash);
int ledger_get_multi_state_hash(const char* comp_key, uint8_t* hash);
int ledger_verify_state(const char* key, uint8_t* hash, uint32_t hash_len);

#endif
