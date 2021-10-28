package chaincode

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-private-chaincode/ecc_go/chaincode/enclave"
	"github.com/hyperledger/fabric-private-chaincode/ecc_go/chaincode/ercc"
	"github.com/hyperledger/fabric-private-chaincode/ecc_go/chaincode/fakes"
	"github.com/hyperledger/fabric-private-chaincode/internal/crypto"
	"github.com/hyperledger/fabric-private-chaincode/internal/endorsement"
	"github.com/hyperledger/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

//go:generate counterfeiter -o fakes/enclave.go -fake-name EnclaveStub . enclaveStub
//lint:ignore U1000 This is just used to generate fake
type enclaveStub interface {
	enclave.StubInterface
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

func newECC(ec *enclave.EnclaveStub, val *fakes.Validator, ex *fakes.Extractors, ercc *fakes.ErccStub) *EnclaveChaincode {
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

func newRealEc() *enclave.EnclaveStub {
	return enclave.NewEnclaveStub()
}

func TestInitECC(t *testing.T) {
	_, val, ex, ercc := newFakes()
	ecc := newECC(newRealEc(), val, ex, ercc)
	stub := &fakes.ChaincodeStub{}

	// test init
	r := ecc.Init(stub)
	assert.Equal(t, shim.Success(nil), r)

	// test invalid invocation
	stub.GetFunctionAndParametersReturns("whatever", nil)
	r = ecc.Invoke(stub)
	assert.Equal(t, shim.Error("invalid invocation"), r)
}

func TestEnclave(t *testing.T) {
	stub := &fakes.ChaincodeStub{}
	stub.GetFunctionAndParametersReturns("__initEnclave", nil)
	_, _, ex, _ := newFakes()
	ecc := newECC(newRealEc(), nil, ex, nil)

	attestParams := []byte("someAttestationParams")
	ccParams := &protos.CCParameters{
		ChaincodeId: "SomeChaincodeId",
	}
	hostParams := &protos.HostParameters{
		PeerMspId:    "",
		PeerEndpoint: "",
		Certificate:  nil,
	}

	ex.GetInitEnclaveMessageReturns(&protos.InitEnclaveMessage{AttestationParams: attestParams}, nil)
	ex.GetChaincodeParamsReturns(ccParams, nil)
	ex.GetHostParamsReturns(hostParams, nil)

	r := ecc.Invoke(stub)
	assert.EqualValues(t, shim.OK, r.Status)
	payload, err := base64.StdEncoding.DecodeString(string(r.Payload))
	assert.NoError(t, err)

	credentials := &protos.Credentials{}
	proto.Unmarshal(payload, credentials)

	stub.GetFunctionAndParametersReturns("__invoke", nil)
	ep := &crypto.EncryptionProviderImpl{
		CSP: crypto.GetDefaultCSP(),
		GetCcEncryptionKey: func() ([]byte, error) {
			attestedData := &protos.AttestedData{}
			err := proto.Unmarshal(credentials.SerializedAttestedData.GetValue(), attestedData)
			return []byte(base64.StdEncoding.EncodeToString(attestedData.GetChaincodeEk())), err
		}}

	ctx, _ := ep.NewEncryptionContext()
	requestBytes, _ := ctx.Conceal("init", []string{"House"})
	request, _ := base64.StdEncoding.DecodeString(requestBytes)
	chaincodeRequestMessage := &protos.ChaincodeRequestMessage{}
	proto.Unmarshal(request, chaincodeRequestMessage)
	ex.GetSerializedChaincodeRequestReturns(request, nil)
	r = ecc.Invoke(stub)

	requestBytes, _ = ctx.Conceal("create", []string{"Auction"})
	request, _ = base64.StdEncoding.DecodeString(requestBytes)
	proto.Unmarshal(request, chaincodeRequestMessage)
	ex.GetSerializedChaincodeRequestReturns(request, nil)
	r = ecc.Invoke(stub)

	assert.EqualValues(t, shim.OK, r.Status)
	p, err := base64.StdEncoding.DecodeString(string(r.Payload))
	assert.NoError(t, err)
	fmt.Println(p)

}

func TestEndorse(t *testing.T) {
	stub := &fakes.ChaincodeStub{}
	stub.GetFunctionAndParametersReturns("__endorse", nil)
	_, val, ex, ercc := newFakes()
	ecc := newECC(newRealEc(), val, ex, ercc)
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
	expectError(t, fmt.Sprintf("ccParams don't match"), r)

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
