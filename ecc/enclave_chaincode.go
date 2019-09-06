/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/hyperledger-labs/fabric-private-chaincode/ecc/crypto"
	"github.com/hyperledger-labs/fabric-private-chaincode/ecc/enclave"
	"github.com/hyperledger-labs/fabric-private-chaincode/ecc/ercc"
	"github.com/hyperledger-labs/fabric-private-chaincode/ecc/tlcc"
	"github.com/hyperledger-labs/fabric-private-chaincode/utils"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
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
	logger.Debug("NewEcc")
	return &EnclaveChaincode{
		erccStub: &ercc.EnclaveRegistryStubImpl{},
		tlccStub: &tlcc.TLCCStubImpl{},
		enclave:  enclave.NewEnclave(),
		verifier: &crypto.ECDSAVerifier{},
	}
}

// Init sets the chaincode state to "init"
func (t *EnclaveChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	mrenclave, err := t.enclave.MrEnclave()
	if err != nil {
		return shim.Error(err.Error())
	}
	logger.Debugf("Init: chaincode [mrenclave=%s]", mrenclave)
	if err := stub.PutState(utils.MrEnclaveStateKey, []byte(mrenclave)); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

// Invoke receives transactions and forwards to op handlers
func (t *EnclaveChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, _ := stub.GetFunctionAndParameters()
	logger.Debugf("Invoke is running [%s]", function)

	// Look first if there are system functions (and handle them)
	if function == "__setup" { // create enclave and register at ercc
		// Note: above string is also used in ecc_valdation_logic.go,
		// so if you change here you also will have to change there ...
		// If/when we refactor we should define such stuff somewhere as constants..
		return t.setup(stub)
	} else if function == "__init" { // pass CC init to chaincode
		return t.init(stub)
	} else if function == "__getEnclavePk" { //get Enclave PK
		return t.getEnclavePk(stub)
	}

	// Remaining functions are user functions, so pass them on the enclave and
	return t.invoke(stub)
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

	// we just add mrenclave to the proposal readset
	_, err = stub.GetState(utils.MrEnclaveStateKey)
	if err != nil {
		return shim.Error(err.Error())
	}

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
// init -
// ============================================================
func (t *EnclaveChaincode) init(stub shim.ChaincodeStubInterface) pb.Response {
	// check if we have an enclave already
	if t.enclave == nil {
		return shim.Error("ecc: Enclave not initialized! Run setup first!")
	}

	// get and json encode parameters
	_, argss := stub.GetFunctionAndParameters() // ignore function-name == __init
	jsonArgs, err := json.Marshal(argss)
	if err != nil {
		return shim.Error(err.Error())
	}

	// call enclave
	responseData, signature, err := t.enclave.Init(jsonArgs, stub, t.tlccStub)
	if err != nil {
		return shim.Error(fmt.Sprintf("ecc: Error while calling CC init enclave: %s", err))
	}

	enclavePk, err := t.enclave.GetPublicKey()
	if err != nil {
		return shim.Error(fmt.Sprintf("ecc: Error retrieving enclave pk: %s", err))
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
// invoke -
// ============================================================
func (t *EnclaveChaincode) invoke(stub shim.ChaincodeStubInterface) pb.Response {
	// check if we have an enclave already
	if t.enclave == nil {
		return shim.Error("ecc: Enclave not initialized! Run setup first!")
	}

	// get and json-encode parameters
	// Note: client side call of '{ "Args": [ arg1, arg2, .. ] }' and '{ "Function": "arg1", "Args": [ arg2, .. ] }' are identical ...
	argss := stub.GetStringArgs()
	jsonArgs, err := json.Marshal(argss)
	if err != nil {
		return shim.Error(err.Error())
	}

	pk := []byte(nil) // we don't really support a secure channel to the client yet ..
	// TODO: one of the place to fix when integrating end-to-end secure channel to client

	// call enclave
	responseData, signature, err := t.enclave.Invoke(jsonArgs, pk, stub, t.tlccStub)
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
