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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"unsafe"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger-labs/fabric-private-chaincode/ecc/chaincode/crypto"
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

// invoke calls the enclave for transaction processing, takes arguments
// and the current chaincode state as input and returns a new chaincode state
//func (e *StubImpl) Invoke(shimStub shim.ChaincodeStubInterface) ([]byte, error) {
//	var err error
//
//	if shimStub == nil {
//		return nil, errors.New("Need shim")
//	}
//
//	index := registry.Register(&Stubs{shimStub})
//	defer registry.Release(index)
//	ctx := unsafe.Pointer(&index)
//
//	// response
//	cresmProtoBytesLenOut := C.uint32_t(0) // We pass maximal length separatedly; set to zero so we can detect valid responses
//	cresmProtoBytesPtr := C.malloc(MAX_RESPONSE_SIZE)
//	defer C.free(cresmProtoBytesPtr)
//
//	// get signed proposal
//	signedProposal, err := shimStub.GetSignedProposal()
//	if err != nil {
//		return nil, fmt.Errorf("cannot get signed proposal")
//	}
//	signedProposalBytes, err := proto.Marshal(signedProposal)
//	if err != nil {
//		return nil, fmt.Errorf("cannot get signed proposal bytes")
//	}
//	signedProposalPtr := C.CBytes(signedProposalBytes)
//	defer C.free(unsafe.Pointer(signedProposalPtr))
//
//	//ASSUME HERE input is not the protobuf, so let's buildit (rmeove block later)
//	argss := shimStub.GetStringArgs()
//	argsByteArray := make([][]byte, len(argss))
//	for i, v := range argss {
//		argsByteArray[i] = []byte(v)
//		logger.Debugf("arg %d: %s", i, argsByteArray[i])
//	}
//	cleartextChaincodeRequestMessageProto := &fpcpb.CleartextChaincodeRequest{
//		Input: &peer.ChaincodeInput{Args: argsByteArray},
//	}
//	cleartextChaincodeRequestMessageProtoBytes, err := proto.Marshal(cleartextChaincodeRequestMessageProto)
//	if err != nil {
//		return nil, fmt.Errorf("marshal error")
//	}
//	crmProto := &fpcpb.ChaincodeRequestMessage{
//		// TODO: eventually this should be an encrypted CleartextRequestMessage
//		EncryptedRequest: cleartextChaincodeRequestMessageProtoBytes,
//	}
//	crmProtoBytes, err := proto.Marshal(crmProto)
//	if err != nil {
//		return nil, fmt.Errorf("marshal error")
//	}
//	crmProtoBytesPtr := C.CBytes(crmProtoBytes)
//	defer C.free(unsafe.Pointer(crmProtoBytesPtr))
//	//REMOVE BLOCK ABOVE once protobuf supported e2e
//
//	e.sem.Acquire(context.Background(), 1)
//	// invoke enclave
//	invoke_ret := C.sgxcc_invoke(e.eid,
//		(*C.uint8_t)(signedProposalPtr),
//		(C.uint32_t)(len(signedProposalBytes)),
//		(*C.uint8_t)(crmProtoBytesPtr),
//		(C.uint32_t)(len(crmProtoBytes)),
//		(*C.uint8_t)(cresmProtoBytesPtr), (C.uint32_t)(MAX_RESPONSE_SIZE), &cresmProtoBytesLenOut,
//		ctx)
//	e.sem.Release(1)
//	if invoke_ret != 0 {
//		return nil, fmt.Errorf("Invoke failed. Reason: %d", int(invoke_ret))
//	}
//	cresmProtoBytes := C.GoBytes(cresmProtoBytesPtr, C.int(cresmProtoBytesLenOut))
//
//	//ASSUME HERE we get the b64 encoded response protobuf, pull encrypted response out and return it
//	cresmProto := &fpcpb.ChaincodeResponseMessage{}
//	err = proto.Unmarshal(cresmProtoBytes, cresmProto)
//	if err != nil {
//		return nil, fmt.Errorf("unmarshal error")
//	}
//
//	// TODO: this should be eventually be an (encrypted) fabric Response object rather than the response string ...
//	return cresmProto.EncryptedResponse, nil
//}

func (e *EnclaveStub) ChaincodeInvoke(stub shim.ChaincodeStubInterface) ([]byte, error) {
	if !e.isInitialized {
		return nil, fmt.Errorf("enclave not yet initialized")
	}

	// register our stub for callbacks
	index := registry.register(&Stubs{stub})
	defer registry.release(index)
	ctx := unsafe.Pointer(&index)

	proposal, err := stub.GetSignedProposal()
	if err != nil {
		return nil, err
	}

	// TODO give this signed proposal to enclave
	serializedProposal, err := proto.Marshal(proposal)
	if err != nil {
		return nil, err
	}
	_ = serializedProposal

	// args
	jsonArgs, err := json.Marshal(stub.GetStringArgs())
	if err != nil {
		return nil, err
	}
	argsPtr := C.CString(string(jsonArgs))
	defer C.free(unsafe.Pointer(argsPtr))

	// TODO to be removed! not used
	// client pk used for args encryption
	pkPtr := C.CString(string([]byte(nil)))
	defer C.free(unsafe.Pointer(pkPtr))

	// response
	responseLenOut := C.uint32_t(0) // We pass maximal length separatedly; set to zero so we can detect valid responses
	responsePtr := C.malloc(maxResponseSize)
	defer C.free(responsePtr)

	// signature
	const SignatureSize = 64
	signaturePtr := C.malloc(SignatureSize)
	defer C.free(signaturePtr)

	// TODO
	// - call enclave with serialzed proposal as argument
	// - get response back

	e.sem.Acquire(context.Background(), 1)
	// invoke enclave
	invoke_ret := C.sgxcc_invoke(e.eid,
		argsPtr,
		pkPtr,
		(*C.uint8_t)(responsePtr), C.uint32_t(maxResponseSize), &responseLenOut,
		(*C.ec256_signature_t)(signaturePtr),
		ctx)
	e.sem.Release(1)

	// Note: we do try to return the response in all cases, even then there is an error ...
	var sig []byte
	if invoke_ret == 0 {
		sig, err = crypto.MarshalEnclaveSignature(C.GoBytes(signaturePtr, C.int(SignatureSize)))
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

	// pubkey
	const PUB_KEY_SIZE = 64
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
	pk, err := crypto.MarshalEnclavePk(C.GoBytes(pubkeyPtr, C.int(PUB_KEY_SIZE)))
	if err != nil {
		return nil, err
	}
	hashedPk := sha256.Sum256(pk)
	enclaveId := hex.EncodeToString(hashedPk[:])

	response := &protos.ChaincodeResponseMessage{}
	//err = proto.Unmarshal(enclaveResponseBytes, response)
	//if err != nil {
	//	return nil, err
	//}

	response.RwSet = nil // TODO set rw set to response
	response.Proposal = proposal
	response.Signature = sig
	response.EnclaveId = enclaveId
	response.EncryptedResponse = C.GoBytes(responsePtr, C.int(responseLenOut))

	// serialized and return updated response
	return proto.Marshal(response)
}
