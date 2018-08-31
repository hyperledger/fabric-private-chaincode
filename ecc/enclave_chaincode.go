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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/hyperledger-labs/fabric-secure-chaincode/ecc/crypto"
	"github.com/hyperledger-labs/fabric-secure-chaincode/ecc/enclave"
	"github.com/hyperledger-labs/fabric-secure-chaincode/ecc/ercc"
	"github.com/hyperledger-labs/fabric-secure-chaincode/ecc/tlcc"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger-labs/fabric-secure-chaincode/utils"
)

const enclaveLibFile = "enclave/lib/enclave.signed.so"

var logger = shim.NewLogger("ecc")

// EnclaveChaincode struct
type EnclaveChaincode struct {
	erccStub ercc.EnclaveRegistryStub
	tlccStub tlcc.TLCCStub
	enclave  enclave.Stub
	verifier crypto.Verifier
}

// NewEcc is a helpful factory method for creating this beauty
func NewEcc() *EnclaveChaincode {
	return &EnclaveChaincode{
		erccStub: &ercc.EnclaveRegistryStubImpl{},
		tlccStub: &tlcc.TLCCStubImpl{},
		enclave:  enclave.NewEnclave(),
		verifier: &crypto.ECDSAVerifier{},
	}
}

// Init sets the chaincode state to "init"
func (t *EnclaveChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debugf("ecc: Init chaincode [%s]", enclave.MrEnclave)
	if err := stub.PutState(utils.MrEnclaveStateKey, []byte(enclave.MrEnclave)); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

// Invoke receives transactions and forwards to op handlers
func (t *EnclaveChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, _ := stub.GetFunctionAndParameters()
	logger.Debugf("ecc: invoke is running [%s]", function)

	if function == "setup" { // create enclave and register at ercc
		return t.setup(stub)
	} else if function == "getEnclavePk" { //get Enclave PK
		return t.getEnclavePk(stub)
	} else {
		return t.invoke(stub)
	}
}

// ============================================================
// setup -
// ============================================================
func (t *EnclaveChaincode) setup(stub shim.ChaincodeStubInterface) pb.Response {
	// TODO check that args are valid
	args := stub.GetStringArgs()
	erccName := args[1]
	channelName := stub.GetChannelID()

	// check if there is already an enclave
	if t.enclave == nil {
		return shim.Error("ecc: Enclave has already been initialized! Destroy first!!")
	}

	// create new Enclave
	// TODO we should return error in case there is any :)
	if err := t.enclave.Create(enclaveLibFile); err != nil {
		return shim.Error(fmt.Sprintf("ecc: Error while creating enclave %s", err))
	}

	//get spid from ercc
	spid, err := t.erccStub.GetSPID(stub, erccName, channelName)
	if err != nil {
		return shim.Error(err.Error())
	}
	logger.Debugf("ecc: SPID from ercc: %x", spid)

	// ask enclave for quote
	quoteAsBytes, enclavePk, err := t.enclave.GetRemoteAttestationReport(spid)
	if err != nil {
		return shim.Error(fmt.Sprintf("ecc: Error while creating attestation report: %s", err))
	}

	enclavePkBase64 := base64.StdEncoding.EncodeToString(enclavePk)
	quoteBase64 := base64.StdEncoding.EncodeToString(quoteAsBytes)

	// register enclave at ercc
	if err = t.erccStub.RegisterEnclave(stub, erccName, channelName, []byte(enclavePkBase64), []byte(quoteBase64)); err != nil {
		return shim.Error(err.Error())
	}

	logger.Debugf("ecc: registration done; next binding")
	// get target info from our new enclave
	eccTargetInfo, err := t.enclave.GetTargetInfo()
	if err != nil {
		return shim.Error(fmt.Sprintf("Error while getting target info: %s", err))
	}

	// get report and pk from tlcc using target info from ecc enclave
	tlccReport, tlccPk, err := t.tlccStub.GetReport(stub, "tlcc", channelName, eccTargetInfo)
	if err != nil {
		return shim.Error(err.Error())
	}

	// call enclave binding
	if err = t.enclave.Bind(tlccReport, tlccPk); err != nil {
		return shim.Error(fmt.Sprintf("Error while binding: %s", err))
	}

	return shim.Success([]byte(enclavePkBase64))
}

// ============================================================
// invoke -
// ============================================================
func (t *EnclaveChaincode) invoke(stub shim.ChaincodeStubInterface) pb.Response {
	// check if we have an enclave already
	if t.enclave == nil {
		return shim.Error("ecc: Enclave not initialized! Run setup first!")
	}
	argss := stub.GetStringArgs()
	args := []byte(argss[0])
	pk := []byte(argss[1])

	// call enclave
	responseData, signature, err := t.enclave.Invoke(args, pk, stub, t.tlccStub)
	if err != nil {
		return shim.Error(fmt.Sprintf("ecc: Error while invoking enclave: %s", err))
	}

	enclavePk, err := t.enclave.GetPublicKey()
	if err != nil {
		return shim.Error(fmt.Sprintf("ecc: Error while retrieving enclave pk: %s", err))
	}

	response := &utils.Response{
		ResponseData: responseData,
		Signature:    signature,
		PublicKey:    enclavePk,
	}
	responseBytes, _ := json.Marshal(response)

	return shim.Success(responseBytes)
}

// ============================================================
// getEnclavePk -
// ============================================================
func (t *EnclaveChaincode) getEnclavePk(stub shim.ChaincodeStubInterface) pb.Response {
	// check if we have an enclave already
	if t.enclave == nil {
		return shim.Error("ecc: Enclave not initialized! Run setup first!")
	}

	// get enclaves public key
	enclavePk, err := t.enclave.GetPublicKey()
	if err != nil {
		return shim.Error(fmt.Sprintf("ecc: Error while retrieving enclave pk %s", err))
	}

	// marshal response
	responseBytes, _ := json.Marshal(&utils.Response{PublicKey: enclavePk})
	return shim.Success(responseBytes)
}

func (t *EnclaveChaincode) destroy() {
	if err := t.enclave.Destroy(); err != nil {
		panic("ecc: Can not destory enclave!!!")
	}
}

func main() {
	// create enclave chaincode
	t := NewEcc()
	defer t.destroy()

	// start chaincode
	if err := shim.Start(t); err != nil {
		logger.Errorf("Error starting ecc: %s", err)
	}
}
