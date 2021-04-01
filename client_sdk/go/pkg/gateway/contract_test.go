/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"fmt"
	"testing"

	"github.com/hyperledger-labs/fabric-private-chaincode/client_sdk/go/pkg/gateway/fakes"
	"github.com/hyperledger-labs/fabric-private-chaincode/client_sdk/go/pkg/gateway/internal"
	"github.com/hyperledger-labs/fabric-private-chaincode/internal/crypto"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"github.com/stretchr/testify/assert"
)

//go:generate counterfeiter -o fakes/network.go -fake-name Network . network
type network interface {
	internal.Network
}

//go:generate counterfeiter -o fakes/contract.go -fake-name Contract . gatewayContract
type gatewayContract interface {
	internal.Contract
}

//go:generate counterfeiter -o fakes/transaction.go -fake-name Transaction . transaction
type transaction interface {
	internal.Transaction
}

//go:generate counterfeiter -o fakes/encryption_provider.go -fake-name EncryptionProvider . encryptionProvider
type encryptionProvider interface {
	crypto.EncryptionProvider
}

//go:generate counterfeiter -o fakes/encryption_context.go -fake-name EncryptionContext . encryptionContext
type encryptionContext interface {
	crypto.EncryptionContext
}

func TestNewContract(t *testing.T) {
	chaincodeID := "myChaincode"

	mockNetwork := &fakes.Network{}
	mockNetwork.GetContractReturns(&gateway.Contract{})

	// should try to get chaincode and ercc contracts
	contract := GetContract(mockNetwork, chaincodeID)
	assert.NotNil(t, contract)
	assert.Equal(t, chaincodeID, mockNetwork.GetContractArgsForCall(0))
	assert.Equal(t, "ercc", mockNetwork.GetContractArgsForCall(1))
}

func TestContractName(t *testing.T) {

	chaincodeID := "myChaincode"

	mockContract := &fakes.Contract{}
	mockContract.NameReturns(chaincodeID)

	contract := &contractState{
		contract: mockContract,
	}

	// should return chaincodeId
	name := contract.Name()
	assert.Equal(t, chaincodeID, name)
	assert.Equal(t, mockContract.NameCallCount(), 1)
}

func TestContractEvaluateTransactionSuccess(t *testing.T) {

	expectedResult := []byte("result")

	txn := &fakes.Transaction{}
	txn.EvaluateReturnsOnCall(0, expectedResult, nil)

	mockContract := &fakes.Contract{}
	mockContract.CreateTransactionReturnsOnCall(0, txn, nil)

	// ercc returns peers when getPeerEndpoints() is called
	mockERCC := &fakes.Contract{}
	mockERCC.EvaluateTransactionReturns([]byte("peer1,peer2,peer3"), nil)

	// mock encryption
	mockEncryptionContext := &fakes.EncryptionContext{}
	expectedEvalArgs := "someEncryptedArgs"
	mockEncryptionContext.ConcealCalls(func(f string, args []string) (string, error) {
		return expectedEvalArgs, nil
	})
	mockEncryptionContext.RevealCalls(func(input []byte) ([]byte, error) {
		return input, nil
	})

	mockEncryptionProvider := &fakes.EncryptionProvider{}
	mockEncryptionProvider.NewEncryptionContextReturns(mockEncryptionContext, nil)

	contract := &contractState{
		contract: mockContract,
		ercc:     mockERCC,
		ep:       mockEncryptionProvider,
	}

	// success
	resp, err := contract.EvaluateTransaction("someFunction", "arg1", "arg2")
	assert.Equal(t, expectedResult, resp)
	assert.NoError(t, err)

	// check that create transaction was called with "__invoke"
	name, f := mockContract.CreateTransactionArgsForCall(0)
	assert.Equal(t, "__invoke", name)
	assert.NotNil(t, f)

	// check that CreateTransaction was invoked only once
	assert.Equal(t, 1, mockContract.CreateTransactionCallCount())

	// check that the transaction was evaluates with correct args
	assert.Equal(t, 1, txn.EvaluateCallCount())
	assert.Len(t, txn.EvaluateArgsForCall(0), 1)
	assert.Equal(t, expectedEvalArgs, txn.EvaluateArgsForCall(0)[0])

}

func TestContractEvaluateAndSubmitTransactionFail(t *testing.T) {

	expectedResult := []byte("result")

	txn := &fakes.Transaction{}
	txn.EvaluateReturnsOnCall(0, expectedResult, nil)

	// see what happens if creation of encryption context returns error
	mockEncryptionProvider := &fakes.EncryptionProvider{}
	mockEncryptionProvider.NewEncryptionContextReturns(nil, fmt.Errorf("encryption Context Creation failed"))
	contract := &contractState{ep: mockEncryptionProvider}

	// failed
	resp, err := contract.EvaluateTransaction("someFunction", "arg1", "arg2")
	assert.Nil(t, resp)
	assert.Error(t, err)
	resp, err = contract.SubmitTransaction("someFunction", "arg1", "arg2")
	assert.Nil(t, resp)
	assert.Error(t, err)

	// see what happens if conceal returns an error
	mockEncryptionContext := &fakes.EncryptionContext{}
	mockEncryptionContext.ConcealCalls(func(f string, args []string) (string, error) {
		return "", fmt.Errorf("conceal failed")
	})

	mockEncryptionProvider.NewEncryptionContextReturns(mockEncryptionContext, nil)

	// failed
	resp, err = contract.EvaluateTransaction("someFunction", "arg1", "arg2")
	assert.Nil(t, resp)
	assert.Error(t, err)
	resp, err = contract.SubmitTransaction("someFunction", "arg1", "arg2")
	assert.Nil(t, resp)
	assert.Error(t, err)

	// see what happens if error while queryChaincodeEndPoints
	mockERCC := &fakes.Contract{}
	mockERCC.EvaluateTransactionReturns(nil, fmt.Errorf("ercc error"))
	mockContract := &fakes.Contract{}

	mockEncryptionContext.ConcealCalls(func(f string, args []string) (string, error) {
		return "", nil
	})

	contract = &contractState{
		contract: mockContract,
		ercc:     mockERCC,
		ep:       mockEncryptionProvider,
	}

	// failed
	resp, err = contract.EvaluateTransaction("someFunction", "arg1", "arg2")
	assert.Nil(t, resp)
	assert.Error(t, err)
	resp, err = contract.SubmitTransaction("someFunction", "arg1", "arg2")
	assert.Nil(t, resp)
	assert.Error(t, err)

	// see what happens if creating transaction fails
	mockContract.CreateTransactionReturns(nil, fmt.Errorf("error while creating transaction"))
	mockERCC.EvaluateTransactionReturns(nil, nil)

	// failed
	resp, err = contract.EvaluateTransaction("someFunction", "arg1", "arg2")
	assert.Nil(t, resp)
	assert.Error(t, err)
	resp, err = contract.SubmitTransaction("someFunction", "arg1", "arg2")
	assert.Nil(t, resp)
	assert.Error(t, err)

	// see what happens if __endorse fails
	txn.EvaluateReturnsOnCall(0, expectedResult, nil)
	mockContract.CreateTransactionReturns(txn, nil)
	mockContract.SubmitTransactionReturns(nil, fmt.Errorf("endorse failed"))

	// failed
	resp, err = contract.SubmitTransaction("someFunction", "arg1", "arg2")
	assert.Nil(t, resp)
	assert.Error(t, err)
}

func TestContractSubmitTransaction(t *testing.T) {
	expectedResult := []byte("result")

	invokeTx := &fakes.Transaction{}
	invokeTx.EvaluateReturnsOnCall(0, expectedResult, nil)

	mockContract := &fakes.Contract{}
	mockContract.CreateTransactionReturnsOnCall(0, invokeTx, nil)

	// ercc returns peers when getPeerEndpoints() is called
	mockERCC := &fakes.Contract{}
	mockERCC.EvaluateTransactionReturns([]byte("peer1,peer2,peer3"), nil)

	// mock encryption
	mockEncryptionContext := &fakes.EncryptionContext{}
	expectedEvalArgs := "someEncryptedArgs"
	mockEncryptionContext.ConcealCalls(func(f string, args []string) (string, error) {
		return expectedEvalArgs, nil
	})
	mockEncryptionContext.RevealCalls(func(input []byte) ([]byte, error) {
		return input, nil
	})

	mockEncryptionProvider := &fakes.EncryptionProvider{}
	mockEncryptionProvider.NewEncryptionContextReturns(mockEncryptionContext, nil)

	contract := &contractState{
		contract: mockContract,
		ercc:     mockERCC,
		ep:       mockEncryptionProvider,
	}

	// success
	resp, err := contract.SubmitTransaction("someFunction", "arg1", "arg2")
	assert.Equal(t, expectedResult, resp)
	assert.NoError(t, err)

	// check that create transaction was first called with "__invoke"
	name, f := mockContract.CreateTransactionArgsForCall(0)
	assert.Equal(t, "__invoke", name)
	assert.NotNil(t, f)

	// check that CreateTransaction was invoked once
	assert.Equal(t, 1, mockContract.CreateTransactionCallCount())

	// check that SubmitTransaction was invoked once
	assert.Equal(t, 1, mockContract.SubmitTransactionCallCount())
}

func TestContractRegisterEvent(t *testing.T) {
	// just check that it is correctly wired
	mockContract := &fakes.Contract{}
	contract := &contractState{contract: mockContract}
	contract.RegisterEvent("someEvent")
	assert.Equal(t, 1, mockContract.RegisterEventCallCount())
	assert.Equal(t, "someEvent", mockContract.RegisterEventArgsForCall(0))
}

func TestContractUnregister(t *testing.T) {
	// just check that it is correctly wired
	mockContract := &fakes.Contract{}
	contract := &contractState{contract: mockContract}

	type registration struct {
	}
	reg := &registration{}
	contract.Unregister(reg)
	assert.Equal(t, 1, mockContract.UnregisterCallCount())
	assert.Equal(t, reg, mockContract.UnregisterArgsForCall(0))

}
