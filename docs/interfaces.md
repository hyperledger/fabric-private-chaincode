

# FPC Interfaces

This document defines the interfaces exposed by the FPC components.

## ERCC

The ERCC interface is implemented through a "normal" chaincode interface.

```go
// chaincode interface (exposed to ecc and clients) implemented by invoke
func Invoke(stub shim.ChaincodeStubInterface) pb.Response {}

// note that the apiKey is an optional argument; if not present here, it will be accessed through the decorator
registerEnclave(enclavePkBase64 string, quoteBase64 string, apiKey string) error

getAttestationReport(enclavePkHashBase64 string) (attestationReport []byte)

getSPID() (spid []byte)
```
See https://github.com/hyperledger-labs/fabric-private-chaincode/blob/master/ercc/ercc.go

## TLCC

The TLCC interface is implemented through a "normal" chaincode interface.

```go
// chaincode interface (exposed to ecc and others) implemented by invoke
func Invoke(stub shim.ChaincodeStubInterface) pb.Response {}

GET_LOCAL_ATT_REPORT(targetInfo []byte) (reportBase64 string, enclavePkBase64 string) as json string

VERIFY_STATE(key string, nonce string, isRangeQuery bool) (cmacBase64 string)

// used by the admin to join the channel
JOIN_CHANNEL() string
```
See https://github.com/hyperledger-labs/fabric-private-chaincode/blob/master/tlcc/tlcc.go

```c++
// interface expose to TLCC_enclave
void ocall_print_string([in, string] const char *str);
```
See https://github.com/hyperledger-labs/fabric-private-chaincode/blob/master/tlcc_enclave/enclave/enclave.edl


## TLCC_Enclave
```c++
// common
public int ecall_init(void);

public int ecall_create_report(
        [in] const sgx_target_info_t *target_info,
        [out] sgx_report_t *report,
        [out, size=64] uint8_t *pubkey);

public int ecall_get_pk([out, size=64] uint8_t *pubkey);

// tlcc specific
public int ecall_join_channel(
        [in, size=gen_len] uint8_t *genesis, uint32_t gen_len);

public int ecall_next_block(
        [user_check] uint8_t *block_bytes, uint32_t block_size);

public int ecall_print_state(void);

public int ecall_get_state_metadata(
        [in, string] const char *key,
        [in, size=32] uint8_t *nonce,
        [out] sgx_cmac_128bit_tag_t *cmac);

public int ecall_get_multi_state_metadata(
        [in, string] const char *comp_key,
        [in, size=32] uint8_t *nonce,
        [out] sgx_cmac_128bit_tag_t *cmac);
```
See https://github.com/hyperledger-labs/fabric-private-chaincode/blob/master/common/enclave/common.edl

See https://github.com/hyperledger-labs/fabric-private-chaincode/blob/master/tlcc_enclave/enclave/enclave.edl

## ECC

The ECC interface is implemented through a "normal" chaincode interface. Methods with leading underscores
are treated as FPC commands. Normal `invoke` invocations are forwarded to a FPC chaincode enclave.

```go
// chaincode interface (exposed to admin/clients) implemented by invoke
func Invoke(stub shim.ChaincodeStubInterface) pb.Response {}

__setup (erccName string) (enclavePkBase64 []byte)
__getEnclavePk () (utils.Response{PublicKey: enclavePk})
```
See https://github.com/hyperledger-labs/fabric-private-chaincode/blob/master/ecc/enclave_chaincode.go

This interface is implemented by ECC to let a chaincode enclave call into the peer

```c++
void ocall_print_string([in, string] const char *str);

void ocall_get_creator_name(
        [out, size=max_msp_id_len] char *msp_id, uint32_t max_msp_id_len,
        [out, size=max_dn_len] char *dn, uint32_t max_dn_len,
        [user_check] void *u_shim_ctx);

void ocall_get_state(
        [in, string] const char *key,
        [out, size=max_val_len] uint8_t *val, uint32_t max_val_len, [out] uint32_t *val_len,
        [in, out] sgx_cmac_128bit_tag_t *cmac,
        [user_check] void *u_shim_ctx);

void ocall_put_state(
        [in, string] const char *key,
        [in, size=val_len] uint8_t *val, uint32_t val_len,
        [user_check] void *u_shim_ctx);

void ocall_get_state_by_partial_composite_key(
        [in, string] const char *comp_key,
        [out, size=max_len] uint8_t *values, uint32_t max_len, [out] uint32_t *values_len,
        [in, out] sgx_cmac_128bit_tag_t *cmac,
        [user_check] void *u_shim_ctx);
```
See https://github.com/hyperledger-labs/fabric-private-chaincode/blob/master/ecc_enclave/enclave/enclave.edl

## ECC_Enclave

The ECC_Enclave interface specifies interface of an FPC chaincode enclave. 

```c++
// defined common
public int ecall_init(void);

public int ecall_create_report(
        [in] const sgx_target_info_t *target_info,
        [out] sgx_report_t *report,
        [out, size=64] uint8_t *pubkey);

// returns the enclave's public key
public int ecall_get_pk([out, size=64] uint8_t *pubkey);

// ecc specific
public int ecall_bind_tlcc(
        [in] const sgx_report_t *report,
        [in, size=64] const uint8_t *pubkey);

// invoke a FPC chaincode
public int ecall_cc_invoke(
        [in, string] const char *encoded_args,
        [in, string] const char *pk,
        [out, size=response_len_in] uint8_t *response, uint32_t response_len_in,
        [out] uint32_t *response_len_out,
        [out] sgx_ec256_signature_t *signature,
        [user_check] void *u_shim_ctx);
```
See https://github.com/hyperledger-labs/fabric-private-chaincode/blob/master/common/enclave/common.edl

See https://github.com/hyperledger-labs/fabric-private-chaincode/blob/master/ecc_enclave/enclave/enclave.edl

## FPC_SHIM

This interface is exposed to a FPC chaincode. The chaincode must implement `invoke` and can access the ledger state using the corresponding access methods.

```c++
// must be implemented by a FPC chaincode
int invoke(uint8_t* response,
    uint32_t max_response_len,
    uint32_t* actual_response_len,
    shim_ctx_ptr_t ctx);

// exposed to FPC chaincode
void put_state(const char* key, uint8_t* val, uint32_t val_len, shim_ctx_ptr_t ctx);

void get_state(
    const char* key, uint8_t* val, uint32_t max_val_len, uint32_t* val_len, shim_ctx_ptr_t ctx);

void get_state_by_partial_composite_key(
    const char* comp_key, std::map<std::string, std::string>& values, shim_ctx_ptr_t ctx);

void put_public_state(const char* key, uint8_t* val, uint32_t val_len, shim_ctx_ptr_t ctx);

void get_public_state(
    const char* key, uint8_t* val, uint32_t max_val_len, uint32_t* val_len, shim_ctx_ptr_t ctx);

void get_public_state_by_partial_composite_key(
    const char* comp_key, std::map<std::string, std::string>& values, shim_ctx_ptr_t ctx);

int get_string_args(std::vector<std::string>& argss, shim_ctx_ptr_t ctx);

int get_func_and_params(
    std::string& func_name, std::vector<std::string>& params, shim_ctx_ptr_t ctx);

void get_creator_name(char* msp_id,  // MSP id of organization to which transaction creator belongs
    uint32_t max_msp_id_len,         // size of allocated buffer for msp_id
    char* dn,                        // distinguished name of transaction creator
    uint32_t max_dn_len,             // size of allocated buffer for dn
    shim_ctx_ptr_t ctx);

extern int get_random_bytes(uint8_t* buffer, size_t length);

void log_critical(const char* format, ...);
void log_error(const char* format, ...);
void log_warning(const char* format, ...);
void log_notice(const char* format, ...);
void log_info(const char* format, ...);
void log_debug(const char* format, ...);
```
See https://github.com/hyperledger-labs/fabric-private-chaincode/blob/master/ecc_enclave/enclave/shim.h



# State
In order to understand the interfaces better; we also include here the component state.


## ERCC

```go
type EnclaveRegistryCC struct {
	ra  attestation.Verifier
	ias attestation.IntelAttestationService
}

// Intel IAS verification key (same as used by TLCC Enclave)
const IntelPubPEM = `...`
const iasURL = "https://api.trustedservices.intel.com/sgx/dev/attestation/v3/report"
```

The ercc state is entirely stored on the ledger state using `putState` operations.

Ledger state: 
```go
map[string][]byte
```
containing a serialzed (json string) `IASAttestationReport` under a hash of the enclave public key as base64 string `enclavePkHashBase64`.

```go
type IASAttestationReport struct {
	EnclavePk                   []byte `json:"EnclavePk"`
	IASReportSignature          string `json:"IASReport-Signature"`
	IASReportSigningCertificate string `json:"IASReport-Signing-Certificate"`
	IASReportBody               []byte `json:"IASResponseBody"`
}
```

See https://github.com/hyperledger-labs/fabric-private-chaincode/blob/master/ercc/attestation/ias.go#L57

## TLCC
An TLCC instance keeps a local reference of the enclave id. This reference is volatile. Also a reference to the peer is kept.

```go
type TrustedLedgerCC struct {
	enclave enclave.Stub
	peer    *peer.Peer
}

type StubImpl struct {
	eid C.enclave_id_t
}
```


## TLCC_Enclave
Enclave public/private ecdsa keys
```c++
// TLCC enclave public/private keys
sgx_ec256_private_t enclave_sk = {0};
sgx_ec256_public_t enclave_pk = {0};

// root cert store for orderer org
static X509_STORE* root_certs_orderer = NULL;
// root cert store for application orgs
static X509_STORE* root_certs_apps = NULL;

// ledger integrity-metadata state
static kvs_t state;      
static spinlock_t lock;  // state lock

// sequence number counter
static uint32_t sequence_number = -1;

// Intel IAS verification key
static const char* INTEL_PUB_PEM = ...

// used to create cmacs to exchange data with a chaincode enclave
static sgx_cmac_128bit_key_t session_key = {}
```


## ECC
An ECC instance keeps a local reference of the enclave id. This reference is volatile.

```go
type EnclaveChaincode struct {
	erccStub ercc.EnclaveRegistryStub
	tlccStub tlcc.TLCCStub
	enclave  enclave.Stub
	verifier crypto.Verifier
}
```

## ECC_Enclave
```c++
// Chaincode enclave public/private keys
sgx_ec256_private_t enclave_sk = {0};
sgx_ec256_public_t enclave_pk = {0};

// TLCC public key
sgx_ec256_public_t tlcc_pk = {0};

// a session key that is used to verify cmacs receiveed from TLCC
sgx_cmac_128bit_key_t session_key = {};

// ledger state encryption key
sgx_aes_gcm_128bit_key_t state_encryption_key = {};

// unused
static sgx_thread_mutex_t global_mutex = SGX_THREAD_MUTEX_INITIALIZER;
```
