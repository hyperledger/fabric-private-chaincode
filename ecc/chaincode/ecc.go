/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"encoding/base64"
	"fmt"

	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-private-chaincode/ecc/chaincode/enclave"
	"github.com/hyperledger/fabric-private-chaincode/ecc/chaincode/ercc"
	"github.com/hyperledger/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger/fabric-private-chaincode/internal/utils"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/protoutil"
)

var logger = flogging.MustGetLogger("ecc")

// EnclaveChaincode struct
type EnclaveChaincode struct {
	enclave enclave.StubInterface
}

func NewChaincodeEnclave() shim.Chaincode {
	return &EnclaveChaincode{
		enclave: enclave.NewEnclaveStub(),
	}
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

	initMsg, err := extractInitEnclaveMessage(stub)
	if err != nil {
		errMsg := fmt.Sprintf("InitEnclave msg extraction failed: %s", err.Error())
		return shim.Error(errMsg)
	}

	// fetch cc params and host params
	chaincodeParams, err := extractChaincodeParams(stub)
	if err != nil {
		errMsg := fmt.Sprintf("chaincode params extraction failed: %s", err.Error())
		return shim.Error(errMsg)
	}
	serializedChaincodeParams, err := protoutil.Marshal(chaincodeParams)
	if err != nil {
		errMsg := fmt.Sprintf("chaincode params marshalling failed: %s", err.Error())
		return shim.Error(errMsg)
	}

	hostParams, err := extractHostParams(stub, initMsg)
	if err != nil {
		errMsg := fmt.Sprintf("host params extraction failed: %s", err.Error())
		return shim.Error(errMsg)
	}
	serializedHostParams, err := protoutil.Marshal(hostParams)
	if err != nil {
		errMsg := fmt.Sprintf("host params marshalling failed: %s", err.Error())
		return shim.Error(errMsg)
	}

	// main enclave initialization function
	credentialsBytes, err := t.enclave.Init(serializedChaincodeParams, serializedHostParams, initMsg.AttestationParams)
	if err != nil {
		errMsg := fmt.Sprintf("Enclave Init function failed: %s", err.Error())
		return shim.Error(errMsg)
	}

	// return credentials
	return shim.Success([]byte(base64.StdEncoding.EncodeToString(credentialsBytes)))
}

func (t *EnclaveChaincode) invoke(stub shim.ChaincodeStubInterface) pb.Response {
	// call enclave
	var errMsg string

	// prep chaincode request message as input
	_, args := stub.GetFunctionAndParameters()
	if len(args) != 1 {
		return shim.Error("no chaincodeRequestMessage as argument found")
	}
	chaincodeRequestMessageB64 := args[0]
	chaincodeRequestMessage, err := base64.StdEncoding.DecodeString(chaincodeRequestMessageB64)
	if err != nil {
		errMsg := fmt.Sprintf("cannot base64 decode ChaincodeRequestMessage ('%s'): %s", chaincodeRequestMessageB64, err.Error())
		return shim.Error(errMsg)
	}

	signedChaincodeResponseMessage, errInvoke := t.enclave.ChaincodeInvoke(stub, chaincodeRequestMessage)
	if errInvoke != nil {
		errMsg = fmt.Sprintf("t.enclave.Invoke failed: %s", errInvoke)
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

	chaincodeParams, err := extractChaincodeParams(stub)
	if err != nil {
		errMsg := fmt.Sprintf("cannot extract chaincode params: %s", err.Error())
		logger.Errorf(errMsg)
		return shim.Error(errMsg)
	}

	signedResponseMsg, responseMsg, err := extractChaincodeResponseMessages(stub)
	if err != nil {
		errMsg := fmt.Sprintf("cannot extract chaincode response message: %s", err.Error())
		logger.Errorf(errMsg)
		return shim.Error(errMsg)
	}

	logger.Infof("try to get credentials from ERCC for channel: %s ccId: %s EnclaveId: %s ", chaincodeParams.ChannelId, chaincodeParams.ChaincodeId, responseMsg.EnclaveId)

	// get corresponding enclave credentials from ercc
	credentials, err := ercc.QueryEnclaveCredentials(stub, chaincodeParams.ChannelId, chaincodeParams.ChaincodeId, responseMsg.EnclaveId)
	if err != nil {
		return shim.Error(err.Error())
	}

	if credentials == nil {
		return shim.Error(fmt.Sprintf("no credentials found for enclaveId = %s", responseMsg.EnclaveId))
	}

	var attestedData protos.AttestedData
	if err := ptypes.UnmarshalAny(credentials.SerializedAttestedData, &attestedData); err != nil {
		return shim.Error(err.Error())
	}

	// check cc params match credentials
	// check cc params chaincode def
	if ccParamsMatch(attestedData.CcParams, chaincodeParams) {
		return shim.Error("ccParams don't match")
	}

	// check cc param.MSPID matches MSPID of endorser (Post-MVP)

	// validate enclave endorsement signature
	logger.Debugf("Validating endorsement")
	err = utils.Validate(signedResponseMsg, &attestedData)
	if err != nil {
		return shim.Error(err.Error())
	}

	// replay read/writes from kvrwset from enclave (to prepare commitment to ledger) and extract kvrwset for subsequent validation
	logger.Debugf("Replaying rwset")
	err = utils.ReplayReadWrites(stub, responseMsg.FpcRwSet)
	if err != nil {
		return shim.Error(err.Error())
	}

	logger.Debugf("Endorsement successful")
	return shim.Success([]byte("OK")) // make sure we have a non-empty return on success so we can distinguish success from failure in cli ...
}

func ccParamsMatch(expected, actual *protos.CCParameters) bool {
	return expected.ChaincodeId != actual.ChaincodeId ||
		expected.ChannelId != actual.ChannelId ||
		expected.Version != actual.Version ||
		expected.Sequence != actual.Sequence
}
