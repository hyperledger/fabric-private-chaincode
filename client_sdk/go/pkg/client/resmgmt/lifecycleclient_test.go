/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resmgmt_test

import (
	"fmt"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/stretchr/testify/assert"

	"github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/client/fgosdkresmgmt"
	"github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/client/resmgmt/fakes"
	"github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/sgx"
)

//go:generate counterfeiter -o fakes/channelclient.go -fake-name ChannelClient . ChannelClient
//lint:ignore U1000 This is just used to generate fake
type chClient interface {
	resmgmt.ChannelClient
}

//go:generate counterfeiter -o fakes/credential_converter.go -fake-name CredentialConverter . credentialConverter
//lint:ignore U1000 This is just used to generate fake
type credConverter interface {
	resmgmt.CredentialConverter
}

const (
	channelID           = "mychannel"
	chaincodeId         = "my-fpc-chaincode"
	enclavePeerEndpoint = "mypeer.myorg.example.com"
	attestationType     = "simulation"
	expectedTxID        = "someTxID"
)

func setupClient(client resmgmt.ChannelClient, converter resmgmt.CredentialConverter) *resmgmt.Client {
	getChannelClient := func(channelId string) (resmgmt.ChannelClient, error) {
		return client, nil
	}

	return &resmgmt.Client{GetChannelClient: getChannelClient, Converter: converter}
}

func createClientContext(fabCtx context.Client) context.ClientProvider {
	return func() (context.Client, error) {
		return fabCtx, nil
	}
}

func TestCreateNewClient(t *testing.T) {
	clientCtx := fcmocks.NewMockContext(mspmocks.NewMockSigningIdentity("test", "Org1MSP"))

	client, err := resmgmt.New(fgosdkresmgmt.NewChannelClientProvider(createClientContext(clientCtx)).ChannelClient)
	assert.NotNil(t, client)
	assert.NoError(t, err)
}

func TestLifecycleInitEnclaveFailedWithInvalidRequest(t *testing.T) {
	fakeChannelClient := &fakes.ChannelClient{}
	fakeConverter := &fakes.CredentialConverter{}
	client := setupClient(fakeChannelClient, fakeConverter)

	var err error
	var request resmgmt.LifecycleInitEnclaveRequest

	// empty (no ChaincodeID)
	request = resmgmt.LifecycleInitEnclaveRequest{}
	_, err = client.LifecycleInitEnclave(channelID, request)
	assert.Error(t, err)

	// no EnclavePeerEndpoint
	request = resmgmt.LifecycleInitEnclaveRequest{ChaincodeID: chaincodeId}
	_, err = client.LifecycleInitEnclave(channelID, request)
	assert.Error(t, err)

	// no AttestationParams
	request = resmgmt.LifecycleInitEnclaveRequest{ChaincodeID: chaincodeId, EnclavePeerEndpoint: enclavePeerEndpoint}
	_, err = client.LifecycleInitEnclave(channelID, request)
	assert.Error(t, err)

	// invalid AttestationParams
	// TODO implement me once
	//request = LifecycleInitEnclaveRequest{ChaincodeID: chaincodeId, EnclavePeerEndpoint: enclavePeerEndpoint, AttestationParams: &sgx.AttestationParams{
	//	AttestationType: "InvalidType",
	//}}
	//_, err = client.LifecycleInitEnclave(channelID, request)
	//assert.Error(t, err)
}

func TestLifecycleInitEnclaveFailedToCreateChannelClient(t *testing.T) {
	expectedError := fmt.Errorf("ChannelClientError")
	getChannelClient := func(channelId string) (resmgmt.ChannelClient, error) {
		return nil, expectedError
	}

	client := &resmgmt.Client{getChannelClient, nil}

	initReq := resmgmt.LifecycleInitEnclaveRequest{
		ChaincodeID:         chaincodeId,
		EnclavePeerEndpoint: enclavePeerEndpoint, // define the peer where we wanna init our enclave
		AttestationParams: &sgx.AttestationParams{
			AttestationType: attestationType,
		},
	}

	_, err := client.LifecycleInitEnclave(channelID, initReq)
	assert.ErrorIs(t, err, expectedError)
}

func TestLifecycleInitEnclaveFailedToInitEnclave(t *testing.T) {
	expectedError := fmt.Errorf("someQueryError")
	fakeChannelClient := &fakes.ChannelClient{}
	fakeChannelClient.QueryReturns(nil, expectedError)
	fakeConverter := &fakes.CredentialConverter{}
	client := setupClient(fakeChannelClient, fakeConverter)

	initReq := resmgmt.LifecycleInitEnclaveRequest{
		ChaincodeID:         chaincodeId,
		EnclavePeerEndpoint: enclavePeerEndpoint, // define the peer where we wanna init our enclave
		AttestationParams: &sgx.AttestationParams{
			AttestationType: attestationType,
		},
	}

	_, err := client.LifecycleInitEnclave(channelID, initReq)
	assert.ErrorIs(t, err, expectedError)
}

func TestLifecycleInitEnclaveFailedToConvertCredentials(t *testing.T) {
	expectedError := fmt.Errorf("conversionError")

	fakeChannelClient := &fakes.ChannelClient{}
	fakeChannelClient.QueryReturns(nil, nil)
	fakeConverter := &fakes.CredentialConverter{}
	fakeConverter.ConvertCredentialsReturns("", expectedError)
	client := setupClient(fakeChannelClient, fakeConverter)

	initReq := resmgmt.LifecycleInitEnclaveRequest{
		ChaincodeID:         chaincodeId,
		EnclavePeerEndpoint: enclavePeerEndpoint,
		AttestationParams: &sgx.AttestationParams{
			AttestationType: attestationType,
		},
	}

	_, err := client.LifecycleInitEnclave(channelID, initReq)
	assert.ErrorIs(t, err, expectedError)
}

func TestLifecycleInitEnclaveFailedToRegisterEnclave(t *testing.T) {
	expectedError := fmt.Errorf("someRegisterError")
	fakeChannelClient := &fakes.ChannelClient{}
	fakeChannelClient.ExecuteReturns("", expectedError)
	fakeConverter := &fakes.CredentialConverter{}
	client := setupClient(fakeChannelClient, fakeConverter)

	initReq := resmgmt.LifecycleInitEnclaveRequest{
		ChaincodeID:         chaincodeId,
		EnclavePeerEndpoint: enclavePeerEndpoint,
		AttestationParams: &sgx.AttestationParams{
			AttestationType: attestationType,
		},
	}

	_, err := client.LifecycleInitEnclave(channelID, initReq)
	assert.ErrorIs(t, err, expectedError)
}

func TestLifecycleInitEnclaveSuccess(t *testing.T) {
	fakeChannelClient := &fakes.ChannelClient{}
	fakeChannelClient.QueryReturns(nil, nil)
	fakeChannelClient.ExecuteReturns(expectedTxID, nil)
	fakeConverter := &fakes.CredentialConverter{}

	client := setupClient(fakeChannelClient, fakeConverter)

	initReq := resmgmt.LifecycleInitEnclaveRequest{
		ChaincodeID:         chaincodeId,
		EnclavePeerEndpoint: enclavePeerEndpoint, // define the peer where we wanna init our enclave
		AttestationParams: &sgx.AttestationParams{
			AttestationType: attestationType,
		},
	}

	txId, err := client.LifecycleInitEnclave(channelID, initReq)
	assert.NoError(t, err)
	assert.Equal(t, expectedTxID, txId)

	assert.Equal(t, 1, fakeChannelClient.QueryCallCount())
	assert.Equal(t, 1, fakeChannelClient.ExecuteCallCount())

	chaincodeID, Fcn, Args, _ := fakeChannelClient.QueryArgsForCall(0)
	assert.Equal(t, chaincodeId, chaincodeID)
	assert.Equal(t, resmgmt.InitEnclaveCMD, Fcn)
	assert.Len(t, Args, 1)

	chaincodeID, Fcn, Args = fakeChannelClient.ExecuteArgsForCall(0)
	assert.Equal(t, resmgmt.ERCC, chaincodeID)
	assert.Equal(t, resmgmt.RegisterEnclaveCMD, Fcn)
	assert.Len(t, Args, 1)
}
