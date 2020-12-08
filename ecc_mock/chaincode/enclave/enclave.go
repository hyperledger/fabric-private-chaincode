package enclave

import "C"
import (
	"context"
	"encoding/json"
	"fmt"
	"unsafe"

	"github.com/hyperledger-labs/fabric-private-chaincode/ecc/crypto"
	"github.com/hyperledger-labs/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"golang.org/x/sync/semaphore"
)

// #cgo CFLAGS: -I${SRCDIR}/ecc-enclave-include -I${SRCDIR}/../../../common/sgxcclib
// #cgo LDFLAGS: -L${SRCDIR}/ecc-enclave-lib -lsgxcc
// #include "common-sgxcclib.h"
// #include "sgxcclib.h"
//
import "C"

const enclaveLibFile = "enclave/lib/enclave.signed.so"
const maxResponseSize = 1024 * 100 // Let's be really conservative ...

type EnclaveStub struct {
	eid C.enclave_id_t
	sem *semaphore.Weighted
}

// NewEnclave starts a new enclave
func NewEnclaveStub() StubInterface {
	return &EnclaveStub{sem: semaphore.NewWeighted(8)}
}

func (e *EnclaveStub) Init(chaincodeParams, hostParams, attestationParams []byte) ([]byte, error) {
	var eid C.enclave_id_t

	// prepare output buffer for credentials
	credentialsBuffer := C.malloc(maxResponseSize)
	defer C.free(credentialsBuffer)
	credentialsSize := C.uint32_t(0)

	err := e.sem.Acquire(context.Background(), 1)
	if err != nil {
		return nil, err
	}

	// call the enclave
	ret := C.sgxcc_create_enclave(
		&eid,
		C.CString(enclaveLibFile),
		(*C.uint8_t)(C.CBytes(attestationParams)),
		C.uint32_t(len(attestationParams)),
		(*C.uint8_t)(C.CBytes(chaincodeParams)),
		C.uint32_t(len(chaincodeParams)),
		(*C.uint8_t)(C.CBytes(hostParams)),
		C.uint32_t(len(hostParams)),
		(*C.uint8_t)(credentialsBuffer),
		C.uint32_t(maxResponseSize),
		&credentialsSize)

	if ret != 0 {
		msg := fmt.Sprintf("can not create enclave (%s): Reason: %v", enclaveLibFile, ret)
		logger.Error(msg)
		return nil, fmt.Errorf(msg)
	}
	e.eid = eid
	e.sem.Release(1)
	logger.Infof("Enclave created with eid=%d", e.eid)

	// return credential bytes from sgx call
	return C.GoBytes(credentialsBuffer, C.int(credentialsSize)), nil
}

func (e *EnclaveStub) GenerateCCKeys() (*protos.SignedCCKeyRegistrationMessage, error) {
	panic("implement me")
}

func (e *EnclaveStub) ExportCCKeys(credentials *protos.Credentials) (*protos.SignedExportMessage, error) {
	panic("implement me")
}

func (e *EnclaveStub) ImportCCKeys() (*protos.SignedCCKeyRegistrationMessage, error) {
	panic("implement me")
}

func (e *EnclaveStub) GetEnclaveId() (string, error) {
	panic("implement me")
}

func (e *EnclaveStub) ChaincodeInvoke(stub shim.ChaincodeStubInterface) ([]byte, error) {

	index := registry.Register(&Stubs{stub})
	defer registry.Release(index)
	ctx := unsafe.Pointer(&index)

	// get and json-encode parameters
	// Note: client side call of '{ "Args": [ arg1, arg2, .. ] }' and '{ "Function": "arg1", "Args": [ arg2, .. ] }' are identical ...
	jsonArgs, err := json.Marshal(stub.GetStringArgs())
	if err != nil {
		return nil, err
	}

	// args
	argsPtr := C.CString(string(jsonArgs))
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
	if invoke_ret == 0 {
		sig, err = crypto.MarshalEnclaveSignature(C.GoBytes(signaturePtr, C.int(SIGNATURE_SIZE)))
		if err != nil {
			sig = nil
		}
	} else {
		err = fmt.Errorf("Invoke failed. Reason: %d", int(invoke_ret))
		// TODO: ideally we would also sign error messages but would
		// require including the error into the signature itself
		// which has involves a rathole of changes, so defer to the
		// time which design & refactor everything to be end-to-end
		// secure ...
	}
	return C.GoBytes(responsePtr, C.int(responseLenOut)), sig, err
}
