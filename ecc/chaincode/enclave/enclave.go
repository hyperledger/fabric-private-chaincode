// +build !mock_ecc

/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package enclave

import "C"
import (
	"context"
	"encoding/base64"
	"fmt"
	"unsafe"

	"github.com/golang/protobuf/proto"
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
	eid           C.enclave_id_t
	sem           *semaphore.Weighted
	isInitialized bool
}

// NewEnclave starts a new enclave
func NewEnclaveStub() StubInterface {
	return &EnclaveStub{sem: semaphore.NewWeighted(8)}
}

func (e *EnclaveStub) Init(chaincodeParams, hostParams, attestationParams []byte) ([]byte, error) {
	if e.isInitialized {
		return nil, fmt.Errorf("enclave already initialized")
	}

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

	e.isInitialized = true

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

// ChaincodeInvoke calls the enclave for transaction processing
func (e *EnclaveStub) ChaincodeInvoke(stub shim.ChaincodeStubInterface) ([]byte, error) {
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
	signedProposalBytes, err := proto.Marshal(proposal)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal signed proposal: %s", err.Error())
	}
	signedProposalPtr := C.CBytes(signedProposalBytes)
	defer C.free(unsafe.Pointer(signedProposalPtr))

	// prep response
	cresmProtoBytesLenOut := C.uint32_t(0) // We pass maximal length separately; set to zero so we can detect valid responses
	cresmProtoBytesPtr := C.malloc(maxResponseSize)
	defer C.free(cresmProtoBytesPtr)

	// prep chaincode request message as input
	_, args := stub.GetFunctionAndParameters()
	if len(args) != 1 {
		return nil, fmt.Errorf("no chaincodeRequestMessage as argument found")
	}
	crmProtoBytes, err := base64.StdEncoding.DecodeString(args[0])
	if err != nil {
		return nil, fmt.Errorf("cannot decode ChaincodeRequestMessage: %s", err.Error())
	}
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
		(*C.uint8_t)(cresmProtoBytesPtr), (C.uint32_t)(maxResponseSize), &cresmProtoBytesLenOut,
		ctx)
	e.sem.Release(1)
	if invokeRet != 0 {
		return nil, fmt.Errorf("invoke failed. Reason: %d", int(invokeRet))
	}
	cresmProtoBytes := C.GoBytes(cresmProtoBytesPtr, C.int(cresmProtoBytesLenOut))

	//ASSUME HERE we get the b64 encoded response protobuf, pull encrypted response out and return it
	cresmProto := &protos.ChaincodeResponseMessage{}
	err = proto.Unmarshal(cresmProtoBytes, cresmProto)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal ChaincodeResponseMessage: %s", err.Error())
	}

	// TODO: this should be eventually be an (encrypted) fabric Response object rather than the response string ...
	return cresmProto.EncryptedResponse, nil
}
