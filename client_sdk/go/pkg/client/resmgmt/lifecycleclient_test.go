/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resmgmt

import (
	"fmt"
	"testing"

	"github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/client/resmgmt/fakes"
	"github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/sgx"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/stretchr/testify/assert"
)

//go:generate counterfeiter -o fakes/channelclient.go -fake-name ChannelClient . channelClient
//lint:ignore U1000 This is just used to generate fake
type chClient interface {
	channelClient
}

//go:generate counterfeiter -o fakes/credential_converter.go -fake-name CredentialConverter . credentialConverter
//lint:ignore U1000 This is just used to generate fake
type credConverter interface {
	credentialConverter
}

const (
	channelID           = "mychannel"
	chaincodeId         = "my-fpc-chaincode"
	enclavePeerEndpoint = "mypeer.myorg.example.com"
	attestationType     = "simulation"
	expectedTxID        = fab.TransactionID("someTxID")
)

func setupClient(client channelClient, converter credentialConverter) *Client {
	getChannelClient := func(channelId string) (channelClient, error) {
		return client, nil
	}

	return &Client{nil, getChannelClient, converter}
}

func createClientContext(fabCtx context.Client) context.ClientProvider {
	return func() (context.Client, error) {
		return fabCtx, nil
	}
}

func TestCreateNewClient(t *testing.T) {
	clientCtx := fcmocks.NewMockContext(mspmocks.NewMockSigningIdentity("test", "Org1MSP"))
	client, err := New(createClientContext(clientCtx))
	assert.NotNil(t, client)
	assert.NoError(t, err)

	// MSP is missing
	invalidClientCtx := fcmocks.NewMockContext(mspmocks.NewMockSigningIdentity("test", ""))
	client, err = New(createClientContext(invalidClientCtx))
	assert.Nil(t, client)
	assert.Error(t, err)
}

func TestLifecycleInitEnclaveFailedWithInvalidRequest(t *testing.T) {
	fakeChannelClient := &fakes.ChannelClient{}
	fakeConverter := &fakes.CredentialConverter{}
	client := setupClient(fakeChannelClient, fakeConverter)

	var err error
	var request LifecycleInitEnclaveRequest

	// empty (no ChaincodeID)
	request = LifecycleInitEnclaveRequest{}
	_, err = client.LifecycleInitEnclave(channelID, request)
	assert.Error(t, err)

	// no EnclavePeerEndpoint
	request = LifecycleInitEnclaveRequest{ChaincodeID: chaincodeId}
	_, err = client.LifecycleInitEnclave(channelID, request)
	assert.Error(t, err)

	// no AttestationParams
	request = LifecycleInitEnclaveRequest{ChaincodeID: chaincodeId, EnclavePeerEndpoint: enclavePeerEndpoint}
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
	getChannelClient := func(channelId string) (channelClient, error) {
		return nil, expectedError
	}

	client := &Client{nil, getChannelClient, nil}

	initReq := LifecycleInitEnclaveRequest{
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
	fakeChannelClient.QueryReturns(channel.Response{}, expectedError)
	fakeConverter := &fakes.CredentialConverter{}
	client := setupClient(fakeChannelClient, fakeConverter)

	initReq := LifecycleInitEnclaveRequest{
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
	fakeChannelClient.QueryReturns(channel.Response{}, nil)
	fakeConverter := &fakes.CredentialConverter{}
	fakeConverter.ConvertCredentialsReturns("", expectedError)
	client := setupClient(fakeChannelClient, fakeConverter)

	initReq := LifecycleInitEnclaveRequest{
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
	fakeChannelClient.ExecuteReturns(channel.Response{}, expectedError)
	fakeConverter := &fakes.CredentialConverter{}
	client := setupClient(fakeChannelClient, fakeConverter)

	initReq := LifecycleInitEnclaveRequest{
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
	fakeChannelClient.QueryReturns(channel.Response{}, nil)
	fakeChannelClient.ExecuteReturns(channel.Response{TransactionID: expectedTxID}, nil)
	fakeConverter := &fakes.CredentialConverter{}

	client := setupClient(fakeChannelClient, fakeConverter)

	initReq := LifecycleInitEnclaveRequest{
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

	initResponse, _ := fakeChannelClient.QueryArgsForCall(0)
	assert.Equal(t, chaincodeId, initResponse.ChaincodeID)
	assert.Equal(t, initEnclaveCMD, initResponse.Fcn)
	assert.Len(t, initResponse.Args, 1)

	registerResponse, _ := fakeChannelClient.ExecuteArgsForCall(0)
	assert.Equal(t, ercc, registerResponse.ChaincodeID)
	assert.Equal(t, registerEnclaveCMD, registerResponse.Fcn)
	assert.Len(t, registerResponse.Args, 1)
}
