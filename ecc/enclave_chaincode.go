/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ecc

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

func CreateMockedECC() *EnclaveChaincode {
	return &EnclaveChaincode{
		erccStub: &ercc.MockEnclaveRegistryStub{},
		tlccStub: &tlcc.MockTLCCStub{},
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
	//TODO: pass sigrl via args
	sigRl := []byte(nil)
	sigRlSize := uint(0)
	channelName := stub.GetChannelID()

	// check if there is already an enclave
	if t.enclave == nil {
		return shim.Error("ecc: Enclave has already been initialized! Destroy first!!")
	}

	// create new Enclave
	// TODO we should return error in case there is any :)
	if err := t.enclave.Create(enclaveLibFile); err != nil {
		errMsg := fmt.Sprintf("t.enclave.Create  failed: %s", err)
		logger.Errorf(errMsg)
		return shim.Error(errMsg)
	}

	//get spid from ercc
	spid, err := t.erccStub.GetSPID(stub, erccName, channelName)
	if err != nil {
		return shim.Error(err.Error())
	}
	logger.Debugf("ecc: SPID from ercc: %x", spid)

	// ask enclave for quote
	quoteAsBytes, enclavePk, err := t.enclave.GetRemoteAttestationReport(spid, sigRl, sigRlSize)
	if err != nil {
		errMsg := fmt.Sprintf("t.enclave.GetRemoteAttestationReport failed: %s", err)
		logger.Errorf(errMsg)
		return shim.Error(errMsg)
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
		errMsg := fmt.Sprintf("t.enclave.GetTargetInfo failed: %s", err)
		logger.Errorf(errMsg)
		return shim.Error(errMsg)
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
	var errMsg string
	responseData, signature, errInit := t.enclave.Init(jsonArgs, stub, t.tlccStub)
	if errInit != nil {
		errMsg = fmt.Sprintf("t.enclave.Init failed: %s", errInit)
		logger.Errorf(errMsg)
		// likely a chaincode error, so we stil want response go back ...
	}

	enclavePk, errPk := t.enclave.GetPublicKey()
	if errPk != nil {
		errMsg = fmt.Sprintf("init t.enclave.GetPublicKey failed. Reason: %s", err)
		logger.Errorf(errMsg)
		// return (and ignore any potential response) as this is a more systematic error
		return shim.Error(errMsg)
	}

	fpcResponse := &utils.Response{
		ResponseData: responseData,
		Signature:    signature,
		PublicKey:    enclavePk,
	}
	responseBytes, _ := json.Marshal(fpcResponse)

	var response pb.Response
	if errInit == nil {
		response = pb.Response{
			Status:  shim.OK,
			Payload: responseBytes,
			Message: errMsg,
		}
	} else {
		response = pb.Response{
			Status:  shim.ERROR,
			Payload: responseBytes,
			Message: errMsg,
		}
	}
	logger.Debugf("init response: %v", response)
	return response
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
	var errMsg string
	responseData, signature, errInvoke := t.enclave.Invoke(jsonArgs, pk, stub, t.tlccStub)
	if errInvoke != nil {
		errMsg = fmt.Sprintf("t.enclave.Invoke failed: %s", errInvoke)
		logger.Errorf(errMsg)
		// likely a chaincode error, so we stil want response go back ...
	}

	enclavePk, errPk := t.enclave.GetPublicKey()
	if errPk != nil {
		errMsg = fmt.Sprintf("invoke t.enclave.GetPublicKey failed. Reason: %s", err)
		logger.Errorf(errMsg)
		// return (and ignore any potential response) as this is a more systematic error
		return shim.Error(errMsg)
	}

	fpcResponse := &utils.Response{
		ResponseData: responseData,
		Signature:    signature,
		PublicKey:    enclavePk,
	}
	responseBytes, _ := json.Marshal(fpcResponse)

	var response pb.Response
	if errInvoke == nil {
		response = pb.Response{
			Status:  shim.OK,
			Payload: responseBytes,
			Message: errMsg,
		}
	} else {
		response = pb.Response{
			Status:  shim.ERROR,
			Payload: responseBytes,
			Message: errMsg,
		}
	}
	logger.Debugf("invoke response: %v", response)
	return response
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
		errMsg := fmt.Sprintf("getEnclavePk t.enclave.GetPublicKey failed. Reason: %s", err)
		logger.Errorf(errMsg)
		return shim.Error(errMsg)
	}

	// marshal response
	responseBytes, _ := json.Marshal(&utils.Response{PublicKey: enclavePk})
	return shim.Success(responseBytes)
}

// TODO: check if Destroy is called
func (t *EnclaveChaincode) Destroy() {
	if err := t.enclave.Destroy(); err != nil {
		errMsg := fmt.Sprintf("t.enclave.Destroy failed: %s", err)
		logger.Errorf(errMsg)
		panic(errMsg)
	}
}
