/*
* Copyright IBM Corp. 2018 All Rights Reserved.
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package enclave

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"
	"unsafe"

	"golang.org/x/sync/semaphore"

	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"gitlab.zurich.ibm.com/sgx-dev/sgx-cc/ecc/crypto"
	"gitlab.zurich.ibm.com/sgx-dev/sgx-cc/ecc/tlcc"
	sgx_utils "gitlab.zurich.ibm.com/sgx-dev/sgx-cc/utils"
)

// #cgo CFLAGS: -I${SRCDIR}/include
// #cgo LDFLAGS: -L${SRCDIR}/lib -lsgxcc
// #include <sgxcclib.h>
// #include <string.h>
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
import "C"

const EPID_SIZE = 8
const SPID_SIZE = 16
const MAX_OUTPUT_SIZE = 1024
const MAX_RESPONSET_SIZE = 1024
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
	r.index++
	r.internal[r.index] = stubs
	r.Unlock()
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

var _logger = func(in string) {
	logger.Info(in)
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
func get_state_by_partial_composite_key(comp_key *C.char, values *C.uint8_t, max_vales_len C.uint32_t, values_len *C.uint32_t, cmac *C.uint8_t, ctx unsafe.Pointer) {
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
	GetRemoteAttestationReport(spid []byte) ([]byte, []byte, error)
	// Return report and enclave PK in DER-encoded PKIX format
	GetLocalAttestationReport(targetInfo []byte) ([]byte, []byte, error)
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
func (e *StubImpl) GetRemoteAttestationReport(spid []byte) ([]byte, []byte, error) {
	// quote
	quote_size := C.sgxcc_get_quote_size()
	quotePtr := C.malloc(C.ulong(quote_size))
	defer C.free(quotePtr)

	// pubkey
	pubkeyPtr := C.malloc(PUB_KEY_SIZE)
	defer C.free(pubkeyPtr)

	// spid
	spidPtr := C.CBytes(spid)
	defer C.free(spidPtr)

	e.sem.Acquire(context.Background(), 1)
	// call enclave
	// TODO read error
	C.sgxcc_get_remote_attestation_report(e.eid, (*C.quote_t)(quotePtr), quote_size,
		(*C.ec256_public_t)(pubkeyPtr), (*C.spid_t)(spidPtr))

	e.sem.Release(1)

	// convert sgx format to DER-encoded PKIX format
	pk, err := crypto.MarshalEnclavePk(C.GoBytes(pubkeyPtr, C.int(PUB_KEY_SIZE)))
	if err != nil {
		return nil, nil, err
	}

	return C.GoBytes(quotePtr, C.int(quote_size)), pk, nil
}

// GetLocalAttestationReport - calls the enclave for attestation, takes SPID as input
// and returns a quote and enclaves public key
func (e *StubImpl) GetLocalAttestationReport(spid []byte) ([]byte, []byte, error) {
	// NOT IMPLEMENTED YET
	return nil, nil, nil
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
	responseLenOut := C.uint32_t(MAX_RESPONSET_SIZE)
	responsePtr := C.malloc(MAX_RESPONSET_SIZE)
	defer C.free(responsePtr)

	// signature
	signaturePtr := C.malloc(SIGNATURE_SIZE)
	defer C.free(signaturePtr)

	e.sem.Acquire(context.Background(), 1)
	// invoke enclave
	// TODO read error
	ret := C.sgxcc_invoke(e.eid,
		argsPtr,
		pkPtr,
		(*C.uint8_t)(responsePtr), C.uint32_t(MAX_RESPONSET_SIZE), &responseLenOut,
		(*C.ec256_signature_t)(signaturePtr),
		ctx)
	e.sem.Release(1)
	if ret != 0 {
		return nil, nil, fmt.Errorf("Invoke failed. Reason: %d", int(ret))
	}

	sig, err := crypto.MarshalEnclaveSignature(C.GoBytes(signaturePtr, C.int(SIGNATURE_SIZE)))
	if err != nil {
		return nil, nil, err
	}
	return C.GoBytes(responsePtr, C.int(responseLenOut)), sig, nil
}

// GetPublicKey returns the enclave ec public key
func (e *StubImpl) GetPublicKey() ([]byte, error) {
	// pubkey
	pubkeyPtr := C.malloc(PUB_KEY_SIZE)
	defer C.free(pubkeyPtr)

	e.sem.Acquire(context.Background(), 1)
	// call enclave
	// TODO read error
	C.sgxcc_get_pk(e.eid, (*C.ec256_public_t)(pubkeyPtr))
	e.sem.Release(1)

	// convert sgx format to DER-encoded PKIX format
	return crypto.MarshalEnclavePk(C.GoBytes(pubkeyPtr, C.int(PUB_KEY_SIZE)))
}

// Create starts a new enclave instance
func (e *StubImpl) Create(enclaveLibFile string) error {
	var eid C.enclave_id_t
	// todo read error
	e.sem.Acquire(context.Background(), 1)
	if ret := C.sgxcc_create_enclave(&eid, C.CString(enclaveLibFile)); ret != 0 {
		return fmt.Errorf("Can not create enclave: Reason: %d", ret)
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
	C.sgxcc_get_target_info(e.eid, (*C.target_info_t)(targetInfoPtr))
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
	// todo read error
	C.sgxcc_destroy_enclave(e.eid)
	return nil
}
