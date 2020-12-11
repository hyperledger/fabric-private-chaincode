/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"encoding/base64"
	"fmt"

	"github.com/hyperledger/fabric/protoutil"

	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger-labs/fabric-private-chaincode/ecc/chaincode/enclave"
	"github.com/hyperledger-labs/fabric-private-chaincode/ecc/chaincode/ercc"
	"github.com/hyperledger-labs/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger-labs/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/flogging"
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

	// fetch cc params and host params
	chaincodeParams, err := extractChaincodeParams(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	serializedChaincodeParams, err := protoutil.Marshal(chaincodeParams)
	if err != nil {
		return shim.Error(err.Error())
	}

	hostParams, err := extractHostParams(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	serializedHostParams, err := protoutil.Marshal(hostParams)
	if err != nil {
		return shim.Error(err.Error())
	}

	attestationParams, err := extractAttestationParams(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	serializedAttestationParams, err := protoutil.Marshal(attestationParams)
	if err != nil {
		return shim.Error(err.Error())
	}

	credentialsBytes, err := t.enclave.Init(serializedChaincodeParams, serializedHostParams, serializedAttestationParams)
	if err != nil {
		return shim.Error(err.Error())
	}

	// return credentials
	return shim.Success([]byte(base64.StdEncoding.EncodeToString(credentialsBytes)))
}

func (t *EnclaveChaincode) invoke(stub shim.ChaincodeStubInterface) pb.Response {
	// call enclave
	var errMsg string
	b64ChaincodeResponseMessage, errInvoke := t.enclave.ChaincodeInvoke(stub)
	if errInvoke != nil {
		errMsg = fmt.Sprintf("t.enclave.Invoke failed: %s", errInvoke)
		logger.Errorf(errMsg)
		// likely a chaincode error, so we still want response go back ...
	}

	var response pb.Response
	if errInvoke == nil {
		response = pb.Response{
			Status:  shim.OK,
			Payload: b64ChaincodeResponseMessage,
			Message: errMsg,
		}
	} else {
		response = pb.Response{
			Status:  shim.ERROR,
			Payload: b64ChaincodeResponseMessage,
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

	responseMsg, err := extractChaincodeResponseMessage(stub)
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

	// replay read/writes from kvrwset from enclave (to prepare commitment to ledger) and extract kvrwset for subsequent validation
	readset, writeset, err := utils.ReplayReadWrites(stub, responseMsg.RwSet)
	if err != nil {
		return shim.Error(err.Error())
	}

	// validate enclave endorsement signature
	err = utils.Validate(responseMsg, readset, writeset, &attestedData)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func ccParamsMatch(expected, actual *protos.CCParameters) bool {
	return expected.ChaincodeId != actual.ChaincodeId ||
		expected.ChannelId != actual.ChannelId ||
		expected.Version != actual.Version ||
		expected.Sequence != actual.Sequence
}
