/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fpc

import (
	"encoding/base64"
	"encoding/json"
	"log"

	"github.com/hyperledger-labs/fabric-private-chaincode/client_sdk/go/fpc/crypto"
	"github.com/hyperledger-labs/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

type ContractInterface interface {
	Name() string
	EvaluateTransaction(name string, args ...string) ([]byte, error)
	SubmitTransaction(name string, args ...string) ([]byte, error)
	//CreateTransaction(name string, opts ...gateway.TransactionOption) (*gateway.Transaction, error)
	RegisterEvent(eventFilter string) (fab.Registration, <-chan *fab.CCEvent, error)
	Unregister(registration fab.Registration)
}

func GetContract(network *gateway.Network, chaincodeId string) ContractInterface {
	contract := network.GetContract(chaincodeId)
	ercc := network.GetContract("ercc")
	return &Contract{
		contract:     contract,
		ercc:         ercc,
		enclavePeers: nil,
		ep: &crypto.EncryptionProviderImpl{GetCcEncryptionKey: func() ([]byte, error) {
			return ercc.EvaluateTransaction("queryChaincodeEncryptionKey", chaincodeId)
		}}}
}

type Contract struct {
	contract     *gateway.Contract
	ercc         *gateway.Contract
	enclavePeers []string
	ep           crypto.EncryptionProvider
}

func (c *Contract) Name() string {
	return c.contract.Name()
}

func (c *Contract) getEnclavePeers() ([]string, error) {
	if len(c.enclavePeers) == 0 {
		resp, err := c.ercc.EvaluateTransaction("queryListEnclaveCredentials", c.Name())
		if err != nil {
			return nil, err
		}

		var credentialsList []string
		err = json.Unmarshal(resp, &credentialsList)
		if err != nil {
			return nil, err
		}

		for _, credentialsBase64 := range credentialsList {
			credentials, err := utils.UnmarshalCredentials(credentialsBase64)
			if err != nil {
				return nil, err
			}

			endpoint, err := utils.ExtractEndpoint(credentials)
			if err != nil {
				return nil, err
			}

			c.enclavePeers = append(c.enclavePeers, endpoint)
		}

	}
	return c.enclavePeers, nil
}

func (c *Contract) EvaluateTransaction(name string, args ...string) ([]byte, error) {
	ctx, err := c.ep.NewEncryptionContext()
	if err != nil {
		return nil, err
	}

	argsBase64, err := ctx.ChaincodeArgs(name, args)
	if err != nil {
		return nil, err
	}

	// call __invoke
	responseBytes, err := c.evaluateTransaction(argsBase64)

	return ctx.Response(responseBytes)
}

func (c *Contract) evaluateTransaction(args ...string) ([]byte, error) {
	peers, err := c.getEnclavePeers()
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

	log.Printf("calling __invoke!\n")
	return txn.Evaluate(args...)
}

func (c *Contract) SubmitTransaction(name string, args ...string) ([]byte, error) {
	ctx, err := c.ep.NewEncryptionContext()
	if err != nil {
		return nil, err
	}

	argsBase64, err := ctx.ChaincodeArgs(name, args)
	if err != nil {
		return nil, err
	}

	// call __invoke
	responseBytes, err := c.evaluateTransaction(argsBase64)

	log.Printf("calling __endorse!\n")
	_, err = c.contract.SubmitTransaction("__endorse", base64.StdEncoding.EncodeToString(responseBytes))
	if err != nil {
		return nil, err
	}

	return ctx.Response(responseBytes)
}

//func (c *Contract) CreateTransaction(name string, opts ...gateway.TransactionOption) (*gateway.Transaction, error) {
//	return c.CreateTransaction(name, opts...)
//}

func (c *Contract) RegisterEvent(eventFilter string) (fab.Registration, <-chan *fab.CCEvent, error) {
	return c.contract.RegisterEvent(eventFilter)
}

func (c *Contract) Unregister(registration fab.Registration) {
	c.contract.Unregister(registration)
}
