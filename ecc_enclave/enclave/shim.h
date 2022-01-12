/*
 * Copyright IBM Corp. All Rights Reserved.
 * Copyright 2020 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#pragma once

#include <stdbool.h>
#include <map>
#include <string>
#include <vector>

#include "logging.h"

typedef struct t_shim_ctx* shim_ctx_ptr_t;
typedef std::vector<uint8_t> ByteArray;

/*
  FPC Lite Constraints
  =========================

  The 'FPC Lite' variant of FPC does not provide a proof-of-commitment-on-ledger
  and, hence, does not support rollback protection.  As a consequence,
  applications should _not_ release any sensitive information conditioned
  on private ledger data.
  For example: if at some point in time the FPC chaincode executed `put_state` of
  a key-value pair `<k,v>`, the developer should assume that the value `v` might be
  returned at any future execution of `get_state` over the key `k` -- no matter
  whether `<k,v>` was actually committed or not. Similarly, you also might _not_ get
  the value back, regardless of whether it was committed or not.

  Additionally, some of below functions have additional restrictions:
  - `get_state`/`put_state` can be securely supported only for a single key.
    Note, though, that for additional security and the hiding of access
    patterns, a single key is in general a good strategy.
    To illustrate the security issue with more than one key:
    Let us assume that the FPC chaincode writes two key-value
    pairs `<k1,v1>, <k2,v2>`. Also, in another execution, the chaincode
    writes `<k1,v1'>, <k2,v2'>`. In subsequent executions, the developer
    should assume that `get_state` of `k1` might return any value
    between `none`, `v1`, `v1'`, in _any possible combination_
    with `get_state` of `k2` returning `none`, `v2` or `v2'`.
  - the composite key variants are not supported
  - the value returned from `get_creator_name` will be unvalidated, i.e.,
    identity management has to be done on the application level.

*/

// Function which FPC chaincode has to implement
// ==================================================
// - invoke, called when a transaction query or invocation is executed
int invoke(uint8_t* response,
    uint32_t max_response_len,
    uint32_t* actual_response_len,
    shim_ctx_ptr_t ctx);

// Shim Function which FPC chaincode can use
// ==================================================
// TODO (eventually): more documention, e.g., on how are error handled?

// put/get state
//-------------------------------------------------

// Normal, encrypted state

// - store value located at val of size val_len under key key
//   Note:
//   - while the values are encrypted, the key will remain in clear text.
//     So care has to be taken by the programmer that the key doesn't leak
//     anything sensitive!
void put_state(const char* key, uint8_t* val, uint32_t val_len, shim_ctx_ptr_t ctx);

// - look for key and, if found, store it in val and return size in val_len.
//   val must be of size at least max_val_len and the query will fail
//   if the retrieved value would be larger.
//   Absence of key is denoted by val_len == 0 when the function returns.
//   Note:
//   - this function doesn't check whether this was also stored privately.
//     If it was stored with put_public_state, reading the key will fail
//     due to a decryption or decoding failure.
void get_state(
    const char* key, uint8_t* val, uint32_t max_val_len, uint32_t* val_len, shim_ctx_ptr_t ctx);

// - look for composite keys, i.e., return the set of keys and values which match the
//   provided composite (prefix) key comp_key
//   Note:
//   - this function doesn't check whether this was also stored privately.
//     If it was stored with put_public_state, reading the key will fail
//     due to a decryption or decoding failure.
void get_state_by_partial_composite_key(
    const char* comp_key, std::map<std::string, std::string>& values, shim_ctx_ptr_t ctx);

// Public, unencrypted state

// - store value located at val of size val_len under key key in unencrypted form
void put_public_state(const char* key, uint8_t* val, uint32_t val_len, shim_ctx_ptr_t ctx);
//   Note:
//   - As the value will be unencrypted, a transaction creator will be able to read
//     updated state before it is committed to the ledger (and hence allows to prevent
//     it from being committed). For certain types of updates this can lead to attacks,
//     similar to pre-maturely return results in the response before a new transaction
//     state is committed. To counter this, the chaincode programmer must deploy a
//     commit-then-reveal pattern where in a first transaction, the state is privately
//     updated and only in a second transaction, when the state update can be confirmed,
//     the information is released to the public (for put_public_state) and/or to the
//     transactor (in case the sensitive information is revealed in the response).

// - look for key and, if found, store it in val and return size in val_len.
//   val must be of size at least max_val_len and the query will fail
//   if the retrieved value would be larger.
//   Absence of key is denoted by val_len == 0 when the function returns.
//   Note:
//   - this function doesn't check whether this was also stored publically.
//     If not, it would return the encrypted value ....
void get_public_state(
    const char* key, uint8_t* val, uint32_t max_val_len, uint32_t* val_len, shim_ctx_ptr_t ctx);

// - look for composite keys, i.e., return the set of keys and values which match the
//   provided composite (prefix) key comp_key
//   Note:
//   - this function doesn't check whether this was also stored publically.
//     If not, it would return the encrypted value ....
void get_public_state_by_partial_composite_key(
    const char* comp_key, std::map<std::string, std::string>& values, shim_ctx_ptr_t ctx);

// - '*_public_state*' variant of above which does _not_ encrypt
//   This could potentially allow for broadcasting decisions to the public
//   (such as auction results) and provide long-term evidence of outcomes even
//   when the enclaves might have "died".
//   Extract info from the ledger other than a query to the chaincode should be
//   possible via queryBlock() and queryTransaction(); not as convenient as with a
//   chaincode query but still seems useful in practice ..?
//

// TODO (possible extensions): possible extension of above
// - '*_super_private_state*' a variant of above where data is stored encrypted
//   in a private data collection.
//   Fabric private data collection definitely can give a performance boost when
//   dealing with large data as the data doesn't have to go all to the orderer.
//   Enabling this, though, can be done under the cover configuration-driven
//   and does not have to be exposed in the API for that reason?
//   Are there any security reasons why we would want it data collection _and_
//   also have to expose it to the API (as opposed, say configuraton-driven).
//   E.g., are there meaningful cases where the app would differentiate between
//   encrypted data which must be (directly) on ledger whereas others must be in
//   the collection?
//   In the normal fabric case, which provides separate {Get,Put}PrivateData*
//   function this seems necssary due to some data being shared with chaincodes
//   which are not part of the collection but for us it really shouldn't make
//   much of a security difference (the only difference i can see that we would
//   hide access patterns from the orderers. However, peers are much more likely
//   to be access-pattern-attackers as they have inherently more vested interest
//   as CC participants and also do understand much more about the application
//   (hence been better able to exploit the side-channel). Insofar, it doesn't
//   seem to me worth to expose it, at least until wee have a concrete use-case
//   requiring it?
//
// - '*_semi_public_state*' variant of '*_super_private_state*' where data is
//   stored unencrypted in a private data collection
//
// - for composite-key function, the current fabric shim also have additional
//   utility functions such as createCompositeKey, splitCompositeKey: worth supporting? (Seems
//   though primarily syntactic sugar?
//
// - other functions: {get,set}StateValidationParameter, getHistoryForKey. Can/should we ignore?

// - records the given `key` to be deleted in the writeset.
void del_state(const char* key, shim_ctx_ptr_t ctx);

// retrieval for arguments
//-------------------------------------------------
// - retrieve the list of invocation parameters
int get_string_args(std::vector<std::string>& argss, shim_ctx_ptr_t ctx);
// - a different way to retrieve the invocation parameters as function followed by function
// parameters
//   returns -1 if not called with at least function name ..
int get_func_and_params(
    std::string& func_name, std::vector<std::string>& params, shim_ctx_ptr_t ctx);
// Note: both functions should work interchangely, also regardless whether used
// the '{ "args": [ ..] }' or the '{ "function": "name", "args": [ ..] }' syntax  with
// the '--ctor'/'-c' parameter of peer cli command for option 'chaincode [invoke|query].
// Behind the scenes the two variants are always treated as a list of args, with the first
// one being the function.

// transaction APIs
//-------------------------------------------------

// - getChannelID - returns the channel name (ID) the FPC chaincode enclave
//   Note that the channel ID is attested during enclave initialization.
void get_channel_id(std::string& channel_id, shim_ctx_ptr_t ctx);

// - getTxID
void get_tx_id(std::string& tx_id, shim_ctx_ptr_t ctx);

// - getTxTimestamp
// // TODO (possible extensions): enclave has no access to trusted time.  Time
// //   from client is apriori not trusted either. However, at least client has
// //   to commit and in some cases might be trusted
// //   - do endorsers do any cross-check of this value? (Probably makes sense
// //     only in supporting it if there is some plausibility test the endorsing
// //     peers agree)
// //   - Is it commonly used in fabric?
// #include <time.h>
// void get_tx_timestamp(struct timespec* ts,
//     shim_ctx_ptr_t ctx);
//
// - getBinding
// // TODO (possible extensions): from description it seems this is used for replay protection,
// //   though, from
// //     https://fabric-shim.github.io/release-1.4/fabric-shim.ChaincodeStub.html#getBinding
// //   it seems relevant only for some delegation/third-party signature verification
// //   - Is it commonly used in fabric? If not then we should ignore it
// //   - Is this something we can easily support (insecurely short-term / securely long-term)?
//
// // TODO: other tx-related apis which exist but probably doesn't make sense to support
// // - getTransient: if we encrypt everything, then everything is essentially Transient?

// - getSignedProposal - returns a serialized signed proposal
void get_signed_proposal(ByteArray& signed_proposal, shim_ctx_ptr_t ctx);

// - get_creator_name
//   return the distinguished name of the creator (the subject field of the creator cert)
//   as well as the msp_id of the corresponding organization.
//   Note:
//   - The name might be truncated (but guaranteed to be null-terminated)
//     if the provided buffer is too small.
void get_creator_name(char* msp_id,  // MSP id of organization to which transaction creator belongs
    uint32_t max_msp_id_len,         // size of allocated buffer for msp_id
    char* dn,                        // distinguished name of transaction creator
    uint32_t max_dn_len,             // size of allocated buffer for dn
    shim_ctx_ptr_t ctx);

// - get_creator - returns a serialized identity from the signed proposal
//   Note that the returned identity is not validated against the MSP/ledger since
//   the enclave does not have trustworthy data to do so.
void get_creator(ByteArray& creator, shim_ctx_ptr_t ctx);

// Chaincode to Chaincode
//---------------------------
// invokeChaincode
// TODO (possible extensions): Currently not supported (but eventually should)

// Source of Randomness
// --------------------
// Note-1: this is currently implemented as part of crypto.
// Note-2: many chaincode applications require a (secure) source of randomness.
// In Fabric, however, chaincodes with "independent" sources of randomness will produce different
// outputs. Therefore, in multi-endorser settings, endorsers will sign different transactions and,
// when the endorsement policy requires more than one signature, the policy check will simply fail.
// Single-endorser settings should work fine, provided that nobody requires to check/reproduce the
// output. Note-3: due to what highlighted in note-2, a chaincode's source of randomness should
// rather be securely provided by the Fabric infrastructure, ensuring that chaincodes running on
// different platforms can get access to the same random coins.
// ***WARNING***: we implement this function using the SGX random number generator, but we expect to
// upgrade it according to what highlighted in note-3.
extern int get_random_bytes(uint8_t* buffer, size_t length);

// logging
//-------------------------------------------------
void log_critical(const char* format, ...);
void log_error(const char* format, ...);
void log_warning(const char* format, ...);
void log_notice(const char* format, ...);
void log_info(const char* format, ...);
void log_debug(const char* format, ...);
// TODO
// - API design questions
//   - macro vs function?
//     - function looks kind of cleaner.
//     - macros would allow passing context such as file & linenumber
// - implementation
//   - i would remove the DO_* from logging.h and replace the printfs with a
//     golog which would pass the level to go. that way normal log-enablement
//     would work.
//   - a somewhat orthogonal question is whether in production mode
//     a chaincode should ever log anything to the outside, but i think
//     that should be handled on the implementation side hiden from the API here
//
// TODO: implemented above once questions are resolved and API is agreed ..
