/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corp.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/hyperledger-labs/fabric-private-chaincode/ercc/attestation"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/flogging"
)

// #cgo CFLAGS: -I/opt/intel/sgxsdk/include -I${SRCDIR}/../common/crypto
// #cgo LDFLAGS: -L${SRCDIR}/../common/_build -L${SRCDIR}/../common/crypto/_build -Wl,--start-group -lpdo-utils -lupdo-crypto -Wl,--end-group -lcrypto
// #include "stdlib.h"  /* needed for free */
// #include "pdo/common/crypto/verify_ias_report/verify-report.h"
import "C"
import "unsafe"

var logger = flogging.MustGetLogger("ercc")

func main() {
	// start chaincode
	// err := shim.Start(NewTestErcc())
	err := shim.Start(NewErcc())
	if err != nil {
		logger.Errorf("Error starting registry chaincode: %s", err)
	}
}

// EnclaveRegistryCC ...
type EnclaveRegistryCC struct {
	ra  attestation.Verifier
	ias attestation.IntelAttestationService
}

// NewErcc is a helpful factory method for creating this beauty
func NewErcc() *EnclaveRegistryCC {
	logger.Debug("NewErcc called")
	return &EnclaveRegistryCC{
		ra:  attestation.GetVerifier(),
		ias: attestation.GetIAS(),
	}
}

func NewTestErcc() *EnclaveRegistryCC {
	return &EnclaveRegistryCC{
		ra:  &attestation.MockVerifier{},
		ias: &attestation.MockIAS{},
	}
}

// Init setups the EnclaveRegistry by initializing intel verification key. Currently this is hardcoded!
func (ercc *EnclaveRegistryCC) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Init called")
	return shim.Success(nil)
}

// Invoke receives transactions and forwards to op handlers
func (ercc *EnclaveRegistryCC) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	logger.Debugf("Invoke(function=%s, %s) called", function, args)

	if function == "registerEnclave" {
		return ercc.registerEnclave(stub, args)
	} else if function == "getAttestationReport" { //get enclave attestation report
		return ercc.getAttestationReport(stub, args)
	} else if function == "getSPID" { //get SPID
		return ercc.getSPID(stub, args)
	}

	return shim.Error("Received unknown function invocation: " + function)
}

// ============================================================
// registerEnclave -
// ============================================================
func (ercc *EnclaveRegistryCC) registerEnclave(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// args:
	// 0: enclavePkBase64
	// 1: quoteBase64
	// 2: apiKey
	// if apiKey not available as argument we try to read them from decorator

	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting enclave pk and quote to register")
	}

    //TO BE REMOVED
    pCert1 := C.CString("mock certificate")
    defer C.free(unsafe.Pointer(pCert1))
    ret := C.verify_ias_certificate_chain(pCert1)
    if ret != 0 {
        logger.Debugf("call to pdo crypto failed as expected")
    }

	enclavePkAsBytes, err := base64.StdEncoding.DecodeString(args[0])
	if err != nil {
		return shim.Error("Can not parse enclavePkHash: " + err.Error())
	}

	quoteBase64 := args[1]
	quoteAsBytes, err := base64.StdEncoding.DecodeString(quoteBase64)
	if err != nil {
		return shim.Error("Can not parse quoteBase64 string: " + err.Error())
	}

	// get ercc api key for IAS
	var apiKey string
	if len(args) >= 3 {
		apiKey = args[2]
	} else {
		apiKey = string(stub.GetDecorations()["apiKey"])
	}
	apiKey = strings.TrimSpace(apiKey) // make sure there are no trailing newlines and alike ..
	logger.Debugf("registerEnclave: api-key: %s / len(args)=%d", apiKey, len(args))

	// send quote to intel for verification
	attestationReport, err := ercc.ias.RequestAttestationReport(apiKey, quoteAsBytes)
	if err != nil {
		return shim.Error("Error while retrieving attestation report: " + err.Error())
	}

	// TODO get verification public key from ledger
	verificationPK, err := ercc.ias.GetIntelVerificationKey()
	if err != nil {
		return shim.Error("Can not parse verifiaction key: " + err.Error())
	}

	// verify attestation report
	isValid, err := ercc.ra.VerifyAttestationReport(verificationPK, attestationReport)
	if err != nil {
		return shim.Error("Error while attestation report verification: " + err.Error())
	}
	if !isValid {
		return shim.Error("Attestation report is not valid")
	}

	// first verify that enclavePkHash matches the one in the attestation report
	isValid, err = ercc.ra.CheckEnclavePkHash(enclavePkAsBytes, attestationReport)
	if err != nil {
		return shim.Error("Error while checking enclave PK: " + err.Error())
	}
	if !isValid {
		return shim.Error("Enclave PK does not match attestation report!")
	}
	// set enclave public key in attestation report
	attestationReport.EnclavePk = enclavePkAsBytes

	// store attestation report under enclavePk hash in state
	attestationReportAsBytes, err := json.Marshal(attestationReport)
	if err != nil {
		return shim.Error(err.Error())
	}

	// create hash of enclave pk
	enclavePkHash := sha256.Sum256(enclavePkAsBytes)
	enclavePkHashBase64 := base64.StdEncoding.EncodeToString(enclavePkHash[:])
	err = stub.PutState(enclavePkHashBase64, attestationReportAsBytes)

	return shim.Success(nil)
}

// ============================================================
// getAttestationReport -
// ============================================================
func (ercc *EnclaveRegistryCC) getAttestationReport(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	//   0
	// "enclavePkHashBase64"
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting pk of the enclave to query")
	}

	enclavePkHashBase64 := args[0]
	attestationReport, err := stub.GetState(enclavePkHashBase64) //get attestationREPORT for enclavePK from chaincode state
	if err != nil {
		return shim.Error("Failed to get state for " + enclavePkHashBase64)
	} else if attestationReport == nil {
		return shim.Error("EnclavePK does not exist: " + enclavePkHashBase64)
	}

	return shim.Success(attestationReport)
}

// ============================================================
// getSPID -
// ============================================================
func (ercc *EnclaveRegistryCC) getSPID(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// return spid from IASCredentialProvider
	return shim.Success(stub.GetDecorations()["SPID"])
	// return shim.Success(ercc.iascp.GetSPID())
}
