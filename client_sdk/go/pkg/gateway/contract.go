/*
Copyright 2020 IBM All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

// Package gateway enables interaction with a FPC chaincode.
package gateway

import (
	"strings"

	"github.com/hyperledger/fabric/common/flogging"

	"github.com/hyperledger-labs/fabric-private-chaincode/internal/crypto"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

var logger = flogging.MustGetLogger("fpc-client-gateway")

// Contract provides functions to query/invoke FPC chaincodes based on the Gateway API.
//
// Contract is modeled after the Contract object of the gateway package in the standard Fabric Go SDK (https://godoc.org/github.com/hyperledger/fabric-sdk-go/pkg/gateway#Contract),
// but in addition to the normal FPC operations, it performs FPC specific steps such as encryption/decryption of chaincode requests/responses.
//
// A Contract object is created using the GetContract() factory method.
// For an example of its use, see https://github.com/hyperledger-labs/fabric-private-chaincode/blob/master/client_sdk/go/test/main.go
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

	// RegisterEvent registers for chaincode events. Unregister must be called when the registration is no longer needed.
	//  Parameters:
	//  eventFilter is the chaincode event filter (regular expression) for which events are to be received
	//
	//  Returns:
	//  the registration and a channel that is used to receive events. The channel is closed when Unregister is called.
	RegisterEvent(eventFilter string) (fab.Registration, <-chan *fab.CCEvent, error)

	// Unregister removes the given registration and closes the event channel.
	//  Parameters:
	//  registration is the registration handle that was returned from RegisterContractEvent method
	Unregister(registration fab.Registration)
}

// GetContract is the factory method for creating FPC Contract objects.
//  Parameters:
//  network is an initialized Fabric network object
//  chaincodeID is the ID of the target chaincode
//
//  Returns:
//  The contract object
func GetContract(network *gateway.Network, chaincodeID string) Contract {
	contract := network.GetContract(chaincodeID)
	ercc := network.GetContract("ercc")
	return &contractState{
		contract:      contract,
		ercc:          ercc,
		peerEndpoints: nil,
		ep: &crypto.EncryptionProviderImpl{GetCcEncryptionKey: func() ([]byte, error) {
			// Note that this function is called during EncryptionProvider.NewEncryptionContext()
			return ercc.EvaluateTransaction("queryChaincodeEncryptionKey", chaincodeID)
		}}}
}

type contractState struct {
	contract      *gateway.Contract
	ercc          *gateway.Contract
	peerEndpoints []string
	ep            crypto.EncryptionProvider
}

func (c *contractState) Name() string {
	return c.contract.Name()
}

// getPeerEndpoints returns an array of peer endpoints that host the FPC chaincode enclave
// An endpoint is a simple string with the format `host:port`
func (c *contractState) getPeerEndpoints() ([]string, error) {
	if len(c.peerEndpoints) == 0 {
		resp, err := c.ercc.EvaluateTransaction("queryChaincodeEndPoints", c.Name())
		if err != nil {
			return nil, err
		}
		c.peerEndpoints = strings.Split(string(resp), ",")
	}
	return c.peerEndpoints, nil
}

func (c *contractState) EvaluateTransaction(name string, args ...string) ([]byte, error) {
	ctx, err := c.ep.NewEncryptionContext()
	if err != nil {
		return nil, err
	}

	encryptedRequest, err := ctx.Conceal(name, args)
	if err != nil {
		return nil, err
	}

	// call __invoke
	encryptedResponse, err := c.evaluateTransaction(encryptedRequest)
	if err != nil {
		return nil, err
	}

	return ctx.Reveal(encryptedResponse)
}

func (c *contractState) evaluateTransaction(args ...string) ([]byte, error) {
	peers, err := c.getPeerEndpoints()
	if err != nil {
		return nil, err
	}

	// note that WithEndorsingPeers is only used with txn.Submit!!!
	// GO SDK needs to be patched! We should create a PR for that!
	txn, err := c.contract.CreateTransaction(
		"__invoke",
		gateway.WithEndorsingPeers(peers...),
	)
	if err != nil {
		return nil, err
	}

	logger.Debugf("calling __invoke!")
	return txn.Evaluate(args...)
}

func (c *contractState) SubmitTransaction(name string, args ...string) ([]byte, error) {
	ctx, err := c.ep.NewEncryptionContext()
	if err != nil {
		return nil, err
	}

	encryptedRequest, err := ctx.Conceal(name, args)
	if err != nil {
		return nil, err
	}

	// call __invoke
	encryptedResponse, err := c.evaluateTransaction(encryptedRequest)
	if err != nil {
		return nil, err
	}

	logger.Debugf("calling __endorse!")
	_, err = c.contract.SubmitTransaction("__endorse", string(encryptedResponse))
	if err != nil {
		return nil, err
	}

	return ctx.Reveal(encryptedResponse)
}

//func (c *Contract) CreateTransaction(name string, opts ...gateway.TransactionOption) (*gateway.Transaction, error) {
//	return c.CreateTransaction(name, opts...)
//}

func (c *contractState) RegisterEvent(eventFilter string) (fab.Registration, <-chan *fab.CCEvent, error) {
	return c.contract.RegisterEvent(eventFilter)
}

func (c *contractState) Unregister(registration fab.Registration) {
	c.contract.Unregister(registration)
}
