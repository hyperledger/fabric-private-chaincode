<!---
Licensed under Creative Commons Attribution 4.0 International License
https://creativecommons.org/licenses/by/4.0/
--->

# Ledger Enclave - FPC Stub Secure Channel Module

_This is a feature of (post-MVP) "Full" FPC and not part of the MVP FPC Lite variant._

This document describes the external APIs, the properties and the
high-level design of the module implementing a secure channel between
the FPC Stub (ECC_Enclave) and the Ledger Enclave (TLCC_Enclave)
enclaves. 

## Public API

_Note: all functions in this API are expected to retun 0 on success
and some non-zero value to denote failures. Specific error return
codes will be defined by the implementation as appropriate._

### ECC_Enclave

The secure channel modules exposes following public functions to the
core logic of ECC_Enclave via a C library.

```c++
typedef struct tl_ecc_ctx* tl_ecc_ctx_ptr_t;

// set up a secure channel between the fpc stub and trusted ledger enclaves.
// Notes:
// - u_shim_ctx is the handle to the go stub we pass with ecall_invoke so
//   corresponding ocalls can find the correct go stub object ..
// - iff non-NULL tlcc_mrenclave is passed, then setup is only successful
//   if peer attests to that identity. Otherwise, any valid attestation
//   will be accepted and the identity can be retrived using `tl_session_get_tlcc_mrenclave`
// - For properties and guarantees of the channel, see separate section below.
int tl_session_setup(
    void* u_shim_ctx,
    char* channel_id,
    char* chaincode_id,
    char* enclave_id,
    mrenclave_t* tlcc_mrenclave,
    tl_ecc_ctx_ptr_t* ecc_ctx);

// do an RPC to the trusted ledger over the secure channel
// Notes:
// - u_shim_ctx is the handle to the go stub we pass with ecall_invoke so
//   corresponding ocalls can find the correct go stub object ..
int tl_session_request(
    void* u_shim_ctx,
    tl_ecc_ctx_ptr_t ecc_ctx,
    uint8_t* request,
    uint32_t request_len
    uint8_t* response,
    uint32_t* response_len,
    uint32_t max_response_len);

// close session
int tl_session_close(
    void* u_shim_ctx,
    tl_ecc_ctx_ptr_t ecc_ctx
}

// return the channel id associated with an (established) session
// Notes
// - this is served from local context but by the session properties
//   must be consistent with tlcc_enclave's view
int tl_session_get_channel_id(
    tl_ecc_ctx_ptr_t ecc_ctx,
    char* channel_id,
    uint32_t* channel_id_len,
    uint32_t max_channel_id_len);

// return the channel hash associated with an (established) session
// Notes
// - this is served from local context but by the session properties
//   must be consistent with tlcc_enclave's view
int tl_session_get_channel_hash(
    tl_ecc_ctx_ptr_t ecc_ctx,
    channel_hash_t* channel_hash);

// return the mrenclave of authenticated TLCC_Enclave associated with an (established) session
// Notes
// - this is served from local context but by the session properties
//   must be consistent with tlcc_enclave's view
int tl_session_get_tlcc_mrenclave(
    tl_ecc_ctx_ptr_t ecc_ctx,
    mrenclave_t* tlcc_mrenclave);
```


### ECC

(Untrusted) ECC requires a library for the ocall-handlers plus a
callback to enable, in go, cc2scc to tlcc and related mapping from
shim_cxt to the actual go stub. The handler (and related registration)
presumably would be implemented in `enclave/enclave_stub.go`.
```c++
// ECC provided handler for handling cc2scc to tlcc
// Notes:
// - this cannot be done inside the module as we only get an opaque
//   handle to the (go) stub necessary to do cc2cc/cc2scc
typedef int (*tl_cc2scc_handler_t)(
    void* u_shim_ctx,
    uint8_t* request,
    uint32_t request_len
    uint8_t* response,
    uint32_t* response_len,
    uint32_t max_response_len);

// register a request processor function
int tl_session_register_cc2scc_handler(
    tl_cc2scc_handler_t handler_f);
```

### TLCC

C++ library with following interface. Presumably the caller will then
cgo-integrate it in is own go code.

```c++
// function called by Invoke dispatcher to forward session messages
// Note:
// - depending on how Invoke demultiplexes (e.g., based on protobuf type or
//   a separate string tag) this might also by SessionMsg instead of byte array ...
int tl_session_rpc(
    void *u_shim_ctx,
    uint8_t* request,
    uint32_t request_len,
    uint8_t* response,
    uint32_t* response_len,
    uint32_t max_response_len);
```


### TLCC_Enclave

```c++
typedef struct tl_tlcc_ctx* tl_tlcc_ctx_ptr_t;

// TLCC_Enclave provided handler for processing requests and providing corresponding responses
// Notes:
// - in case TLCC_Enclave handler doesn't have to do ocalls requiring the
//   untrusted shim context, that parameter could be dropped here (and
//   below in `tl_session_rpc`..
typedef int (*tl_request_handler_t)(
    void* u_shim_ctx,
    tl_tlcc_ctx_ptr_t ecc_ctx
    uint8_t* request,
    uint32_t request_len
    uint8_t* response,
    uint32_t* response_len,
    uint32_t max_response_len);

// register a request processor function
int tl_session_register_request_handler(
    tl_request_handler_t handler_f);

// return the channel_id associated with an (established) session
// Notes
// - this is served from local context but by the session properties
//   must be consistent with ecc_enclave's view
int tl_session_get_channel_id(
    tl_tlcc_ctx_ptr_t ecc_ctx,
    char* channel_id,
    uint32_t* channel_id_len,
    uint32_t max_channel_id_len);

// return the chaincode_id associated with an (established) session
// Notes
// - this is served from local context but by the session properties
//   must be consistent with ecc_enclave's view
int tl_session_get_chaincode_id(
    tl_tlcc_ctx_ptr_t ecc_ctx,
    char* chaincode_id,
    uint32_t* chaincode_id_len,
    uint32_t max_chaincode_id_len);

// return the enclave_id associated with an (established) session
// Notes
// - this is served from local context but by the session properties
//   must be consistent with ecc_enclave's view
int tl_session_get_enclave_id(
    tl_tlcc_ctx_ptr_t ecc_ctx,
    char* enclave_id);
    uint32_t* enclave_id_len,
    uint32_t max_enclave_id_len);
```


### Properties/Guarantees

- after successful `tl_session_setup` any `tl_session_request` causes `req`
  to be delivered to TLCC_Enclave in an _authentic_ (i.e. integrity protected) but _non
  confidential_ manner; similarly, any response `response` is delivered in an
  authentic manner as return parameter of `tl_session_request`.
- for any request, TLCC_Enclave and ECC_Enclave agree on the set
  `<chaincode_id, enclave_id, tlcc_mrenclave, channel_id, channel_hash>`
  as returned by the `tl_session_get_*` functions and their view on their
  own identity.
- lastly, the module ensures that for the duration of a single
  sessions, the interaction is between the same set of enclave
  _instances_. Or put differently, restarting a tlcc or ecc enclave
  will not allow to hijack a session and confuse the existing
  enclaves. More concretely, it means that all session state and in
  particular the session keys reside only ephemerally in private
  memory but are not persisted.

## Internals

_Note: Below internals are an initial design to validate feasability
of implementing above API but might be subject to change in the actual
implementation._


### Crypto

The actual message authentication code is abstracted below as
mac/MAC.  For actual choice we have following considerations:
- SGX SDK provides us with sgx_key_128bit_t as only choice
  (underlying assumption seems that one would use the key for AES128-GCM).
- that keylength would match CMAC. However, CMAC is not currently offered by PDO crypto.
- HMAC-SHA256 would from security perspective ask for 256-bit key
  (but "technically" supports shorter keys)

This would lead to following options:
1. add CMAC to pdo crypto
2. use HMAC with 128bit keys (and reduced security)
3. use HMAC & use a KDF to stretch the 128bit to 256bit (and better security over (2))


### Session State

#### ECC
```c++

    tl_cc2scc_handler_t cc2scc_handler;
```

#### ECC_Enclave
```c++
    #include <sgx_dh.h>

    // interal session context
    enum tl_ecc_state {INIT1, INIT2, ACTIVE, CLOSING, CLOSED} tl_ecc_state_t;

    struct tl_ecc_ctx {
        session_id_t session_id,
        tl_ecc_state_t state,
        sgx_dh_session_t dh
        sgx_key_128bit_t session_key
        char* channel_id,
        channel_hash_t channel_hash,
        char* chaincode_id,
        char* enclave_id,
        sgx_dh_session_enclave_identity_t tlcc_mrenclave,
    } tl_ecc_ctx_t;
```


#### TLCC_Enclave

```c++
  #include <sgx_dh.h>
    #include <map>

    // internal session context
    // - ACTIVE means tl_session_setup succeeded
    // - ACTIVE_VALIDATED means tl_session_setup succeeded, enclave_id is confirmed
    //   _and_ we have verified chaincode_id & ecc_mrenclave via ERCC
    typedef tl_tlcc_state enum{INIT, ACTIVE, ACTIVE_VALIDATED, CLOSED} tl_tlcc_state_t;

    struct tl_tlcc_ctx {
        uint64_t session_id,
        tl_tlcc_state_t state,
        sgx_dh_session_t dh,
        sgx_key_128bit_t session_key
        char* channel_id,
        channel_hash_t channel_hash,
        char* cc,
        char* enclave_id,
        sgx_dh_session_enclave_identity_t ecc_mrenclave,
    } tl_tlcc_ctx_t;


    // global variables
    // - configured request processor
    static tl_request_handler_t tl_req_handler;

    static uint64_t next_session_id = 0;

    // - in-flight sessions: map session-id -> context
    static std::map<uint64_t, tl_tlcc_ctx_t> tl_sessions;
 ```

### EDL

#### ECC<->ECC_Enclave

```
    untrusted {
        int ocall_tl_session_rpc(
                [in, size=request_len] uint8_t* request, uint32_t request_len,
                [out, size=max_response_len] uint8_t* response, uint32_t max_response_len,
                [out] uint32_t* response_len,
                [user_check] void *u_shim_ctx);
    };
```

#### TLCC<->TLCC_Enclave

```
    trusted {
        public int ecall_tl_session_rpc(
                [in, size=request_len] uint8_t* request, uint32_t request_len,
                [out, size=max_response_len] uint8_t* response, uint32_t max_response_len,
                [out] uint32_t* response_len,
                [user_check] void *u_shim_ctx);
    };
```

### Protocol Messages

See message definition in `protos/fpc/tl_session.proto`

### Execution Flow

#### Initialization

```
- tl_enclave calls tl_session_register_request_handler(&tl_fpc_stub_msg_processor);
- ecc calls tl_session_register_cc2scc_handler(&tl_fpc_cc2scc_forwarder);
```

#### Session Setup

The session setup is built around DH key-exchange provided by the SGX
SDK.
See [Developer Reference Guide v2.12](https://download.01.org/intel-sgx/latest/linux-latest/docs/Intel_SGX_Developer_Reference_Linux_2.12_Open_Source.pdf)
for more information; in particular, pages 97-102 for an overview and the
corresponding message flows, and page 121 and 305ff for the function definitions.

```
- ecc_enclave calls tl_session_setup(u_shim_ctx, channel_id, chaincode_id, enclave_id, tlcc_mrenclave, &ecc_ctx)
  # tlcc_mrenclave == NULL during enclave creation, otherwise it is value from CC_Params
  - ecc_ctx={session_id=NULL, state=INIT1, dh, channel_id=channel_id, channel_hash=NULL, tlcc_mrenclave, enclave_id=enclave_id, chaincode_id=chaincode_id, session_key=NULL}
  - sgx_dh_init_session(initiator, ecc_ctx.dh)
  - init_req = SessionMsg.stp_int_req{channel_id=ecc_ctx.channel_id, chaincode_id=ecc_ctx.chaincode_id, enclave_id=ecc_ctx.enclave_id}.serialize()
  - ocall_tl_session_rpc(u_shim_ctx, init_req)
    - ecc::session-library
      - initiates cc2scc(init_req) by calling cc2scc_handler(u_shim_ctx, ....)
        - tl::invoke dispatcher
          - ecall_tl_session_rpc(u_shim_ctx, init_req)
            - tl_enclave:session-library
              - init_req.deserialize()
              - dispatching based on type of init_req
                - tlcc_ctx = {session_id=next_session_id++, state=INIT, dh, channel_id=init_req.channel_id, channel_hash=channel_hash_for(tlcc_ctx.channel_id), ecc_mrenclave, enclave_id=init_req.enclave_id, chaincode_id=init_req.chaincode_id, session_key=NULL}
                - tl_sessions.insert(tlcc_ctx.session_id, tlcc_ctx)
                - sgx_dh_init_session(responder, tlcc_ctx.dh)
                - init_rsp = SessionMsg.stp_int_rsp{session_id=tlcc_ctx.session_id, msg1}
                - sgx_dh_responder_gen_msg1(init_rsp.msg1, tlcc_ctx.dh)
              - return(init_rsp.serialize()}
        - tl::invoke dispatcher (continued)
          - returns init_rsp as result of cc2scc
    - ecc::session-library (continued)
      - returns init_rsp.deserialize() as result of ocall
- ecc_enclave:tl_session_setup (continued)
  - ecc_ctx.session_id = init_rsp.session_id
  - compl_req = SessionMsg.stp_cmp_req{session_id=ecc_ctx.session_id, msg2}
  - sgx_dh_initiator_proc_msg1(init_rsp.msg1, compl_req.msg2, ecc_ctx.dh)
   - ecc_ctx.state = INIT2
  - ocall_tl_session_rpc(u_shim_ctx, compl_req.serialize())
    - ecc::session-library
      - initiates cc2scc(compl_req) by calling cc2scc_handler(u_shim_ctx, ....)
        - tl::invoke dispatcher
          - ecall_tl_session_rpc(u_shim_ctx, compl_req)
            - comp_req.deserialize()
            - tl_enclave:session-library
              - dispatching based on type of compl_req
                - tlcc_ctx = tl_sessions[compl_req.session_id]
                - compl_rsp = SessionMsg.stp_cmp_req{session_id=tlcc_ctx.session_id, msg3, channel_id=tlcc_ctx.channel_id, channel_hash=tlcc_ctx.channel_hash, chaincode_id=tlcc_ctx.chaincode_id, enclave_id=tlcc_ctx.enclave_id}
                - sgx_dh_responder_proc_msg2(compl_req.msg2, compl_rsp.msg3, tlcc_ctx.dh, tlcc_ctx.session_key, tlcc_ctx.ecc_mrenclave)
                  # Notes
                  # - tlcc_ctx.enclave_id is at this point still unvalidated!
                  #   Only a following request confirms the consistency with the ecc view!
                  # - as ecc might not yet be registered when opening the session,
                  #   we can also not yet confirm that tlcc_ctx.chaincode_id, tlcc_ctx.enclave_id, tlcc_ctx.ecc_mrenclave are correct and consistent with ercc
                  #   this validation, though, can be delayed until request handler calls a corresponding getter (which
                  #   by definition) must be during processing of a tx by which time ecc must be registered
                - tlcc_ctx.state = ACTIVE
              - compl_resp.mac = MAC(tlcc_ctx.session_key, compl_resp.payload)
             - return(compl_resp.serialize()}
        - tl::invoke dispatcher (continued)
          - returns compl_rsp as result of cc2scc
    - ecc::session-library (continued)
      - returns compl_rsp.deserialize() as result of ocall
- ecc_enclave:tl_session_setup (continued)
  - var tlcc_mrenclave;
  - sgx_dh_initiator_proc_msg3(compl_rsp.msg3, ecc_ctx, ecc_ctx.session_key, tlcc_mrenclave)
  - verify (tlcc_mrenclave==ecc_ctx.tlcc_mrenclave || ecc_ctx.tlcc_mrenclave==NULL)
  - ecc_ctx.tlcc_mrenclave = tlcc_mrenclave
  - verify (compl_resp.mac == MAC(ecc_ctx.session_key,compl_resp.payload)) &&
           (compl_resp.payload.enclave_id==ecc_ctx.enclave_id) &&
           (compl_resp.payload.channel_id==ecc_ctx.channel_id)
    # this validates that SessionSetupInitRequest was not tampered with
  - ecc_ctx.channel_hash = compl_rsp.channel_hash
  - ecc_ctx.state = ACTIVE
  - return success
```

#### Session Requests

```
- ecc_enclave calls tl_session_request(ecc_ctx, u_shim_ctx, request)
  - tx_req = SessionMsg.tx_req{session_id = ecc_ctx.session_id, request = request)
  - tx_req.nonce = random()
  - tx_req.mac = MAC(ecc_ctx.session_key, tx_req.payload)
  - ocall_tl_session_rpc(u_shim_ctx, tx_req.serialize())
    - ecc::session-library
      - initiates cc2scc(tx_req) by calling cc2scc_handler(u_shim_ctx, ....)
        - tl::invoke dispatcher
          - ecall_tl_session_rpc(u_shim_ctx, tx_req)
            - tx_req.deserialize()
            - tl_enclave:session-library
              - dispatching based on type of tx_req
                - tlcc_ctx = tl_sessions[tx_req.session_id]
                - verify (tx_req.mac == MAC(tlcc_ctx.session_key,tx_req.payload))
                - tx_rsp = SessionMsg.tx_rsp{session_id = tlcc_ctx.session_id, rsp)
                - tx_rsp.payload.response = tl_req_handler(tx_req.request)
  -             - tx_rsp.nonce = tx_req.nonce
                - tx_rsp.mac = MAC(tlcc_ctx.session_key, tx_rsp.payload)
                - return(req_resp.serialize()}
        - tl::invoke dispatcher (continued)
          - returns tx_rsp as result of cc2scc
    - ecc::session-library (continued)
      - returns tx_rsp.deserialize() as result of ocall
- ecc_enclave:tl_session_request (continued)
  - verify (tx_rsp.mac == MAC(ecc_ctx.session_key,tx_resp.payload)) &&
           (tx_rsp.nonce == tx_req.nonce)
  - return resp.payload.response
```

#### Session Teardown

```
- ecc_enclave calls tl_session_close(ecc_ctx, u_shim_ctx)
  - cls_req = SessionMsg.cls_req{session_id = ecc_ctx.session_id)
  - cls_req.mac = MAC(ecc_ctx.session_key, cls_req.payload)
  - ecc_ctx.state = CLOSING
  - ocall_tl_session_rpc(u_shim_ctx, cls_req.serialize())
    - ecc::session-library
      - initiates cc2scc(cls_req) by calling cc2scc_handler(u_shim_ctx, ....)
        - tl::invoke dispatcher
          - ecall_tl_session_rpc(u_shim_ctx, cls_req)
            - cls_req.deserialize()
            - tl_enclave:session-library
              - dispatching based on type of cls_req
                - tlcc_ctx = tl_sessions[cls_req.session_id]
                - verify (cls_req.mac == MAC(tlcc_ctx.session_key,cls_req.payload))
                - cls_rsp = SessionMsg.cls_rsp{session_id = tlcc_ctx.session_id)
                - cls_rsp.mac = MAC(tlcc_ctx.session_key, cls_rsp.payload)
                - tlcc_ctx.state = CLOSED
                - tl_sessions.erase(cls_req.session_id)
                - return(req_resp.serialize()}
        - tl::invoke dispatcher (continued)
          - returns cls_rsp as result of cc2scc
    - ecc::session-library (continued)
      - returns cls_rsp.deserialize() as result of ocall
- ecc_enclave:tl_session_request (continued)
  - verify (cls_rsp.mac == MAC(ecc_ctx.session_key,cls_resp.payload))
  - ecc_ctx.state = CLOSED
```
