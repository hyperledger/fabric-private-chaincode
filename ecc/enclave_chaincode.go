/*
   Copyright IBM Corp. All Rights Reserved.
   Copyright 2020 Intel Corporation

   SPDX-License-Identifier: Apache-2.0
*/

package ecc

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger-labs/fabric-private-chaincode/ecc/crypto"
	"github.com/hyperledger-labs/fabric-private-chaincode/ecc/enclave"
	"github.com/hyperledger-labs/fabric-private-chaincode/ecc/ercc"
	"github.com/hyperledger-labs/fabric-private-chaincode/ecc/tlcc"
	fpcpb "github.com/hyperledger-labs/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger-labs/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/core/endorser"
)

const enclaveLibFile = "enclave/lib/enclave.signed.so"

var logger = flogging.MustGetLogger("ecc")

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
	} else if function == "__getEnclavePk" { //get Enclave PK
		return t.getEnclavePk(stub)
	} else if function == "initEnclave" {
		return t.initEnclave(stub)
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

	// write mrenclave to ledger
	mrenclave, err := t.enclave.MrEnclave()
	if err != nil {
		return shim.Error(err.Error())
	}

	// TODO this will change in the future; we will need to fetch MRENCLAVE from the chaincode definition
	// check if MRENCLAVE is already on the ledger; if not, write it; if exists compare with enclave;
	exists, err := stub.GetState(utils.MrEnclaveStateKey)
	if err != nil {
		return shim.Error(err.Error())
	}

	if len(exists) == 0 {
		logger.Debugf("Set chaincode with mrenclave=%s", mrenclave)
		if err := stub.PutState(utils.MrEnclaveStateKey, []byte(mrenclave)); err != nil {
			return shim.Error(err.Error())
		}
	} else {
		if mrenclave != string(exists) {
			errMsg := fmt.Sprintf("ecc: MRENCLAVE has already been defined for this chaincode. Expected %s but got %s", string(exists), mrenclave)
			logger.Errorf(errMsg)
			return shim.Error(errMsg)
		}
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
// initEnclave -
// ============================================================
func (t *EnclaveChaincode) initEnclave(stub shim.ChaincodeStubInterface) pb.Response {
	// TODO admin authentication?
	// TODO check one time create?
	logger.Debugf("initEnclave")

	// check if there is already an enclave
	if t.enclave == nil {
		return shim.Error("ecc: Enclave has already been initialized! Destroy first!!")
	}

	args := stub.GetStringArgs()

	if len(args) != 2 {
		errMsg := fmt.Sprintf("initEnclave: unexpected params number")
		logger.Errorf(errMsg)
		return shim.Error(errMsg)
	}

	// grab initenclave proto <<<
	b64InitEnclave := args[1]
	logger.Debugf("b64InitEnclave: %s", b64InitEnclave)
	initEnclaveBytes, err := base64.StdEncoding.DecodeString(b64InitEnclave)
	if err != nil {
		errMsg := fmt.Sprintf("initEnclave: initEnclaveBytes: %s", err)
		logger.Errorf(errMsg)
		return shim.Error(errMsg)
	}

	initEnclaveProto := &fpcpb.InitEnclaveMessage{}
	err = proto.Unmarshal(initEnclaveBytes, initEnclaveProto)
	if err != nil {
		errMsg := fmt.Sprintf("initEnclave: cannot unmarshall initEnclave proto")
		logger.Errorf(errMsg)
		return shim.Error(errMsg)
	}

	hostParamsString := initEnclaveProto.GetPeerEndpoint()
	if hostParamsString == "" {
		errMsg := fmt.Sprintf("initEnclave: empty peer endpoint")
		logger.Errorf(errMsg)
		return shim.Error(errMsg)
	}
	logger.Debugf("HostParams: %s", hostParamsString)

	b64AttestationParamsBytes := initEnclaveProto.GetAttestationParams()
	if b64AttestationParamsBytes == nil {
		errMsg := fmt.Sprintf("initEnclave: null attestation params")
		logger.Errorf(errMsg)
		return shim.Error(errMsg)
	}

	attestationParamsBytes, err := base64.StdEncoding.DecodeString(string(b64AttestationParamsBytes))
	if err != nil {
		errMsg := fmt.Sprintf("initEnclave: bad input b64 attestation parameters")
		logger.Errorf(errMsg)
		return shim.Error(errMsg)
	}
	// grab initenclave proto >>>

	// create Host_Parameters <<<
	creatorMspId, err := cid.GetMSPID(stub)
	if err != nil {
		errMsg := fmt.Sprintf("initEnclave: get creator's mspid: %s", err)
		logger.Errorf(errMsg)
		return shim.Error(errMsg)
	}
	logger.Debugf("Creator MSPID: %s", creatorMspId)

	hostParametersProto := &fpcpb.HostParameters{
		//TODO: rename PeerMspId to CreatorMspId, because the value is the creator's mspid, not the peer's
		PeerMspId:    creatorMspId,
		PeerEndpoint: hostParamsString,
	}
	hostParametersProtoBytes, err := proto.Marshal(hostParametersProto)
	if err != nil {
		errMsg := fmt.Sprintf("initEnclave: cannot marshall host params")
		logger.Errorf(errMsg)
		return shim.Error(errMsg)
	}
	// create HostParameters >>>

	// create CCParameters <<<
	// get signed proposal to extract chaicode id
	signedProposal, err := stub.GetSignedProposal()
	if err != nil {
		errMsg := fmt.Sprintf("initEnclave: cannot get signed proposal")
		logger.Errorf(errMsg)
		return shim.Error(errMsg)
	}
	unpackedProposal, err := endorser.UnpackProposal(signedProposal)
	if err != nil {
		errMsg := fmt.Sprintf("initEnclave: cannot unpack proposal")
		logger.Errorf(errMsg)
		return shim.Error(errMsg)
	}
	logger.Debugf("Chaincode id: %s", unpackedProposal.ChaincodeName)

	//get chaincode definition
	ccDef, err := utils.GetChaincodeDefinition(unpackedProposal.ChaincodeName, stub)
	if err != nil {
		errMsg := fmt.Sprintf("initEnclave: cannot get ccdefinition")
		logger.Errorf(errMsg)
		return shim.Error(errMsg)
	}
	logger.Debugf("Chaincode definition: %d %s", ccDef.Sequence, ccDef.Version)

	//get channel id
	channelName := stub.GetChannelID()
	logger.Debugf("Channel ID: %s", channelName)

	// produce cc params
	ccParametersProto := &fpcpb.CCParameters{
		ChaincodeId: unpackedProposal.ChaincodeName,
		Version:     ccDef.Version,
		Sequence:    ccDef.Sequence,
		ChannelId:   channelName,
	}
	ccParametersBytes, err := proto.Marshal(ccParametersProto)
	if err != nil {
		errMsg := fmt.Sprintf("initEnclave: cannot marshall cc params")
		logger.Errorf(errMsg)
		return shim.Error(errMsg)
	}
	// create CCParameters >>>

	// create new Enclave
	credentials, err := t.enclave.Create(enclaveLibFile, ccParametersBytes, attestationParamsBytes, hostParametersProtoBytes)
	if err != nil {
		errMsg := fmt.Sprintf("initEnclave: failed: %s", err)
		logger.Errorf(errMsg)
		return shim.Error(errMsg)
	}

	// debug credentials, print enclave verification key <<<
	credentialsProto := &fpcpb.Credentials{}
	err = proto.Unmarshal(credentials, credentialsProto)
	if err != nil {
		errMsg := fmt.Sprintf("initEnclave: cannot unmarshall credentials")
		logger.Errorf(errMsg)
		return shim.Error(errMsg)
	}
	attestedDataAny := credentialsProto.GetSerializedAttestedData()
	if attestedDataAny == nil {
		errMsg := fmt.Sprintf("initEnclave: no attested data")
		logger.Errorf(errMsg)
		return shim.Error(errMsg)
	}
	attestedDataProto := &fpcpb.AttestedData{}
	err = ptypes.UnmarshalAny(attestedDataAny, attestedDataProto)
	if err != nil {
		errMsg := fmt.Sprintf("initEnclave: cannot unmarshall attested data")
		logger.Errorf(errMsg)
		return shim.Error(errMsg)
	}
	ccParams := attestedDataProto.GetCcParams()
	if ccParams == nil {
		errMsg := fmt.Sprintf("initEnclave: error getting cc params from attested data")
		logger.Errorf(errMsg)
		return shim.Error(errMsg)
	}
	ccId := ccParams.GetChaincodeId()
	enclaveVKBytes := attestedDataProto.GetEnclaveVk()
	logger.Debugf("initEnclave: enclave verification key: %s", string(enclaveVKBytes))
	logger.Debugf("initEnclave: chaincode id (from attested data): %s", ccId)
	if ccId != unpackedProposal.ChaincodeName {
		errMsg := fmt.Sprintf("initEnclave: wrong serialized chaincode id in attested data")
		logger.Errorf(errMsg)
		return shim.Error(errMsg)
	}
	// debug credentials, print enclave verification key >>>

	// build return value
	b64Credentials := base64.StdEncoding.EncodeToString([]byte(credentials))
	logger.Debugf("initEnclave b64 response: %s", b64Credentials)

	return shim.Success([]byte(b64Credentials))
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
	logger.Debugf("string args: %s", argss)
	jsonArgs, err := json.Marshal(argss)
	if err != nil {
		return shim.Error(err.Error())
	}

	logger.Debugf("json args: %s", jsonArgs)

	//pk := []byte(nil) // we don't really support a secure channel to the client yet ..
	//// TODO: one of the place to fix when integrating end-to-end secure channel to client

	// call enclave
	var errMsg string
	b64ChaincodeResponseMessage, errInvoke := t.enclave.Invoke(jsonArgs, stub)
	if errInvoke != nil {
		errMsg = fmt.Sprintf("t.enclave.Invoke failed: %s", errInvoke)
		logger.Errorf(errMsg)
		// likely a chaincode error, so we stil want response go back ...
	}

	//enclavePk, errPk := t.enclave.GetPublicKey()
	//if errPk != nil {
	//	errMsg = fmt.Sprintf("invoke t.enclave.GetPublicKey failed. Reason: %s", err)
	//	logger.Errorf(errMsg)
	//	// return (and ignore any potential response) as this is a more systematic error
	//	return shim.Error(errMsg)
	//}
	//_ = enclavePk

	//fpcResponse := &utils.Response{
	//	ResponseData: responseData,
	//	Signature:    signature,
	//	PublicKey:    enclavePk,
	//}
	//responseBytes, _ := json.Marshal(fpcResponse)

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
	logger.Debugf("invoke response: %v", response)
	return response
}

// ============================================================
// getEnclavePk -
// ============================================================
func (t *EnclaveChaincode) getEnclavePk(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Error("ecc:  getEnclavePk disabled")
}

// TODO: check if Destroy is called
func (t *EnclaveChaincode) Destroy() {
	if err := t.enclave.Destroy(); err != nil {
		errMsg := fmt.Sprintf("t.enclave.Destroy failed: %s", err)
		logger.Errorf(errMsg)
		panic(errMsg)
	}
}
