package chaincode

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-private-chaincode/ecc/chaincode/ercc"
	"github.com/hyperledger/fabric-private-chaincode/ecc/chaincode/fakes"
	"github.com/hyperledger/fabric-private-chaincode/internal/endorsement"
	"github.com/hyperledger/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/anypb"
)

//go:generate counterfeiter -o fakes/enclave.go -fake-name EnclaveStub . enclaveStub
//lint:ignore U1000 This is just used to generate fake
type enclaveStub interface {
	Enclave
}

//go:generate counterfeiter -o fakes/utils.go -fake-name Extractors . extractors
//lint:ignore U1000 This is just used to generate fake
type extractors interface {
	Extractors
}

//go:generate counterfeiter -o fakes/validation.go -fake-name Validator . validator
//lint:ignore U1000 This is just used to generate fake
type validator interface {
	endorsement.Validation
}

//go:generate counterfeiter -o fakes/chaincodestub.go -fake-name ChaincodeStub . chaincodeStub
//lint:ignore U1000 This is just used to generate fake
type chaincodeStub interface {
	shim.ChaincodeStubInterface
}

//go:generate counterfeiter -o fakes/ercc.go -fake-name ErccStub . erccStub
//lint:ignore U1000 This is just used to generate fake
type erccStub interface {
	ercc.Stub
}

func newECC(ec *fakes.EnclaveStub, val *fakes.Validator, ex *fakes.Extractors, ercc *fakes.ErccStub) *EnclaveChaincode {
	return &EnclaveChaincode{
		Enclave:   ec,
		Validator: val,
		Extractor: ex,
		Ercc:      ercc,
	}
}

func newFakes() (*fakes.EnclaveStub, *fakes.Validator, *fakes.Extractors, *fakes.ErccStub) {
	return &fakes.EnclaveStub{}, &fakes.Validator{}, &fakes.Extractors{}, &fakes.ErccStub{}
}

func TestInitECC(t *testing.T) {
	ec, val, ex, ercc := newFakes()
	ecc := newECC(ec, val, ex, ercc)
	stub := &fakes.ChaincodeStub{}

	// test init
	r := ecc.Init(stub)
	assert.Equal(t, shim.Success(nil), r)

	// test invalid invocation
	stub.GetFunctionAndParametersReturns("whatever", nil)
	r = ecc.Invoke(stub)
	assert.Equal(t, shim.Error("invalid invocation"), r)
}

func TestInitEnclave(t *testing.T) {
	stub := &fakes.ChaincodeStub{}
	stub.GetFunctionAndParametersReturns("__initEnclave", nil)
	ec, _, ex, _ := newFakes()
	ecc := newECC(ec, nil, ex, nil)
	expectedErr := fmt.Errorf("some error")
	expectedCCParams := &protos.CCParameters{
		ChaincodeId: "SomeChaincodeId",
	}
	expectedHostParams := &protos.HostParameters{
		PeerMspId:    "",
		PeerEndpoint: "",
		Certificate:  nil,
	}

	// error getting attestation params
	ex.GetInitEnclaveMessageReturns(nil, expectedErr)
	r := ecc.Invoke(stub)
	expectError(t, fmt.Sprintf("getting initEnclave msg failed: %s", expectedErr), r)

	// error getting chaincode params
	ex.GetInitEnclaveMessageReturns(&protos.InitEnclaveMessage{AttestationParams: []byte("someAttestationParams")}, nil)
	ex.GetChaincodeParamsReturns(nil, expectedErr)
	r = ecc.Invoke(stub)
	expectError(t, fmt.Sprintf("getting chaincode params failed: %s", expectedErr), r)

	// error getting host params
	ex.GetChaincodeParamsReturns(expectedCCParams, nil)
	ex.GetHostParamsReturns(nil, expectedErr)
	r = ecc.Invoke(stub)
	expectError(t, fmt.Sprintf("getting host params failed: %s", expectedErr), r)

	// error when init enclave
	ex.GetHostParamsReturns(expectedHostParams, nil)
	ec.InitReturns(nil, expectedErr)
	r = ecc.Invoke(stub)
	expectError(t, fmt.Sprintf("Enclave Init function failed: %s", expectedErr), r)

	// no error
	expectedCreds := []byte("someCredentials")
	ec.InitReturns(expectedCreds, nil)
	r = ecc.Invoke(stub)
	assert.EqualValues(t, shim.OK, r.Status)
	p, err := base64.StdEncoding.DecodeString(string(r.Payload))
	assert.NoError(t, err)
	assert.EqualValues(t, expectedCreds, p)
}

func TestInvokeEnclave(t *testing.T) {
	stub := &fakes.ChaincodeStub{}
	stub.GetFunctionAndParametersReturns("__invoke", nil)
	ec, _, ex, _ := newFakes()
	ecc := newECC(ec, nil, ex, nil)
	expectedErr := fmt.Errorf("some error")
	expectedResp := []byte("someResponse")

	// error getting chaincode request
	ex.GetSerializedChaincodeRequestReturns(nil, expectedErr)
	r := ecc.Invoke(stub)
	expectError(t, fmt.Sprintf("cannot get chaincode request message from input: %s", expectedErr), r)

	// error when invoking enclave
	ex.GetSerializedChaincodeRequestReturns([]byte("someChaincodeRequest"), nil)
	ec.ChaincodeInvokeReturns(expectedResp, expectedErr)
	r = ecc.Invoke(stub)
	expectError(t, fmt.Sprintf("t.Enclave.Invoke failed: %s", expectedErr), r)
	p, err := base64.StdEncoding.DecodeString(string(r.Payload))
	assert.NoError(t, err)
	assert.EqualValues(t, expectedResp, p)

	// no error
	ex.GetSerializedChaincodeRequestReturns([]byte("someChaincodeRequest"), nil)
	ec.ChaincodeInvokeReturns(expectedResp, nil)
	r = ecc.Invoke(stub)
	assert.EqualValues(t, shim.OK, r.Status)
	p, err = base64.StdEncoding.DecodeString(string(r.Payload))
	assert.NoError(t, err)
	assert.EqualValues(t, expectedResp, p)
	s, scr := ec.ChaincodeInvokeArgsForCall(1)
	assert.Equal(t, stub, s)
	assert.Equal(t, []byte("someChaincodeRequest"), scr)
}

func TestEndorse(t *testing.T) {
	stub := &fakes.ChaincodeStub{}
	stub.GetFunctionAndParametersReturns("__endorse", nil)
	ec, val, ex, ercc := newFakes()
	ecc := newECC(ec, val, ex, ercc)
	expectedErr := fmt.Errorf("some error")
	expectedCCParams := &protos.CCParameters{
		ChaincodeId: "someCCID",
		Version:     "someVersion",
		Sequence:    1,
		ChannelId:   "someChannel",
	}
	expectedSignedResp := &protos.SignedChaincodeResponseMessage{
		ChaincodeResponseMessage: []byte("someMessage"),
		Signature:                []byte("someSignature"),
	}

	expectedResp := &protos.ChaincodeResponseMessage{
		EncryptedResponse:           nil,
		FpcRwSet:                    nil,
		Proposal:                    nil,
		ChaincodeRequestMessageHash: nil,
		EnclaveId:                   "someEnclaveId",
	}

	// error getting chaincode request
	ex.GetChaincodeParamsReturns(nil, expectedErr)
	r := ecc.Invoke(stub)
	expectError(t, fmt.Sprintf("cannot extract chaincode params: %s", expectedErr), r)

	// error getting chaincode response messages
	ex.GetChaincodeParamsReturns(expectedCCParams, nil)
	ex.GetChaincodeResponseMessagesReturns(nil, nil, expectedErr)
	r = ecc.Invoke(stub)
	expectError(t, fmt.Sprintf("cannot extract chaincode response message: %s", expectedErr), r)

	// queryEnclaveCredentials returns error
	ex.GetChaincodeParamsReturns(expectedCCParams, nil)
	ex.GetChaincodeResponseMessagesReturns(expectedSignedResp, expectedResp, nil)
	ercc.QueryEnclaveCredentialsReturns(nil, expectedErr)
	r = ecc.Invoke(stub)
	expectError(t, fmt.Sprintf("%s", expectedErr), r)

	// credentials not found error
	ex.GetChaincodeParamsReturns(expectedCCParams, nil)
	ex.GetChaincodeResponseMessagesReturns(expectedSignedResp, expectedResp, nil)
	ercc.QueryEnclaveCredentialsReturns(nil, nil)
	r = ecc.Invoke(stub)
	expectError(t, fmt.Sprintf("no credentials found for enclaveId = %s", expectedResp.EnclaveId), r)

	// ccparams do not match error
	serializedAttestedData, _ := anypb.New(
		&protos.AttestedData{
			CcParams: &protos.CCParameters{
				ChaincodeId: "someOtherChaincodeId",
				Version:     "someOtherVersion",
				Sequence:    0,
				ChannelId:   "someOtherChannel",
			},
		})
	expectedCred := &protos.Credentials{
		SerializedAttestedData: serializedAttestedData,
	}
	ex.GetChaincodeParamsReturns(expectedCCParams, nil)
	ex.GetChaincodeResponseMessagesReturns(expectedSignedResp, expectedResp, nil)
	ercc.QueryEnclaveCredentialsReturns(expectedCred, nil)
	r = ecc.Invoke(stub)
	expectError(t, "ccParams don't match", r)

	// validate error
	serializedAttestedData, _ = anypb.New(
		&protos.AttestedData{
			CcParams: expectedCCParams,
		})
	expectedCred = &protos.Credentials{
		SerializedAttestedData: serializedAttestedData,
	}
	ex.GetChaincodeParamsReturns(expectedCCParams, nil)
	ex.GetChaincodeResponseMessagesReturns(expectedSignedResp, expectedResp, nil)
	ercc.QueryEnclaveCredentialsReturns(expectedCred, nil)
	val.ValidateReturns(expectedErr)
	r = ecc.Invoke(stub)
	expectError(t, expectedErr.Error(), r)

	// error when checking rwset
	ex.GetChaincodeParamsReturns(expectedCCParams, nil)
	ex.GetChaincodeResponseMessagesReturns(expectedSignedResp, expectedResp, nil)
	ercc.QueryEnclaveCredentialsReturns(expectedCred, nil)
	val.ValidateReturns(nil)
	val.ReplayReadWritesReturns(expectedErr)
	r = ecc.Invoke(stub)
	expectError(t, expectedErr.Error(), r)

	// no error
	ex.GetChaincodeParamsReturns(expectedCCParams, nil)
	ex.GetChaincodeResponseMessagesReturns(expectedSignedResp, expectedResp, nil)
	ercc.QueryEnclaveCredentialsReturns(expectedCred, nil)
	val.ValidateReturns(nil)
	val.ReplayReadWritesReturns(nil)
	r = ecc.Invoke(stub)
	assert.EqualValues(t, shim.OK, r.Status)
	assert.EqualValues(t, []byte("OK"), r.Payload)
}

func expectError(t *testing.T, errorMsg string, r peer.Response) {
	assert.EqualValues(t, shim.ERROR, r.Status)
	assert.EqualValues(t, errorMsg, r.Message)
}
