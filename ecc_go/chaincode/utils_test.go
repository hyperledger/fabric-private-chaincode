package chaincode

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-private-chaincode/ercc/registry/fakes"
	"github.com/hyperledger/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-protos-go/peer/lifecycle"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

const (
	channelId            = "mychannel"
	chaincodeId          = "myChaincode"
	chaincodeSequence    = 1
	chaincodeVersion     = "v1"
	lifecycleChaincodeId = "_lifecycle"
	Mspid                = "myMSP"
	PeerEndpoint         = "someEndpoint"
)

func TestGetInitEnclaveMessage(t *testing.T) {
	ex := &ExtractorImpl{}

	// no initEnclaveMessage
	stub := &fakes.ChaincodeStub{}
	initMsg, err := ex.GetInitEnclaveMessage(stub)
	assert.Nil(t, initMsg)
	assert.Error(t, err)

	msg := &protos.InitEnclaveMessage{PeerEndpoint: PeerEndpoint}

	// wrong message encoding
	stub = &fakes.ChaincodeStub{}
	stub.GetStringArgsReturns([]string{"no-base64", string(protoutil.MarshalOrPanic(msg))})
	initMsg, err = ex.GetInitEnclaveMessage(stub)
	assert.Nil(t, initMsg)
	assert.Error(t, err)

	// not an initMessage
	stub = &fakes.ChaincodeStub{}
	stub.GetStringArgsReturns([]string{"no-base64", utils.MarshallProtoBase64(msg)})
	initMsg, err = ex.GetInitEnclaveMessage(stub)
	assert.NotNil(t, initMsg)
	assert.NoError(t, err)
	assertProtoEqual(t, msg, initMsg)
}

func TestGetSerializedChaincodeRequest(t *testing.T) {
	ex := &ExtractorImpl{}

	// no initEnclaveMessage
	stub := &fakes.ChaincodeStub{}
	serializedReqMsg, err := ex.GetSerializedChaincodeRequest(stub)
	assert.Nil(t, serializedReqMsg)
	assert.Error(t, err)

	expectedMsg := []byte("someChaincodeRequest")

	// no error
	stub = &fakes.ChaincodeStub{}
	stub.GetStringArgsReturns([]string{"no-base64", base64.StdEncoding.EncodeToString(expectedMsg)})
	serializedReqMsg, err = ex.GetSerializedChaincodeRequest(stub)
	assert.NoError(t, err)
	assert.EqualValues(t, expectedMsg, serializedReqMsg)
}

func TestGetChaincodeResponseMessages(t *testing.T) {
	ex := &ExtractorImpl{}

	// no response messages
	stub := &fakes.ChaincodeStub{}
	signedResp, resp, err := ex.GetChaincodeResponseMessages(stub)
	assert.Nil(t, signedResp)
	assert.Nil(t, resp)
	assert.Error(t, err)

	respMsg := &protos.ChaincodeResponseMessage{EnclaveId: "some_enclave_id"}
	signedRespMsg := &protos.SignedChaincodeResponseMessage{
		ChaincodeResponseMessage: utils.MarshalOrPanic(respMsg),
		Signature:                []byte("some_signature"),
	}

	// wrong message encoding
	stub = &fakes.ChaincodeStub{}
	stub.GetStringArgsReturns([]string{"no-base64", string(protoutil.MarshalOrPanic(signedRespMsg))})
	signedResp, resp, err = ex.GetChaincodeResponseMessages(stub)
	assert.Nil(t, signedResp)
	assert.Nil(t, resp)
	assert.Error(t, err)

	// no errors
	stub = &fakes.ChaincodeStub{}
	stub.GetStringArgsReturns([]string{"no-base64", utils.MarshallProtoBase64(signedRespMsg)})
	signedResp, resp, err = ex.GetChaincodeResponseMessages(stub)
	assert.NoError(t, err)
	assertProtoEqual(t, signedRespMsg, signedResp)
	assertProtoEqual(t, respMsg, resp)
}

func TestGetChaincodeParams(t *testing.T) {
	ex := &ExtractorImpl{}

	// getSignedProposal error
	stub := &fakes.ChaincodeStub{}
	stub.GetSignedProposalReturns(nil, fmt.Errorf("some error"))
	pp, err := ex.GetChaincodeParams(stub)
	assert.Nil(t, pp)
	assert.Error(t, err)

	// getChaincodeDefinition error
	stub = &fakes.ChaincodeStub{}
	signedProposal := &peer.SignedProposal{
		ProposalBytes: protoutil.MarshalOrPanic(
			&peer.Proposal{
				Payload: protoutil.MarshalOrPanic(
					&peer.ChaincodeProposalPayload{
						Input: protoutil.MarshalOrPanic(
							&peer.ChaincodeInvocationSpec{
								ChaincodeSpec: &peer.ChaincodeSpec{
									ChaincodeId: &peer.ChaincodeID{Name: chaincodeId},
								},
							}),
					}),
			}),
	}
	stub.GetSignedProposalReturns(signedProposal, nil)
	stub.InvokeChaincodeReturns(peer.Response{
		Status:  shim.ERROR,
		Message: "some error",
	})
	pp, err = ex.GetChaincodeParams(stub)
	assert.Nil(t, pp)
	assert.Error(t, err)

	// no errors
	stub = &fakes.ChaincodeStub{}
	stub.GetSignedProposalReturns(signedProposal, nil)
	stub.GetChannelIDReturns(channelId)
	stub.InvokeChaincodeReturns(peer.Response{
		Status: shim.OK,
		Payload: protoutil.MarshalOrPanic(
			&lifecycle.QueryChaincodeDefinitionResult{
				Sequence: chaincodeSequence,
				Version:  chaincodeVersion,
			}),
	})
	pp, err = ex.GetChaincodeParams(stub)
	assert.NotNil(t, pp)
	assert.NoError(t, err)

	f, _, ch := stub.InvokeChaincodeArgsForCall(0)
	assert.Equal(t, lifecycleChaincodeId, f)
	assert.Equal(t, channelId, ch)

	assert.EqualValues(t, chaincodeId, pp.GetChaincodeId())
	assert.EqualValues(t, chaincodeVersion, pp.GetVersion())
	assert.EqualValues(t, chaincodeSequence, pp.GetSequence())
	assert.EqualValues(t, channelId, pp.GetChannelId())
}

func TestGetHostParams(t *testing.T) {
	ex := &ExtractorImpl{}

	// getMSPID error
	stub := &fakes.ChaincodeStub{}
	hp, err := ex.GetHostParams(stub)
	assert.Nil(t, hp)
	assert.Error(t, err)

	// empty initMessage
	stub = &fakes.ChaincodeStub{}
	sid := &msp.SerializedIdentity{
		Mspid: Mspid,
	}
	stub.GetCreatorReturns(protoutil.MarshalOrPanic(sid), nil)
	hp, err = ex.GetHostParams(stub)
	assert.Nil(t, hp)
	assert.Error(t, err)

	// no errors
	stub = &fakes.ChaincodeStub{}
	stub.GetCreatorReturns(protoutil.MarshalOrPanic(sid), nil)
	initMsg := &protos.InitEnclaveMessage{PeerEndpoint: PeerEndpoint}
	stub.GetStringArgsReturns([]string{"someFunction", utils.MarshallProtoBase64(initMsg)})
	hp, err = ex.GetHostParams(stub)
	assert.NotNil(t, hp)
	assert.NoError(t, err)
	assert.EqualValues(t, Mspid, hp.GetPeerMspId())
	assert.EqualValues(t, PeerEndpoint, hp.GetPeerEndpoint())
	// Note that currently no certs are implemented
	assert.Nil(t, hp.GetCertificate())
}

func assertProtoEqual(t *testing.T, expected, actual proto.Message) bool {
	if !proto.Equal(expected, actual) {
		return assert.Fail(t, fmt.Sprintf("Not equal: \nexpected: %s\nactual : %s", expected, actual))
	}
	return true
}
