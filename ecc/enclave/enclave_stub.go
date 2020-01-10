/*
Copyright 2019 Intel Corporation
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

/*
TODO:
- add everywhere explicit (& consistent) return errors to ocalls/ecalls
*/

package enclave

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"sync"
	"unsafe"

	"github.com/hyperledger-labs/fabric-private-chaincode/ecc/crypto"
	"github.com/hyperledger-labs/fabric-private-chaincode/ecc/tlcc"
	sgx_utils "github.com/hyperledger-labs/fabric-private-chaincode/utils"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"golang.org/x/sync/semaphore"

	"crypto/x509"
	"encoding/pem"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/protos/msp"
)

// #cgo CFLAGS: -I${SRCDIR}/ecc-enclave-include -I${SRCDIR}/../../common/sgxcclib
// #cgo LDFLAGS: -L${SRCDIR}/ecc-enclave-lib -lsgxcc
// #include "common-sgxcclib.h"
// #include "sgxcclib.h"
// #include <stdio.h>
// #include <string.h>
//
// /*
//    Below extern definitions should really be done by cgo but without we get following warning:
//       warning: implicit declaration of function ▒▒▒_GoStringPtr_▒▒▒ [-Wimplicit-function-declaration]
// */
// extern const char *_GoStringPtr(_GoString_ s);
// extern size_t _GoStringLen(_GoString_ s);
//
// static inline void _cpy_bytes(uint8_t* target, uint8_t* val, uint32_t size)
// {
//   memcpy(target, val, size);
// }
//
// static inline void _set_int(uint32_t* target, uint32_t val)
// {
//   *target = val;
// }
//
// static inline void _cpy_str(char* target, _GoString_ val, uint32_t max_size)
// {
//   #define MIN(x, y) (((x) < (y)) ? (x) : (y))
//   // Note: have to do MIN as _GoStringPtr() might not be NULL terminated.
//   // Also _GoStringLen returns the length without \0 ...
//   size_t goStrLen = _GoStringLen(val)+1;
//   snprintf(target, MIN(max_size, goStrLen), "%s", _GoStringPtr(val));
// }
//
import "C"

const EPID_SIZE = 8
const SPID_SIZE = 16
const MAX_RESPONSE_SIZE = 1024 * 100 // Let's be really conservative ...
const SIGNATURE_SIZE = 64
const PUB_KEY_SIZE = 64
const TARGET_INFO_SIZE = 512
const CMAC_SIZE = 16
const ENCLAVE_TCS_NUM = 8

var logger = flogging.MustGetLogger("ecc_enclave")

// just a container struct used for the callbacks
type Stubs struct {
	shimStub shim.ChaincodeStubInterface
	tlccStub tlcc.TLCCStub
}

// have a global registry
var registry = NewRegistry()

// used to store shims for callbacks
type Registry struct {
	sync.RWMutex
	index    int
	internal map[int]*Stubs
}

func NewRegistry() *Registry {
	return &Registry{
		internal: make(map[int]*Stubs),
	}
}

func (r *Registry) Register(stubs *Stubs) int {
	r.Lock()
	defer r.Unlock()
	r.index++
	r.internal[r.index] = stubs
	return r.index
}

func (r *Registry) Release(i int) {
	r.Lock()
	delete(r.internal, i)
	r.Unlock()
}

func (r *Registry) Get(i int) *Stubs {
	r.RLock()
	stubs, ok := r.internal[i]
	r.RUnlock()
	if !ok {
		panic(fmt.Errorf("No shim for: %d", i))
	}
	return stubs
}

//export golog
func golog(str *C.char) {
	logger.Infof("%s", C.GoString(str))
}

// TODO: this seems dead-code? remove?
var _logger = func(in string) {
	logger.Info(in)
}

//export get_creator_name
func get_creator_name(msp_id *C.char, max_msp_id_len C.uint32_t, dn *C.char, max_dn_len C.uint32_t, ctx unsafe.Pointer) {
	stubs := registry.Get(*(*int)(ctx))

	// TODO (eventually): replace/simplify below via ext.ClientIdentity,
	// should also make it easier to eventually return more than only
	// msp & dn ..

	serializedID, err := stubs.shimStub.GetCreator()
	if err != nil {
		panic("error while getting creator")
	}
	sId := &msp.SerializedIdentity{}
	err = proto.Unmarshal(serializedID, sId)
	if err != nil {
		panic("Could not deserialize a SerializedIdentity")
	}

	bl, _ := pem.Decode(sId.IdBytes)
	if bl == nil {
		panic("Failed to decode PEM structure")
	}
	cert, err := x509.ParseCertificate(bl.Bytes)
	if err != nil {
		panic("Unable to parse certificate %s")
	}

	var goMspId = sId.Mspid
	C._cpy_str(msp_id, goMspId, max_msp_id_len)

	var goDn = cert.Subject.String()
	C._cpy_str(dn, goDn, max_dn_len)
	// TODO (eventually): return the eror case of the dn buffer being too small
}

//export get_state
func get_state(key *C.char, val *C.uint8_t, max_val_len C.uint32_t, val_len *C.uint32_t, cmac *C.uint8_t, ctx unsafe.Pointer) {
	stubs := registry.Get(*(*int)(ctx))

	// check if composite key
	key_str := C.GoString(key)
	if sgx_utils.IsSGXCompositeKey(key_str, sgx_utils.SEP) {
		key_str = sgx_utils.TransformToCompositeKey(stubs.shimStub, key_str, sgx_utils.SEP)
	}

	data, err := stubs.shimStub.GetState(key_str)
	if err != nil {
		panic("error while getting state")
	}
	if C.uint32_t(len(data)) > max_val_len {
		C._set_int(val_len, C.uint32_t(0))
		// NOTE: there is currently no way to explicitly return an error
		// to distinguish from absence of key.  However, iff key exist
		// and we return an error, this should trigger an integrity
		// error, so the shim implicitly notice the difference.
		return
	}
	C._cpy_bytes(val, (*C.uint8_t)(C.CBytes(data)), C.uint32_t(len(data)))
	C._set_int(val_len, C.uint32_t(len(data)))

	// ask tlcc for verification
	// TODO note that TLCC is currently hardcoded
	genCMAC, err := stubs.tlccStub.VerifyState(stubs.shimStub, "tlcc", stubs.shimStub.GetChannelID(), key_str, nil, false)
	if err != nil {
		panic("error while getting cmac: " + err.Error())
	}
	C._cpy_bytes(cmac, (*C.uint8_t)(C.CBytes(genCMAC)), C.uint32_t(CMAC_SIZE))
}

//export put_state
func put_state(key *C.char, val unsafe.Pointer, val_len C.int, ctx unsafe.Pointer) {
	stubs := registry.Get(*(*int)(ctx))

	// check if composite key
	key_str := C.GoString(key)
	if sgx_utils.IsSGXCompositeKey(key_str, sgx_utils.SEP) {
		key_str = sgx_utils.TransformToCompositeKey(stubs.shimStub, key_str, sgx_utils.SEP)
	}

	if stubs.shimStub.PutState(key_str, C.GoBytes(val, val_len)) != nil {
		panic("error while putting state")
	}
}

//export get_state_by_partial_composite_key
func get_state_by_partial_composite_key(comp_key *C.char, values *C.uint8_t, max_values_len C.uint32_t, values_len *C.uint32_t, cmac *C.uint8_t, ctx unsafe.Pointer) {
	stubs := registry.Get(*(*int)(ctx))

	// split and get a proper composite key
	comp := sgx_utils.SplitSGXCompositeKey(C.GoString(comp_key), sgx_utils.SEP)
	iter, err := stubs.shimStub.GetStateByPartialCompositeKey(comp[0], comp[1:])
	if err != nil {
		panic("error while range query")
	}
	defer iter.Close()

	var buf bytes.Buffer
	buf.WriteString("[")
	for iter.HasNext() {
		item, err := iter.Next()
		if err != nil {
			panic("Error " + err.Error())
		}
		buf.WriteString("{\"key\":\"")
		buf.WriteString(sgx_utils.TransformToSGX(item.Key, sgx_utils.SEP))
		buf.WriteString("\",\"value\":\"")
		buf.Write(item.Value)
		if iter.HasNext() {
			buf.WriteString("\"},")
		} else {
			buf.WriteString("\"}")
		}
	}
	buf.WriteString("]")
	data := buf.Bytes()

	if C.uint32_t(len(data)) > max_values_len {
		C._set_int(values_len, C.uint32_t(0))
		// NOTE: there is currently no way to explicitly return an error
		// to distinguish from absence of key.  However, iff key exist
		// and we return an error, this should trigger an integrity
		// error, so the shim implicitly notice the difference.
		return
	}
	C._cpy_bytes(values, (*C.uint8_t)(C.CBytes(data)), C.uint32_t(len(data)))
	C._set_int(values_len, C.uint32_t(len(data)))

	// ask tlcc for verification
	genCMAC, err := stubs.tlccStub.VerifyState(stubs.shimStub, "tlcc", stubs.shimStub.GetChannelID(), C.GoString(comp_key), nil, true)
	if err != nil {
		panic("error while getting cmac: " + err.Error())
	}
	C._cpy_bytes(cmac, (*C.uint8_t)(C.CBytes(genCMAC)), C.uint32_t(CMAC_SIZE))
}

// Stub interface
type Stub interface {
	// Return quote and enclave PK in DER-encoded PKIX format
	GetRemoteAttestationReport(spid []byte, sig_rl []byte, sig_rl_size uint) ([]byte, []byte, error)
	// Return report and enclave PK in DER-encoded PKIX format
	GetLocalAttestationReport(targetInfo []byte) ([]byte, []byte, error)
	// Init chaincode
	Init(args []byte, shimStub shim.ChaincodeStubInterface, tlccStub tlcc.TLCCStub) ([]byte, []byte, error)
	// Invoke chaincode
	Invoke(args []byte, pk []byte, shimStub shim.ChaincodeStubInterface, tlccStub tlcc.TLCCStub) ([]byte, []byte, error)
	// Returns enclave PK in DER-encoded PKIX formatk
	GetPublicKey() ([]byte, error)
	// Creates an enclave from a given enclave lib file
	Create(enclaveLibFile string) error
	// Gets Enclave Target Information
	GetTargetInfo() ([]byte, error)
	// Bind to tlcc
	Bind(report, pk []byte) error
	// Destroys enclave
	Destroy() error
	// Returns expected MRENCLAVE
	MrEnclave() (string, error)
}

// StubImpl implements the interface
type StubImpl struct {
	eid C.enclave_id_t
	sem *semaphore.Weighted
}

// NewEnclave starts a new enclave
func NewEnclave() Stub {
	return &StubImpl{sem: semaphore.NewWeighted(ENCLAVE_TCS_NUM)}
}

// GetRemoteAttestationReport - calls the enclave for attestation, takes SPID as input
// and returns a quote and enclaves public key
func (e *StubImpl) GetRemoteAttestationReport(spid []byte, sig_rl []byte, sig_rl_size uint) ([]byte, []byte, error) {
	//sig_rl
	var sig_rlPtr unsafe.Pointer
	if sig_rl == nil {
		sig_rlPtr = unsafe.Pointer(sig_rlPtr)
	} else {
		sig_rlPtr = C.CBytes(sig_rl)
		defer C.free(sig_rlPtr)
	}

	// quote size
	quoteSize := C.uint32_t(0)
	ret := C.sgxcc_get_quote_size((*C.uint8_t)(sig_rlPtr), C.uint(sig_rl_size), (*C.uint32_t)(unsafe.Pointer(&quoteSize)))
	if ret != 0 {
		return nil, nil, fmt.Errorf("C.sgxcc_get_quote_size failed. Reason: %d", int(ret))
	}

	// pubkey
	pubkeyPtr := C.malloc(PUB_KEY_SIZE)
	defer C.free(pubkeyPtr)

	// spid
	spidPtr := C.CBytes(spid)
	defer C.free(spidPtr)

	// prepare quote space
	quotePtr := C.malloc(C.ulong(quoteSize))
	defer C.free(quotePtr)

	// call enclave
	e.sem.Acquire(context.Background(), 1)
	ret = C.sgxcc_get_remote_attestation_report(e.eid, (*C.quote_t)(quotePtr), C.uint32_t(quoteSize),
		(*C.ec256_public_t)(pubkeyPtr), (*C.spid_t)(spidPtr), (*C.uint8_t)(sig_rlPtr), C.uint32_t(sig_rl_size))
	e.sem.Release(1)
	if ret != 0 {
		return nil, nil, fmt.Errorf("C.sgxcc_get_remote_attestation_report failed. Reason: %d", int(ret))
	}

	// convert sgx format to DER-encoded PKIX format
	pk, err := crypto.MarshalEnclavePk(C.GoBytes(pubkeyPtr, C.int(PUB_KEY_SIZE)))
	if err != nil {
		return nil, nil, err
	}

	return C.GoBytes(quotePtr, C.int(quoteSize)), pk, nil
}

// GetLocalAttestationReport - calls the enclave for attestation, takes SPID as input
// and returns a quote and enclaves public key
func (e *StubImpl) GetLocalAttestationReport(spid []byte) ([]byte, []byte, error) {
	// NOT IMPLEMENTED YET
	return nil, nil, nil
}

// invoke calls the enclave for processing of (cc) init , takes arguments
// and the current chaincode state as input and returns a new chaincode state
func (e *StubImpl) Init(args []byte, shimStub shim.ChaincodeStubInterface, tlccStub tlcc.TLCCStub) ([]byte, []byte, error) {
	if shimStub == nil {
		return nil, nil, errors.New("Need shim")
	}

	// index := Register(Stubs{shimStub, tlccStub})
	index := registry.Register(&Stubs{shimStub, tlccStub})
	defer registry.Release(index)
	// defer Release(index)
	ctx := unsafe.Pointer(&index)

	// args
	argsPtr := C.CString(string(args))
	defer C.free(unsafe.Pointer(argsPtr))

	// response
	responseLenOut := C.uint32_t(0) // We pass maximal length separatedly; set to zero so we can detect valid responses
	responsePtr := C.malloc(MAX_RESPONSE_SIZE)
	defer C.free(responsePtr)

	// signature
	signaturePtr := C.malloc(SIGNATURE_SIZE)
	defer C.free(signaturePtr)

	e.sem.Acquire(context.Background(), 1)
	// invoke (init) enclave
	init_ret := C.sgxcc_init(e.eid,
		argsPtr,
		(*C.uint8_t)(responsePtr), C.uint32_t(MAX_RESPONSE_SIZE), &responseLenOut,
		(*C.ec256_signature_t)(signaturePtr),
		ctx)
	e.sem.Release(1)
	// Note: we do try to return the response in all cases, even then there is an error ...
	var sig []byte = nil
	var err error
	if init_ret == 0 {
		sig, err = crypto.MarshalEnclaveSignature(C.GoBytes(signaturePtr, C.int(SIGNATURE_SIZE)))
		if err != nil {
			sig = nil
		}
	} else {
		err = fmt.Errorf("Init failed. Reason: %d", int(init_ret))
		// TODO: ideally we would also sign error messages but would
		// require including the error into the signature itself
		// which has involves a rathole of changes, so defer to the
		// time which design & refactor everything to be end-to-end
		// secure ...
	}
	return C.GoBytes(responsePtr, C.int(responseLenOut)), sig, err
}

// invoke calls the enclave for transaction processing, takes arguments
// and the current chaincode state as input and returns a new chaincode state
func (e *StubImpl) Invoke(args []byte, pk []byte, shimStub shim.ChaincodeStubInterface, tlccStub tlcc.TLCCStub) ([]byte, []byte, error) {
	if shimStub == nil {
		return nil, nil, errors.New("Need shim")
	}

	// index := Register(Stubs{shimStub, tlccStub})
	index := registry.Register(&Stubs{shimStub, tlccStub})
	defer registry.Release(index)
	// defer Release(index)
	ctx := unsafe.Pointer(&index)

	// args
	argsPtr := C.CString(string(args))
	defer C.free(unsafe.Pointer(argsPtr))

	// client pk used for args encryption
	pkPtr := C.CString(string(pk))
	defer C.free(unsafe.Pointer(pkPtr))

	// response
	responseLenOut := C.uint32_t(0) // We pass maximal length separatedly; set to zero so we can detect valid responses
	responsePtr := C.malloc(MAX_RESPONSE_SIZE)
	defer C.free(responsePtr)

	// signature
	signaturePtr := C.malloc(SIGNATURE_SIZE)
	defer C.free(signaturePtr)

	e.sem.Acquire(context.Background(), 1)
	// invoke enclave
	invoke_ret := C.sgxcc_invoke(e.eid,
		argsPtr,
		pkPtr,
		(*C.uint8_t)(responsePtr), C.uint32_t(MAX_RESPONSE_SIZE), &responseLenOut,
		(*C.ec256_signature_t)(signaturePtr),
		ctx)
	e.sem.Release(1)
	// Note: we do try to return the response in all cases, even then there is an error ...
	var sig []byte = nil
	var err error
	if invoke_ret == 0 {
		sig, err = crypto.MarshalEnclaveSignature(C.GoBytes(signaturePtr, C.int(SIGNATURE_SIZE)))
		if err != nil {
			sig = nil
		}
	} else {
		err = fmt.Errorf("Invoke failed. Reason: %d", int(invoke_ret))
		// TODO: (see above Init for comment applying also here)
	}
	return C.GoBytes(responsePtr, C.int(responseLenOut)), sig, err
}

// GetPublicKey returns the enclave ec public key
func (e *StubImpl) GetPublicKey() ([]byte, error) {
	// pubkey
	pubkeyPtr := C.malloc(PUB_KEY_SIZE)
	defer C.free(pubkeyPtr)

	e.sem.Acquire(context.Background(), 1)
	// call enclave
	ret := C.sgxcc_get_pk(e.eid, (*C.ec256_public_t)(pubkeyPtr))
	e.sem.Release(1)
	if ret != 0 {
		return nil, fmt.Errorf("C.sgxcc_get_pk failed. Reason: %d", int(ret))
	}

	// convert sgx format to DER-encoded PKIX format
	return crypto.MarshalEnclavePk(C.GoBytes(pubkeyPtr, C.int(PUB_KEY_SIZE)))
}

// Create starts a new enclave instance
func (e *StubImpl) Create(enclaveLibFile string) error {
	var eid C.enclave_id_t
	e.sem.Acquire(context.Background(), 1)
	if ret := C.sgxcc_create_enclave(&eid, C.CString(enclaveLibFile)); ret != 0 {
		return fmt.Errorf("Can not create enclave (lib %s): Reason: %d", enclaveLibFile, ret)
	}
	e.eid = eid
	e.sem.Release(1)
	logger.Infof("Enclave created with %d", e.eid)
	return nil
}

func (e *StubImpl) GetTargetInfo() ([]byte, error) {
	targetInfoPtr := C.malloc(TARGET_INFO_SIZE)
	defer C.free(targetInfoPtr)

	e.sem.Acquire(context.Background(), 1)
	ret := C.sgxcc_get_target_info(e.eid, (*C.target_info_t)(targetInfoPtr))
	if ret != 0 {
		return nil, fmt.Errorf("C.sgxcc_get_target_info failed. Reason: %d", int(ret))
	}
	e.sem.Release(1)

	return C.GoBytes(targetInfoPtr, TARGET_INFO_SIZE), nil
}

func (e *StubImpl) Bind(report, pk []byte) error {
	// Attention!!!!
	// here we set the report and pk pointer to NULL if not provided
	if report == nil || pk == nil {
		logger.Infof("No report pk provided! Call bind with NULL")
		e.sem.Acquire(context.Background(), 1)
		C.sgxcc_bind(e.eid, (*C.report_t)(nil), (*C.ec256_public_t)(nil))
		e.sem.Release(1)
		return nil
	}

	reportPtr := C.CBytes(report)
	defer C.free(reportPtr)

	// TODO transform pk to sgx
	transPk, err := crypto.UnmarshalEnclavePk(pk)
	if err != nil {
		return err
	}
	pkPtr := C.CBytes(transPk)
	defer C.free(pkPtr)

	e.sem.Acquire(context.Background(), 1)
	C.sgxcc_bind(e.eid, (*C.report_t)(reportPtr), (*C.ec256_public_t)(pkPtr))
	e.sem.Release(1)
	return nil
}

// Destroy kills the current enclave instance
func (e *StubImpl) Destroy() error {
	ret := C.sgxcc_destroy_enclave(e.eid)
	if ret != 0 {
		return fmt.Errorf("C.sgxcc_destroy_enclave failed. Reason: %d", int(ret))
	}
	return nil
}

func (e *StubImpl) MrEnclave() (string, error) {
	binMrEnclave, err := ioutil.ReadFile("mrenclave")
	if err != nil {
		return "", fmt.Errorf("Error reading MrEnclave from file: Reason %s", err.Error())
	}

	if len(binMrEnclave) == 0 {
		return "", fmt.Errorf("Error reading MrEnclave from file: Reason mrenclave is empty")
	}

	return string(binMrEnclave), nil
}
