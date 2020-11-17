/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"fmt"

	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger-labs/fabric-private-chaincode/ecc_mock/chaincode/enclave"
	"github.com/hyperledger-labs/fabric-private-chaincode/ecc_mock/chaincode/ercc"
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
	chaincodeParams, err := ExtractChaincodeParams(stub)
	if err != nil {
		shim.Error(err.Error())
	}

	hostParams, err := extractHostParams(stub)
	if err != nil {
		shim.Error(err.Error())
	}

	attestationParams, err := extractAttestationParams(stub)
	if err != nil {
		shim.Error(err.Error())
	}

	credentialsBytes, err := t.enclave.Init(chaincodeParams, hostParams, attestationParams)
	if err != nil {
		shim.Error(err.Error())
	}

	// return credentials
	return shim.Success(credentialsBytes)
}

func (t *EnclaveChaincode) invoke(stub shim.ChaincodeStubInterface) pb.Response {

	// call enclave
	responseBytes, errInvoke := t.enclave.ChaincodeInvoke(stub)
	if errInvoke != nil {
		errMsg := fmt.Sprintf("t.enclave.ChaincodeInvoke failed: %s", errInvoke)
		logger.Errorf(errMsg)
		// likely a chaincode error, so we still want response go back ...
		return pb.Response{
			Status:  shim.ERROR,
			Payload: responseBytes,
			Message: errMsg,
		}
	}

	return shim.Success(responseBytes)
}

func (t *EnclaveChaincode) endorse(stub shim.ChaincodeStubInterface) pb.Response {

	chaincodeParams, err := ExtractChaincodeParams(stub)
	if err != nil {
		shim.Error(err.Error())
	}

	responseMsg, err := extractChaincodeResponseMessage(stub)
	if err != nil {
		shim.Error(err.Error())
	}

	// get corresponding enclave credentials from ercc
	credentials, err := ercc.QueryEnclaveCredentials(stub, chaincodeParams.ChannelId, chaincodeParams.ChaincodeId, responseMsg.EnclaveId)
	if err != nil {
		shim.Error(err.Error())
	}

	var attestedData protos.AttestedData
	if err := ptypes.UnmarshalAny(credentials.SerializedAttestedData, &attestedData); err != nil {
		shim.Error(err.Error())
	}

	// check cc params match credentials
	// check cc params chaincode def
	if ccParamsMatch(attestedData.CcParams, chaincodeParams) {
		shim.Error("ccParams don't match")
	}

	// check cc param.MSPID matches MSPID of endorser (Post-MVP)

	// extract read/writes from kvrwset,
	readset, writeset, err := utils.ReplayReadWrites(stub, responseMsg.RwSet)
	if err != nil {
		shim.Error(err.Error())
	}

	// validate enclave endorsement signature
	err = utils.Validate(responseMsg, readset, writeset, attestedData)
	if err != nil {
		shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func ccParamsMatch(expected, actual *protos.CCParameters) bool {
	return expected.ChaincodeId != actual.ChaincodeId ||
		expected.ChannelId != actual.ChannelId ||
		expected.Version != actual.Version ||
		expected.Sequence != actual.Sequence
}
