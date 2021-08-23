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

	"github.com/hyperledger/fabric-private-chaincode/internal/crypto"
)

var logger = flogging.MustGetLogger("fpc-client-gateway")

// Transaction interface that is needed by the FPC contract implementation
type Transaction interface {
	Evaluate(args ...string) ([]byte, error)
}

// Contract interface
type Contract interface {
	Name() string
	EvaluateTransaction(name string, args ...string) ([]byte, error)
	SubmitTransaction(name string, args ...string) ([]byte, error)
	CreateTransaction(name string, peerEndpoints ...string) (Transaction, error)
}

type ContractProvider interface {
	GetContract(id string) Contract
}

// GetContract is the factory method for creating FPC Contract objects.
//  Parameters:
//  network is an initialized Fabric network object
//  chaincodeID is the ID of the target chaincode
//
//  Returns:
//  The contract object
func GetContract(network ContractProvider, chaincodeID string) *ContractState {
	contract := network.GetContract(chaincodeID)
	ercc := network.GetContract("ercc")

	return &ContractState{
		Contract:      contract,
		ERCC:          ercc,
		peerEndpoints: nil,
		EP: &crypto.EncryptionProviderImpl{
			CSP: crypto.GetDefaultCSP(),
			GetCcEncryptionKey: func() ([]byte, error) {
				// Note that this function is called during EncryptionProvider.NewEncryptionContext()
				return ercc.EvaluateTransaction("queryChaincodeEncryptionKey", chaincodeID)
			}}}
}

type ContractState struct {
	// note that we wrap the target chaincode and ercc with an adapter that
	// implements the internal.Contract interface. This removes the direct
	// dependency to gateway.Contract struct as provided by the Fabric Go SDK,
	// and therefore allows better of this component.
	Contract      Contract
	ERCC          Contract
	peerEndpoints []string
	EP            crypto.EncryptionProvider
}

func (c *ContractState) Name() string {
	return c.Contract.Name()
}

func (c *ContractState) EvaluateTransaction(name string, args ...string) ([]byte, error) {
	ctx, err := c.EP.NewEncryptionContext()
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

func (c *ContractState) SubmitTransaction(name string, args ...string) ([]byte, error) {
	ctx, err := c.EP.NewEncryptionContext()
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
	_, err = c.Contract.SubmitTransaction("__endorse", string(encryptedResponse))
	if err != nil {
		return nil, err
	}

	return ctx.Reveal(encryptedResponse)
}

// getPeerEndpoints returns an array of peer endpoints that host the FPC chaincode enclave
// An endpoint is a simple string with the format `host:port`
func (c *ContractState) getPeerEndpoints() ([]string, error) {
	if len(c.peerEndpoints) == 0 {
		resp, err := c.ERCC.EvaluateTransaction("queryChaincodeEndPoints", c.Name())
		if err != nil {
			return nil, err
		}
		c.peerEndpoints = strings.Split(string(resp), ",")
	}
	return c.peerEndpoints, nil
}

func (c *ContractState) evaluateTransaction(args ...string) ([]byte, error) {
	peers, err := c.getPeerEndpoints()
	if err != nil {
		return nil, err
	}

	txn, err := c.Contract.CreateTransaction(
		"__invoke",
		peers...,
	)
	if err != nil {
		return nil, err
	}

	logger.Debugf("calling __invoke!")
	return txn.Evaluate(args...)
}
