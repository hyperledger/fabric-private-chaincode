/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#pragma once

#include <stdbool.h>
#include <map>
#include <string>
#include <vector>

typedef struct t_shim_ctx* shim_ctx_ptr_t;

// Function which FPC chaincode has to implement
// ==================================================
// - init, called when the chaincode is instantiated
int init(uint8_t* response,
    uint32_t max_response_len,
    uint32_t* actual_response_len,
    shim_ctx_ptr_t ctx);

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
// the '--ctor'/'-c' parameter of peer cli command for option 'chaincode [instantiate|invoke|query].
// Behind the scenes the two variants are always treated as a list of args, with the first
// one being the function.

// transaction APIs
//-------------------------------------------------

// - getChannelID
// // TOD0 (possible extensions): might be useful to support and should be easy?
// //     If this is just the name, would it be useful also to have a variant which
// //     has the unique id ("content-addressable"/genesis-block-hash)?
// void get_channel_id(char* channel_id,
//     uint32_t max_channel_id_len,
//     shim_ctx_ptr_t ctx);

// - TxID
// // TODO (possible extensions): at least coming from a Sawtooth/PDO perspective,
// //   i would think access to this info might be important for cross-cc transactions?
// //   - Is it commonly used in fabric?
// //   - Is this something we can easily support (insecurely short-term / securely long-term)?
// void get_tx_id(char* tx_id,
//     uint32_t max_tx_id_len,
//     shim_ctx_ptr_t ctx);

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
// // - getSignedProposal: should be easy to support but probably not worth?

// - creator
//   return the distinguished name of the creator as well as the msp_id of the corresponding
//   organization.
//   Note:
//   - The name might be truncated (but guaranteed to be null-terminated)
//     if the provided buffer is too small.
void get_creator_name(char* msp_id,  // MSP id of organization to which transaction creator belongs
    uint32_t max_msp_id_len,         // size of allocated buffer for msp_id
    char* dn,                        // distinguished name of transaction creator
    uint32_t max_dn_len,             // size of allocated buffer for dn
    shim_ctx_ptr_t ctx);
//
// TODO (eventually): The go shim GoCreator returns protobuf serialized identity which (usally)
// is the pair of msp_id and a (PEM-encoded) certificate. We might eventually add a function
// also to expose the certificate itself.  However, for most current use-cases the DN should
// be sufficient and makes CC-programming easier.

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
