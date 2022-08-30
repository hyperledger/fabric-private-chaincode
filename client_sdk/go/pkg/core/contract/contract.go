/*
Copyright 2020 IBM All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

// Package contract implements the client-side FPC protocol
package contract

import (
	"strings"

	"github.com/hyperledger/fabric-private-chaincode/internal/crypto"
	"github.com/hyperledger/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric/common/flogging"
)

var logger = flogging.MustGetLogger("fpc-client-contract")

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

type Provider interface {
	GetContract(id string) Contract
}

// GetContract is the factory method for creating FPC Contract objects.
//  Parameters:
//  network is an initialized Fabric network object
//  chaincodeID is the ID of the target chaincode
//
//  Returns:
//  The contractImpl object
func GetContract(p Provider, chaincodeID string) *contractImpl {
	ercc := p.GetContract("ercc")
	return New(p.GetContract(chaincodeID), ercc, nil, &crypto.EncryptionProviderImpl{
		CSP: crypto.GetDefaultCSP(),
		GetCcEncryptionKey: func() ([]byte, error) {
			// Note that this function is called during EncryptionProvider.NewEncryptionContext()
			return ercc.EvaluateTransaction("queryChaincodeEncryptionKey", chaincodeID)
		}})
}

// contractImpl implements the client-side FPC protocol
type contractImpl struct {
	target        Contract
	ercc          Contract
	peerEndpoints []string
	ep            crypto.EncryptionProvider
}

func New(fpc Contract, ercc Contract, peerEndpoints []string, ep crypto.EncryptionProvider) *contractImpl {
	return &contractImpl{
		target:        fpc,
		ercc:          ercc,
		peerEndpoints: peerEndpoints,
		ep:            ep,
	}
}

func (c *contractImpl) Name() string {
	return c.target.Name()
}

func (c *contractImpl) EvaluateTransaction(name string, args ...string) ([]byte, error) {
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

	clearResponseBytes, err := ctx.Reveal(encryptedResponse)
	if err != nil {
		return nil, err
	}

	// unwrap Response.Payload
	return utils.UnwrapResponse(clearResponseBytes)
}

func (c *contractImpl) SubmitTransaction(name string, args ...string) ([]byte, error) {
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
	_, err = c.target.SubmitTransaction("__endorse", string(encryptedResponse))
	if err != nil {
		return nil, err
	}

	clearResponseBytes, err := ctx.Reveal(encryptedResponse)
	if err != nil {
		return nil, err
	}

	// unwrap Response.Payload
	return utils.UnwrapResponse(clearResponseBytes)
}

// getPeerEndpoints returns an array of peer endpoints that host the FPC chaincode enclave
// An endpoint is a simple string with the format `host:port`
func (c *contractImpl) getPeerEndpoints() ([]string, error) {
	if len(c.peerEndpoints) == 0 {
		resp, err := c.ercc.EvaluateTransaction("queryChaincodeEndPoints", c.Name())
		if err != nil {
			return nil, err
		}
		c.peerEndpoints = strings.Split(string(resp), ",")
	}
	return c.peerEndpoints, nil
}

func (c *contractImpl) evaluateTransaction(args ...string) ([]byte, error) {
	peers, err := c.getPeerEndpoints()
	if err != nil {
		return nil, err
	}

	txn, err := c.target.CreateTransaction(
		"__invoke",
		peers...,
	)
	if err != nil {
		return nil, err
	}

	logger.Debugf("calling __invoke!")
	return txn.Evaluate(args...)
}
