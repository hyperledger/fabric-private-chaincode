/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"encoding/base64"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-private-chaincode/ecc/chaincode/ercc"
	"github.com/hyperledger/fabric-private-chaincode/internal/endorsement"
	"github.com/hyperledger/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger/fabric-private-chaincode/internal/utils"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/protoutil"
)

var logger = flogging.MustGetLogger("ecc")

// EnclaveChaincode struct
type EnclaveChaincode struct {
	Enclave   Enclave
	Validator endorsement.Validation
	Extractor Extractors
	Ercc      ercc.Stub
}

// Init sets the chaincode state to "init"
func (t *EnclaveChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke receives transactions and forwards to op handlers
func (t *EnclaveChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, _ := stub.GetFunctionAndParameters()
	logger.Infof("Invoke is running [%s]", function)

	switch function {
	case "__initEnclave":
		return t.initEnclave(stub)
	case "__invoke":
		return t.invoke(stub)
	case "__endorse":
		return t.endorse(stub)
	default:
		return shim.Error("invalid invocation")
	}
}

func (t *EnclaveChaincode) initEnclave(stub shim.ChaincodeStubInterface) pb.Response {
	// extract all enclave inputs from invocation params
	initMsg, err := t.Extractor.GetInitEnclaveMessage(stub)
	if err != nil {
		errMsg := fmt.Sprintf("getting initEnclave msg failed: %s", err.Error())
		logger.Errorf(errMsg)
		return shim.Error(errMsg)
	}

	chaincodeParams, err := t.Extractor.GetChaincodeParams(stub)
	if err != nil {
		errMsg := fmt.Sprintf("getting chaincode params failed: %s", err.Error())
		logger.Errorf(errMsg)
		return shim.Error(errMsg)
	}
	serializedChaincodeParams, err := protoutil.Marshal(chaincodeParams)
	if err != nil {
		return shim.Error(err.Error())
	}

	hostParams, err := t.Extractor.GetHostParams(stub)
	if err != nil {
		errMsg := fmt.Sprintf("getting host params failed: %s", err.Error())
		logger.Errorf(errMsg)
		return shim.Error(errMsg)
	}

	serializedHostParams, err := protoutil.Marshal(hostParams)
	if err != nil {
		return shim.Error(err.Error())
	}

	// main enclave initialization function
	credentialsBytes, err := t.Enclave.Init(serializedChaincodeParams, serializedHostParams, initMsg.AttestationParams)
	if err != nil {
		errMsg := fmt.Sprintf("Enclave Init function failed: %s", err.Error())
		logger.Errorf(errMsg)
		return shim.Error(errMsg)
	}

	// return credentials
	return shim.Success([]byte(base64.StdEncoding.EncodeToString(credentialsBytes)))
}

func (t *EnclaveChaincode) invoke(stub shim.ChaincodeStubInterface) pb.Response {
	var errMsg string

	serializedChaincodeRequest, err := t.Extractor.GetSerializedChaincodeRequest(stub)
	if err != nil {
		errMsg = fmt.Sprintf("cannot get chaincode request message from input: %s", err.Error())
		logger.Errorf(errMsg)
		return shim.Error(errMsg)
	}

	signedChaincodeResponseMessage, errInvoke := t.Enclave.ChaincodeInvoke(stub, serializedChaincodeRequest)
	if errInvoke != nil {
		errMsg = fmt.Sprintf("t.Enclave.Invoke failed: %s", errInvoke)
		logger.Errorf(errMsg)
		// likely a chaincode error, so we still want response go back ...
	}

	signedChaincodeResponseMessageB64 := []byte(base64.StdEncoding.EncodeToString(signedChaincodeResponseMessage))
	logger.Debugf("base64-encoded response message: '%s'", signedChaincodeResponseMessageB64)

	var response pb.Response
	if errInvoke == nil {
		response = pb.Response{
			Status:  shim.OK,
			Payload: signedChaincodeResponseMessageB64,
			Message: errMsg,
		}
	} else {
		response = pb.Response{
			Status:  shim.ERROR,
			Payload: signedChaincodeResponseMessageB64,
			Message: errMsg,
		}
	}

	return response
}

func (t *EnclaveChaincode) endorse(stub shim.ChaincodeStubInterface) pb.Response {

	chaincodeParams, err := t.Extractor.GetChaincodeParams(stub)
	if err != nil {
		errMsg := fmt.Sprintf("cannot extract chaincode params: %s", err.Error())
		logger.Errorf(errMsg)
		return shim.Error(errMsg)
	}

	signedResponseMsg, responseMsg, err := t.Extractor.GetChaincodeResponseMessages(stub)
	if err != nil {
		errMsg := fmt.Sprintf("cannot extract chaincode response message: %s", err.Error())
		logger.Errorf(errMsg)
		return shim.Error(errMsg)
	}

	logger.Infof("try to get credentials from ERCC for channel: %s ccId: %s EnclaveId: %s ", chaincodeParams.ChannelId, chaincodeParams.ChaincodeId, responseMsg.EnclaveId)

	// get corresponding enclave credentials from ercc
	credentials, err := t.Ercc.QueryEnclaveCredentials(stub, chaincodeParams.ChannelId, chaincodeParams.ChaincodeId, responseMsg.EnclaveId)
	if err != nil {
		return shim.Error(err.Error())
	}
	if credentials == nil {
		return shim.Error(fmt.Sprintf("no credentials found for enclaveId = %s", responseMsg.EnclaveId))
	}

	attestedData, err := utils.UnmarshalAttestedData(credentials.SerializedAttestedData)
	if err != nil {
		return shim.Error(err.Error())
	}

	// check cc params match credentials
	// check cc params chaincode def
	if !ccParamsMatch(attestedData.CcParams, chaincodeParams) {
		return shim.Error("ccParams don't match")
	}

	// check cc param.MSPID matches MSPID of endorser (Post-MVP)

	// validate enclave endorsement signature
	logger.Debugf("Validating endorsement")
	err = t.Validator.Validate(signedResponseMsg, attestedData)
	if err != nil {
		return shim.Error(err.Error())
	}

	// replay read/writes from kvrwset from Enclave (to prepare commitment to ledger) and extract kvrwset for subsequent validation
	logger.Debugf("Replaying rwset")
	err = t.Validator.ReplayReadWrites(stub, responseMsg.FpcRwSet)
	if err != nil {
		return shim.Error(err.Error())
	}

	logger.Debugf("Endorsement successful")
	return shim.Success([]byte("OK")) // make sure we have a non-empty return on success so we can distinguish success from failure in cli ...
}

func ccParamsMatch(expected, actual *protos.CCParameters) bool {
	return expected.ChaincodeId == actual.ChaincodeId &&
		expected.ChannelId == actual.ChannelId &&
		expected.Version == actual.Version &&
		expected.Sequence == actual.Sequence
}
