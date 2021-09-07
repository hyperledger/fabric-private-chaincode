/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package lifecycle_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/core/lifecycle"
	"github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/core/lifecycle/fakes"
	"github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/sgx"
)

//go:generate counterfeiter -o fakes/channelclient.go -fake-name ChannelClient . chClient
//lint:ignore U1000 This is just used to generate fake
type chClient interface {
	lifecycle.ChannelClient
}

//go:generate counterfeiter -o fakes/credential_converter.go -fake-name CredentialConverter . credConverter
//lint:ignore U1000 This is just used to generate fake
type credConverter interface {
	lifecycle.CredentialConverter
}

const (
	channelID           = "mychannel"
	chaincodeId         = "my-fpc-chaincode"
	enclavePeerEndpoint = "mypeer.myorg.example.com"
	attestationType     = "simulation"
	expectedTxID        = "someTxID"
)

func setupClient(client lifecycle.ChannelClient, converter lifecycle.CredentialConverter) *lifecycle.Client {
	getChannelClient := func(channelId string) (lifecycle.ChannelClient, error) {
		return client, nil
	}

	return &lifecycle.Client{GetChannelClient: getChannelClient, Converter: converter}
}

func TestCreateNewClient(t *testing.T) {
	client, err := lifecycle.New(nil)
	assert.Nil(t, client)
	assert.Error(t, err, "invalid arguments, channel client loader is nil")
}

func TestLifecycleInitEnclaveFailedWithInvalidRequest(t *testing.T) {
	fakeChannelClient := &fakes.ChannelClient{}
	fakeConverter := &fakes.CredentialConverter{}
	client := setupClient(fakeChannelClient, fakeConverter)

	var err error
	var request lifecycle.LifecycleInitEnclaveRequest

	// empty (no ChaincodeID)
	request = lifecycle.LifecycleInitEnclaveRequest{}
	_, err = client.LifecycleInitEnclave(channelID, request)
	assert.Error(t, err)

	// no EnclavePeerEndpoint
	request = lifecycle.LifecycleInitEnclaveRequest{ChaincodeID: chaincodeId}
	_, err = client.LifecycleInitEnclave(channelID, request)
	assert.Error(t, err)

	// no AttestationParams
	request = lifecycle.LifecycleInitEnclaveRequest{ChaincodeID: chaincodeId, EnclavePeerEndpoint: enclavePeerEndpoint}
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
	getChannelClient := func(channelId string) (lifecycle.ChannelClient, error) {
		return nil, expectedError
	}

	client := &lifecycle.Client{GetChannelClient: getChannelClient}

	initReq := lifecycle.LifecycleInitEnclaveRequest{
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

	initReq := lifecycle.LifecycleInitEnclaveRequest{
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

	initReq := lifecycle.LifecycleInitEnclaveRequest{
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

	initReq := lifecycle.LifecycleInitEnclaveRequest{
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

	initReq := lifecycle.LifecycleInitEnclaveRequest{
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
	assert.Equal(t, lifecycle.InitEnclaveCMD, Fcn)
	assert.Len(t, Args, 1)

	chaincodeID, Fcn, Args = fakeChannelClient.ExecuteArgsForCall(0)
	assert.Equal(t, lifecycle.ERCC, chaincodeID)
	assert.Equal(t, lifecycle.RegisterEnclaveCMD, Fcn)
	assert.Len(t, Args, 1)
}
