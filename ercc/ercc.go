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

package main

import (
	"strings"

	"crypto/sha256"
	"encoding/base64"
	"encoding/json"

	"github.com/hyperledger-labs/fabric-secure-chaincode/ercc/attestation"
	"github.com/hyperledger-labs/fabric-secure-chaincode/ercc/attestation/mock"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("ercc")

// EnclaveRegistryCC ...
type EnclaveRegistryCC struct {
	ra  attestation.Verifier
	ias attestation.IntelAttestationService
}

// NewErcc is a helpful factory method for creating this beauty
func NewErcc() *EnclaveRegistryCC {
	logger.Debug("NewErcc called")
	return &EnclaveRegistryCC{
		ra:  &attestation.VerifierImpl{},
		ias: attestation.NewIAS(),
	}
}

func NewTestErcc() *EnclaveRegistryCC {
	return &EnclaveRegistryCC{
		ra:  &mock.MockVerifier{},
		ias: &mock.MockIAS{},
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
	isValid, err := ercc.ra.VerifyAttestionReport(verificationPK, attestationReport)
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

func main() {
	// start chaincode
	// err := shim.Start(NewTestErcc())
	err := shim.Start(NewErcc())
	if err != nil {
		logger.Errorf("Error starting registry chaincode: %s", err)
	}
}
