/*
Copyright 2020 IBM All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

// Package gateway enables interaction with a FPC chaincode.
package gateway

import (
	"github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/core/contract"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

// Contract provides functions to query/invoke FPC chaincodes based on the Gateway API.
//
// Contract is modeled after the Contract object of the gateway package in the standard Fabric Go SDK (https://godoc.org/github.com/hyperledger/fabric-sdk-go/pkg/gateway#Contract),
// but in addition to the normal FPC operations, it performs FPC specific steps such as encryption/decryption of chaincode requests/responses.
//
// A Contract object is created using the GetContract() factory method.
// For an example of its use, see `contract_test.go`
type Contract interface {
	// Name returns the name of the smart contract
	Name() string

	// EvaluateTransaction will evaluate a transaction function and return its results.
	// The transaction function 'name'
	// will be evaluated on the endorsing peers but the responses will not be sent to
	// the ordering service and hence will not be committed to the ledger.
	// This can be used for querying the world state.
	//  Parameters:
	//  name is the name of the transaction function to be invoked in the smart contract.
	//  args are the arguments to be sent to the transaction function.
	//
	//  Returns:
	EvaluateTransaction(name string, args ...string) ([]byte, error)

	// SubmitTransaction will submit a transaction to the ledger. The transaction function 'name'
	// will be evaluated on the endorsing peers and then submitted to the ordering service
	// for committing to the ledger.
	//  Parameters:
	//  name is the name of the transaction function to be invoked in the smart contract.
	//  args are the arguments to be sent to the transaction function.
	//
	//  Returns:
	//  The return value of the transaction function in the smart contract.
	SubmitTransaction(name string, args ...string) ([]byte, error)
}

// Network interface that is needed by the FPC contract implementation
type Network interface {
	GetContract(chaincodeID string) *gateway.Contract
}

type gatewayContract struct {
	c *gateway.Contract
}

func (c *gatewayContract) Name() string {
	return c.c.Name()
}

func (c *gatewayContract) EvaluateTransaction(name string, args ...string) ([]byte, error) {
	return c.c.EvaluateTransaction(name, args...)
}

func (c *gatewayContract) SubmitTransaction(name string, args ...string) ([]byte, error) {
	return c.c.SubmitTransaction(name, args...)
}

func (c *gatewayContract) CreateTransaction(name string, peerEndpoints ...string) (contract.Transaction, error) {
	return c.c.CreateTransaction(name, gateway.WithEndorsingPeers(peerEndpoints...))
}

type contractProvider struct {
	network Network
}

func (cp *contractProvider) GetContract(id string) contract.Contract {
	return &gatewayContract{cp.network.GetContract(id)}
}

// GetContract is the factory method for creating FPC Contract objects.
//  Parameters:
//  network is an initialized Fabric network object
//  chaincodeID is the ID of the target chaincode
//
//  Returns:
//  The contract object
func GetContract(network Network, chaincodeID string) Contract {
	return contract.GetContract(&contractProvider{network: network}, chaincodeID)
}
