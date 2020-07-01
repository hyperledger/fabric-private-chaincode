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
    pCert1 := C.CString("-----BEGIN CERTIFICATE-----\nMIIEoTCCAwmgAwIBAgIJANEHdl0yo7CWMA0GCSqGSIb3DQEBCwUAMH4xCzAJBgNV\nBAYTAlVTMQswCQYDVQQIDAJDQTEUMBIGA1UEBwwLU2FudGEgQ2xhcmExGjAYBgNV\nBAoMEUludGVsIENvcnBvcmF0aW9uMTAwLgYDVQQDDCdJbnRlbCBTR1ggQXR0ZXN0\nYXRpb24gUmVwb3J0IFNpZ25pbmcgQ0EwHhcNMTYxMTIyMDkzNjU4WhcNMjYxMTIw\nMDkzNjU4WjB7MQswCQYDVQQGEwJVUzELMAkGA1UECAwCQ0ExFDASBgNVBAcMC1Nh\nbnRhIENsYXJhMRowGAYDVQQKDBFJbnRlbCBDb3Jwb3JhdGlvbjEtMCsGA1UEAwwk\nSW50ZWwgU0dYIEF0dGVzdGF0aW9uIFJlcG9ydCBTaWduaW5nMIIBIjANBgkqhkiG\n9w0BAQEFAAOCAQ8AMIIBCgKCAQEAqXot4OZuphR8nudFrAFiaGxxkgma/Es/BA+t\nbeCTUR106AL1ENcWA4FX3K+E9BBL0/7X5rj5nIgX/R/1ubhkKWw9gfqPG3KeAtId\ncv/uTO1yXv50vqaPvE1CRChvzdS/ZEBqQ5oVvLTPZ3VEicQjlytKgN9cLnxbwtuv\nLUK7eyRPfJW/ksddOzP8VBBniolYnRCD2jrMRZ8nBM2ZWYwnXnwYeOAHV+W9tOhA\nImwRwKF/95yAsVwd21ryHMJBcGH70qLagZ7Ttyt++qO/6+KAXJuKwZqjRlEtSEz8\ngZQeFfVYgcwSfo96oSMAzVr7V0L6HSDLRnpb6xxmbPdqNol4tQIDAQABo4GkMIGh\nMB8GA1UdIwQYMBaAFHhDe3amfrzQr35CN+s1fDuHAVE8MA4GA1UdDwEB/wQEAwIG\nwDAMBgNVHRMBAf8EAjAAMGAGA1UdHwRZMFcwVaBToFGGT2h0dHA6Ly90cnVzdGVk\nc2VydmljZXMuaW50ZWwuY29tL2NvbnRlbnQvQ1JML1NHWC9BdHRlc3RhdGlvblJl\ncG9ydFNpZ25pbmdDQS5jcmwwDQYJKoZIhvcNAQELBQADggGBAGcIthtcK9IVRz4r\nRq+ZKE+7k50/OxUsmW8aavOzKb0iCx07YQ9rzi5nU73tME2yGRLzhSViFs/LpFa9\nlpQL6JL1aQwmDR74TxYGBAIi5f4I5TJoCCEqRHz91kpG6Uvyn2tLmnIdJbPE4vYv\nWLrtXXfFBSSPD4Afn7+3/XUggAlc7oCTizOfbbtOFlYA4g5KcYgS1J2ZAeMQqbUd\nZseZCcaZZZn65tdqee8UXZlDvx0+NdO0LR+5pFy+juM0wWbu59MvzcmTXbjsi7HY\n6zd53Yq5K244fwFHRQ8eOB0IWB+4PfM7FeAApZvlfqlKOlLcZL2uyVmzRkyR5yW7\n2uo9mehX44CiPJ2fse9Y6eQtcfEhMPkmHXI01sN+KwPbpA39+xOsStjhP9N1Y1a2\ntQAVo+yVgLgV2Hws73Fc0o3wC78qPEA+v2aRs/Be3ZFDgDyghc/1fgU+7C+P6kbq\nd4poyb6IW8KCJbxfMJvkordNOgOUUxndPHEi/tb/U7uLjLOgPA==\n-----END CERTIFICATE-----")
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
