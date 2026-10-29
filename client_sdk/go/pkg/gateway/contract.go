/*
Copyright 2020 IBM All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

// Package gateway enables interaction with a FPC chaincode.
package gateway

import (
	"strings"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/core/contract"
	"github.com/pkg/errors"
)

// Contract provides functions to query/invoke FPC chaincodes based on the Gateway API.
//
// Contract is modeled after the Contract object of the Fabric Gateway client API (https://pkg.go.dev/github.com/hyperledger/fabric-gateway/pkg/client#Contract),
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

// Network interface that is needed by the FPC contract implementation.
// It is implemented by *client.Network of the Fabric Gateway client API.
type Network interface {
	GetContract(chaincodeName string) *client.Contract
}

type gatewayContract struct {
	c    *client.Contract
	sign identity.Sign
	dial PeerDialer
}

func (c *gatewayContract) Name() string {
	return c.c.ChaincodeName()
}

func (c *gatewayContract) EvaluateTransaction(name string, args ...string) ([]byte, error) {
	result, err := c.c.EvaluateTransaction(name, args...)
	return result, WithErrorDetails(err)
}

func (c *gatewayContract) SubmitTransaction(name string, args ...string) ([]byte, error) {
	result, err := c.c.SubmitTransaction(name, args...)
	return result, WithErrorDetails(err)
}

// CreateTransaction returns a transaction that is evaluated directly at the given enclave peers,
// as the Fabric Gateway routes requests by organization rather than by peer.
func (c *gatewayContract) CreateTransaction(name string, peerEndpoints ...string) (contract.Transaction, error) {
	if len(peerEndpoints) == 0 {
		return nil, errors.New("no enclave peer endpoints given")
	}
	return &peerTransaction{c: c, name: name, endpoints: peerEndpoints}, nil
}

// peerTransaction is evaluated at the first enclave peer that returns a successful response
type peerTransaction struct {
	c         *gatewayContract
	name      string
	endpoints []string
}

func (t *peerTransaction) Evaluate(args ...string) ([]byte, error) {
	var byteArgs [][]byte
	for _, arg := range args {
		byteArgs = append(byteArgs, []byte(arg))
	}

	var errs []string
	for _, endpoint := range t.endpoints {
		result, err := EvaluateAtPeer(t.c.c, t.c.sign, t.c.dial, endpoint, t.name, byteArgs...)
		if err == nil {
			return result, nil
		}
		errs = append(errs, err.Error())
	}
	return nil, errors.Errorf("failed to evaluate %s at enclave peers: %s", t.name, strings.Join(errs, "; "))
}

type contractProvider struct {
	network Network
	sign    identity.Sign
	dial    PeerDialer
}

func (cp *contractProvider) GetContract(id string) contract.Contract {
	return &gatewayContract{c: cp.network.GetContract(id), sign: cp.sign, dial: cp.dial}
}

// GetContract is the factory method for creating FPC Contract objects.
//
//	Parameters:
//	network is a Fabric Gateway network object, e.g., obtained via (*client.Gateway).GetNetwork
//	chaincodeID is the ID of the target chaincode
//	sign is the signing function of the gateway identity, used to sign the requests sent directly to the enclave peers
//	dial returns a connection to an enclave peer, given its endpoint as registered at the enclave registry
//
//	Returns:
//	The contract object
func GetContract(network Network, chaincodeID string, sign identity.Sign, dial PeerDialer) Contract {
	return contract.GetContract(&contractProvider{network: network, sign: sign, dial: dial}, chaincodeID)
}
