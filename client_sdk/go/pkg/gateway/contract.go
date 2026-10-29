/*
Copyright 2020 IBM All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

// Package gateway enables interaction with a FPC chaincode.
package gateway

import (
	"encoding/json"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/core/contract"
	"github.com/hyperledger/fabric-private-chaincode/internal/utils"
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
	c       *client.Contract
	network Network
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

// CreateTransaction returns a transaction that is evaluated at the given enclave peers.
// The Fabric Gateway routes requests by organization rather than by peer, hence the
// peer endpoints are mapped to the organizations that registered an enclave there.
func (c *gatewayContract) CreateTransaction(name string, peerEndpoints ...string) (contract.Transaction, error) {
	orgs, err := c.enclaveOrganizations(peerEndpoints)
	if err != nil {
		return nil, err
	}
	return &gatewayTransaction{c: c.c, name: name, orgs: orgs}, nil
}

func (c *gatewayContract) enclaveOrganizations(peerEndpoints []string) ([]string, error) {
	if len(peerEndpoints) == 0 {
		return nil, nil
	}

	resp, err := c.network.GetContract("ercc").EvaluateTransaction("queryListEnclaveCredentials", c.Name())
	if err != nil {
		return nil, errors.Wrap(WithErrorDetails(err), "failed to query enclave credentials")
	}

	var allCredentials []string
	if len(resp) > 0 {
		if err := json.Unmarshal(resp, &allCredentials); err != nil {
			return nil, errors.Wrap(err, "failed to parse enclave credentials")
		}
	}

	mspIdByEndpoint := make(map[string]string, len(allCredentials))
	for _, credentialsBase64 := range allCredentials {
		credentials, err := utils.UnmarshalCredentials(credentialsBase64)
		if err != nil {
			return nil, err
		}
		endpoint, err := utils.ExtractEndpoint(credentials)
		if err != nil {
			return nil, err
		}
		mspId, err := utils.ExtractPeerMspId(credentials)
		if err != nil {
			return nil, err
		}
		mspIdByEndpoint[endpoint] = mspId
	}

	var orgs []string
	seen := make(map[string]bool)
	for _, endpoint := range peerEndpoints {
		mspId, ok := mspIdByEndpoint[endpoint]
		if !ok {
			return nil, errors.Errorf("no enclave registered at peer %s", endpoint)
		}
		if !seen[mspId] {
			seen[mspId] = true
			orgs = append(orgs, mspId)
		}
	}
	return orgs, nil
}

type gatewayTransaction struct {
	c    *client.Contract
	name string
	orgs []string
}

func (t *gatewayTransaction) Evaluate(args ...string) ([]byte, error) {
	options := []client.ProposalOption{client.WithArguments(args...)}
	if len(t.orgs) > 0 {
		options = append(options, client.WithEndorsingOrganizations(t.orgs...))
	}
	result, err := t.c.Evaluate(t.name, options...)
	return result, WithErrorDetails(err)
}

type contractProvider struct {
	network Network
}

func (cp *contractProvider) GetContract(id string) contract.Contract {
	return &gatewayContract{c: cp.network.GetContract(id), network: cp.network}
}

// GetContract is the factory method for creating FPC Contract objects.
//
//	Parameters:
//	network is a Fabric Gateway network object, e.g., obtained via (*client.Gateway).GetNetwork
//	chaincodeID is the ID of the target chaincode
//
//	Returns:
//	The contract object
func GetContract(network Network, chaincodeID string) Contract {
	return contract.GetContract(&contractProvider{network: network}, chaincodeID)
}
