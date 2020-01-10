/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package enclave

import (
	"fmt"
	"unsafe"

	"github.com/hyperledger-labs/fabric-private-chaincode/ecc/crypto"
	"github.com/hyperledger/fabric/common/flogging"
)

// #cgo CFLAGS: -I${SRCDIR}/include -I${SRCDIR}/../../common/sgxcclib
// #cgo LDFLAGS: -L${SRCDIR}/lib -ltl
// #include "common-sgxcclib.h"
// #include <trusted_ledger.h>
import "C"

const EPID_SIZE = 8
const SPID_SIZE = 16
const SIGNATURE_SIZE = 64
const PUB_KEY_SIZE = 64
const REPORT_SIZE = 432
const TARGET_INFO_SIZE = 512
const CMAC_SIZE = 16

var logger = flogging.MustGetLogger("tl-enclave")

//export golog
func golog(str *C.char) {
	logger.Infof("%s", C.GoString(str))
}

// Stub interface
type Stub interface {
	GetTargetInfo() ([]byte, error)
	// Return report and enclave PK in DER-encoded PKIX format
	GetLocalAttestationReport(targetInfo []byte) ([]byte, []byte, error)
	// Creates an enclave from a given enclave lib file
	Create(enclaveLibFile string) error
	// Init enclave with a given genesis block
	InitWithGenesis(blockBytes []byte) error
	// give enclave next block to validate and append to the ledger
	NextBlock(blockBytes []byte) error
	// verifies state and returns cmac
	GetStateMetadata(key string, nonce []byte, isRangeQuery bool) ([]byte, error)
	// Destroys enclave
	Destroy() error
}

// StubImpl implements the interface
type StubImpl struct {
	eid C.enclave_id_t
}

// NewEnclave starts a new enclave
func NewEnclave() Stub {
	return &StubImpl{}
}

func (e *StubImpl) GetTargetInfo() ([]byte, error) {
	// TODO what is the correct target info size
	targetInfo := make([]byte, TARGET_INFO_SIZE)
	targetInfoPtr := C.CBytes(targetInfo)
	defer C.free(targetInfoPtr)

	ret := C.sgxcc_get_target_info(e.eid, (*C.target_info_t)(targetInfoPtr))
	if ret != 0 {
		return nil, fmt.Errorf("C.sgxcc_get_target_info failed. Reason: %d", int(ret))
	}

	return targetInfo, nil
}

func (e *StubImpl) GetLocalAttestationReport(targetInfo []byte) ([]byte, []byte, error) {

	// report
	report := make([]byte, REPORT_SIZE)
	reportPtr := C.CBytes(report)
	defer C.free(reportPtr)

	// pubkey
	pubkey := make([]byte, PUB_KEY_SIZE)
	pubkeyPtr := C.CBytes(pubkey)
	defer C.free(pubkeyPtr)

	// targetInfo
	targetInfoPtr := C.CBytes(targetInfo)
	defer C.free(targetInfoPtr)

	// call enclave
	ret := C.sgxcc_get_local_attestation_report(e.eid,
		(*C.target_info_t)(targetInfoPtr),
		(*C.report_t)(reportPtr),
		(*C.ec256_public_t)(pubkeyPtr))
	if ret != 0 { // 0 is SGX_SUCCESS
		return nil, nil, fmt.Errorf("C.sgxcc_get_local_attestation_report failed. Reason: %d", int(ret))
	}

	// convert sgx format to DER-encoded PKIX format
	pk, err := crypto.MarshalEnclavePk(C.GoBytes(pubkeyPtr, C.int(PUB_KEY_SIZE)))
	if err != nil {
		return nil, nil, err
	}

	return C.GoBytes(reportPtr, C.int(REPORT_SIZE)), pk, nil
}

func (e *StubImpl) InitWithGenesis(blockBytes []byte) error {
	blockBytesPtr := C.CBytes(blockBytes)
	blockBytesLen := len(blockBytes)
	defer C.free(blockBytesPtr)

	ret := C.tlcc_init_with_genesis(e.eid,
		(*C.uint8_t)(blockBytesPtr), C.uint32_t(blockBytesLen))
	if ret != 0 { // 0 is SGX_SUCCESS
		return fmt.Errorf("C.tlcc_init_with_genesis failed. Reason: %d", int(ret))
	}

	return nil
}

func (e *StubImpl) NextBlock(block []byte) error {
	blockLen := len(block)
	blockPtr := C.CBytes(block)
	defer C.free(blockPtr)

	_, err := C.tlcc_send_block(e.eid,
		(*C.uint8_t)(blockPtr), C.uint32_t(blockLen))

	if err != nil {
		return err
	}

	return nil
}

func (e *StubImpl) GetStateMetadata(key string, nonce []byte, isRangeQuery bool) ([]byte, error) {
	// key
	keyc := C.CString(key)
	defer C.free(unsafe.Pointer(keyc))

	// nonce
	noncePtr := C.CBytes(nonce)
	defer C.free(noncePtr)

	// cmac
	cmac := make([]byte, CMAC_SIZE)
	cmacPtr := C.CBytes(cmac)
	defer C.free(cmacPtr)

	if isRangeQuery {
		ret := C.tlcc_get_multi_state_metadata(e.eid, keyc,
			(*C.uint8_t)(noncePtr),
			(*C.cmac_t)(cmacPtr))
		if ret != 0 { // 0 is SGX_SUCCESS
			return nil, fmt.Errorf("C.tlcc_get_multi_state_metadata failed. Reason: %d", int(ret))
		}
	} else {
		ret := C.tlcc_get_state_metadata(e.eid, keyc,
			(*C.uint8_t)(noncePtr),
			(*C.cmac_t)(cmacPtr))
		if ret != 0 { // 0 is SGX_SUCCESS
			return nil, fmt.Errorf("C.tlcc_get_state_metadata failed. Reason: %d", int(ret))
		}
	}
	return C.GoBytes(cmacPtr, C.int(CMAC_SIZE)), nil
}

// Create starts a new enclave instance
func (e *StubImpl) Create(enclaveLibFile string) error {
	var eid C.enclave_id_t

	f := C.CString(enclaveLibFile)
	defer C.free(unsafe.Pointer(f))

	// todo read error
	ret := C.sgxcc_create_enclave(&eid, f)
	if ret != 0 {
		return fmt.Errorf("C.sgxcc_create_enclave (lib %s) failed. Reason: %d", enclaveLibFile, ret)
	}

	e.eid = eid
	return nil
}

// Destroy kills the current enclave instance
func (e *StubImpl) Destroy() error {
	// todo read error
	ret := C.sgxcc_destroy_enclave(e.eid)
	if ret != 0 {
		return fmt.Errorf("C.sgxcc_destroy_enclave failed. Reason: %d", int(ret))
	}

	return nil
}
