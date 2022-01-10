//go:build !mock_ecc
// +build !mock_ecc

/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package enclave

import (
	"context"
	"fmt"
	"unsafe"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/protoutil"
	"golang.org/x/sync/semaphore"
)

// #cgo CFLAGS: -I${SRCDIR}/ecc-enclave-include -I${SRCDIR}/../../../common/sgxcclib
// #cgo LDFLAGS: -L${SRCDIR}/ecc-enclave-lib -lsgxcc -lgcov
// #include "common-sgxcclib.h"
// #include "sgxcclib.h"
//
import "C"

const enclaveLibFile = "enclave/lib/enclave.signed.so"

var logger = flogging.MustGetLogger("enclave")

// EnclaveStub translates invocations into an enclave using cgo
type EnclaveStub struct {
	eid           C.enclave_id_t
	sem           *semaphore.Weighted
	isInitialized bool
}

func NewEnclaveStub() *EnclaveStub {
	return &EnclaveStub{sem: semaphore.NewWeighted(8)}
}

func (e *EnclaveStub) Init(chaincodeParams, hostParams, attestationParams []byte) ([]byte, error) {
	// Estimate of the buffer length that is necessary for the credentials. It should be conservative.
	const credentialsBufferMaxLen = 16 * 1024

	if e.isInitialized {
		return nil, fmt.Errorf("enclave already initialized")
	}

	var eid C.enclave_id_t

	// prepare output buffer for credentials
	credentialsBuffer := C.malloc(credentialsBufferMaxLen)
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
		C.uint32_t(credentialsBufferMaxLen),
		&credentialsSize)

	if ret != 0 {
		msg := fmt.Sprintf("can not create enclave (%s): Reason: %v", enclaveLibFile, ret)
		logger.Error(msg)
		return nil, fmt.Errorf(msg)
	}
	e.eid = eid
	e.sem.Release(1)
	logger.Infof("Enclave created with eid=%d", e.eid)

	e.isInitialized = true

	// return credential bytes from sgx call
	return C.GoBytes(credentialsBuffer, C.int(credentialsSize)), nil
}

func (e *EnclaveStub) GenerateCCKeys() ([]byte, error) {
	panic("implement me")
}

func (e *EnclaveStub) ExportCCKeys(credentials []byte) ([]byte, error) {
	panic("implement me")
}

func (e *EnclaveStub) ImportCCKeys() ([]byte, error) {
	panic("implement me")
}

func (e *EnclaveStub) GetEnclaveId() (string, error) {
	panic("implement me")
}

// ChaincodeInvoke calls the enclave for transaction processing
func (e *EnclaveStub) ChaincodeInvoke(stub shim.ChaincodeStubInterface, crmProtoBytes []byte) ([]byte, error) {
	// Estimate of the buffer length where the enclave will write the response.
	const scresmProtoBytesMaxLen = 1024 * 100 // Let's be really conservative ...

	if !e.isInitialized {
		return nil, fmt.Errorf("enclave not yet initialized")
	}

	// register our stub for callbacks
	index := registry.register(&Stubs{stub})
	defer registry.release(index)
	ctx := unsafe.Pointer(&index)

	// prep signed proposal input
	proposal, err := stub.GetSignedProposal()
	if err != nil {
		return nil, fmt.Errorf("cannot get signed proposal: %s", err.Error())
	}
	signedProposalBytes, err := protoutil.Marshal(proposal)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal signed proposal: %s", err.Error())
	}
	signedProposalPtr := C.CBytes(signedProposalBytes)
	defer C.free(unsafe.Pointer(signedProposalPtr))

	// prep response
	scresmProtoBytesLenOut := C.uint32_t(0) // We pass maximal length separately; set to zero so we can detect valid responses
	scresmProtoBytesPtr := C.malloc(scresmProtoBytesMaxLen)
	defer C.free(scresmProtoBytesPtr)

	crmProtoBytesPtr := C.CBytes(crmProtoBytes)
	defer C.free(unsafe.Pointer(crmProtoBytesPtr))

	err = e.sem.Acquire(context.Background(), 1)
	if err != nil {
		return nil, err
	}

	// invoke enclave
	invokeRet := C.sgxcc_invoke(e.eid,
		(*C.uint8_t)(signedProposalPtr),
		(C.uint32_t)(len(signedProposalBytes)),
		(*C.uint8_t)(crmProtoBytesPtr),
		(C.uint32_t)(len(crmProtoBytes)),
		(*C.uint8_t)(scresmProtoBytesPtr), (C.uint32_t)(scresmProtoBytesMaxLen), &scresmProtoBytesLenOut,
		ctx)
	e.sem.Release(1)
	if invokeRet != 0 {
		return nil, fmt.Errorf("invoke failed. Reason: %d", int(invokeRet))
	}

	return C.GoBytes(scresmProtoBytesPtr, C.int(scresmProtoBytesLenOut)), nil
}
