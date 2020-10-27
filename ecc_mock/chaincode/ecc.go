/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	enclave2 "github.com/hyperledger-labs/fabric-private-chaincode/ecc_mock/chaincode/enclave"
	"github.com/hyperledger-labs/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger-labs/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/msp"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/protoutil"
)

var logger = flogging.MustGetLogger("ecc")

// EnclaveChaincode struct
type EnclaveChaincode struct {
	enclave enclave2.MockEnclave
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
	// attestationParamsBase64 := stub.GetStringArgs()[0]

	// fetch cc params and host params

	// create enclave

	// ecall_init(cc params, host params, and attestation params)

	serializedIdentity := &msp.SerializedIdentity{
		Mspid:   "Org1MSP",
		IdBytes: []byte("some bytes"),
	}
	serializedUser := protoutil.MarshalOrPanic(serializedIdentity)

	credentials := &protos.Credentials{
		Attestation: []byte("{\"attestation_type\":\"simulated\",\"attestation\":\"MA==\"}"),
		SerializedAttestedData: &any.Any{
			TypeUrl: proto.MessageName(&protos.AttestedData{}),
			Value: protoutil.MarshalOrPanic(&protos.AttestedData{
				EnclaveVk: []byte("enclaveVKString"),
				CcParams: &protos.CCParameters{
					ChaincodeId: "ercc",
					Version:     "1.0",
					ChannelId:   "mychannel",
					Sequence:    1,
				},
				HostParams: &protos.HostParameters{
					PeerIdentity: serializedUser,
				},
			}),
		},
	}

	credentialBytes, err := proto.Marshal(credentials)
	if err != nil {
		shim.Error(err.Error())
	}

	logger.Infof("return some credentials: %s", credentials)

	// return credentials
	return shim.Success(credentialBytes)
}

func (t *EnclaveChaincode) invoke(stub shim.ChaincodeStubInterface) pb.Response {
	argsBase64 := stub.GetStringArgs()[1]
	if argsBase64 == "" {
		logger.Errorf("empty arguments")
		return shim.Error("empty arguments")
	}

	args, err := base64.StdEncoding.DecodeString(argsBase64)
	if err != nil {
		logger.Errorf(err.Error())
		return shim.Error(err.Error())
	}

	// call enclave
	var errMsg string
	responseData, signature, errInvoke := t.enclave.Invoke(args, nil, stub)
	if errInvoke != nil {
		errMsg = fmt.Sprintf("t.enclave.Invoke failed: %s", errInvoke)
		logger.Errorf(errMsg)
		// likely a chaincode error, so we stil want response go back ...
	}

	fpcResponse := &utils.Response{
		ResponseData: responseData,
		Signature:    signature,
		PublicKey:    []byte("TODO enclave public key"),
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
	return response
}

func (t *EnclaveChaincode) endorse(stub shim.ChaincodeStubInterface) pb.Response {

	// TODO

	// check again chaincode definition and enclave registry

	// perform enclave signature validation

	// perform reads and writes
	// for each  readwriteset
	// 	putState()
	// 	getState()

	return shim.Success(nil)
}
